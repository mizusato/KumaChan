package generator

import (
	. "kumachan/standalone/util/error"
	. "kumachan/interpreter/runtime/def"
	"kumachan/interpreter/lang/textual/ast"
	ch "kumachan/interpreter/compiler/checker"
)


type ExtraDepLocator = func(Dependency) (uint, bool)

type IncrementalCompiler struct {
	typeInfo       *ch.ModuleInfo
	baseLocator    DepLocator
	addedData      [] DataValue
	addedClosures  [] FuncNode
	addedAmount    uint
	addedDepMap    map[Dependency] uint
}

func NewIncrementalCompiler(info *ch.ModuleInfo, base DepLocator) *IncrementalCompiler {
	return &IncrementalCompiler {
		typeInfo:      info,
		baseLocator:   base,
		addedData:     make([] DataValue, 0),
		addedClosures: make([] FuncNode, 0),
		addedAmount:   0,
		addedDepMap:   make(map[Dependency] uint),
	}
}

func (ctx *IncrementalCompiler) AddTempThunk (
	name  string,
	t     ch.Type,
	val   ast.Expr,
) (UserFunctionValue, ([] Value), error) {
	var mod_name = ctx.typeInfo.Module.Name
	var all_dep_values = make([] Value, 0)
	var closure_deps = make(map[*Function] ([] Dependency))
	var expr_ctx = ch.ExprContext {
		ModuleInfo: *(ctx.typeInfo),
	}
	semi, err := ch.Check(val, expr_ctx)
	if err != nil { return nil, nil, err }
	expr, err := ch.AssignTo(t, semi, expr_ctx)
	if err != nil { return nil, nil, err }
	var body = ch.BodyThunk {
		Value: expr,
	}
	thunk_f, refs, errs := CompileFunction (
		body,
		[] string {},
		mod_name,
		name,
		ErrorPointFrom(val.Node),
	)
	if errs != nil { return nil, nil, MergeErrors(errs) }
	var deps = RefsToDeps(refs, &ctx.addedData, &ctx.addedClosures)
	var f_node = FuncNode {
		Underlying:   thunk_f,
		Dependencies: deps,
	}
	var all_deps = make([] Dependency, 0)
	var collect_all_deps func([] Dependency)
	collect_all_deps = func(deps ([] Dependency)) {
		for _, dep := range deps {
			switch D := dep.(type) {
			case DepData:
				var d = ctx.addedData[D.Index]
				all_deps = append(all_deps, dep)
				all_dep_values = append(all_dep_values, d.ToValue())
			case DepClosure:
				var cl = ctx.addedClosures[D.Index]
				all_deps = append(all_deps, dep)
				all_dep_values = append(all_dep_values, &ValFun {
					Underlying: cl.Underlying,
				})
				closure_deps[cl.Underlying] = cl.Dependencies
				collect_all_deps(cl.Dependencies)
			default:
				continue
			}
		}
	}
	collect_all_deps(deps)
	for i, dep := range all_deps {
		ctx.addedDepMap[dep] = (ctx.addedAmount + uint(i))
	}
	// note: shadowing is ok
	var f_offset = ctx.addedAmount + uint(len(all_deps))
	ctx.addedDepMap[DepFunction {
		Module: mod_name,
		Name:   name,
		Index:  0,
	}] = f_offset
	ctx.addedAmount += (uint(len(all_deps)) + 1)
	ctx.typeInfo.Functions[name] = [] ch.FunctionReference {
		ch.FunctionReference {
			ModuleName: mod_name,
			Index:      0,
			IsImported: false,
			Function:   &ch.GenericFunction {
				Section:    "",
				Node:       val.Node,
				Doc:        "",
				Tags:       ch.FunctionTags {},
				Public:     true,
				TypeParams: [] ch.TypeParam {},
				TypeBounds: ch.TypeBounds {},
				Implicit:   nil,
				DeclaredType: ch.Func {
					Input:  &ch.AnonymousType { Repr: ch.Unit {} },
					Output: expr.Type,
				},
				Body: nil,  // unnecessary
				GenericFunctionInfo: ch.GenericFunctionInfo {
					RawImplicit: [] ch.Type {},
					AliasList:   [] string {},
					IsSelfAlias: false,
					IsFromConst: true,
				},
			},
		},
	}
	var extra_locator = func(dep Dependency) (uint, bool) {
		var offset, exists = ctx.addedDepMap[dep]
		return offset, exists
	}
	for _, v := range all_dep_values {
		var f, is_f = v.(UserFunctionValue)
		if is_f {
			var f_node = FuncNode {
				Underlying:   f.Underlying,
				Dependencies: closure_deps[f.Underlying],
			}
			RelocateCode(&f_node, ctx.baseLocator, extra_locator)
		}
	}
	RelocateCode(&f_node, ctx.baseLocator, extra_locator)
	return &ValFun { Underlying: thunk_f }, all_dep_values, nil
}

func (ctx *IncrementalCompiler) SetTempThunkAlias(name string, alias string) {
	var mod_name = ctx.typeInfo.Module.Name
	group, exists := ctx.typeInfo.Functions[name]
	if !(exists) { panic("something went wrong") }
	if !(len(group) == 1) { panic("something went wrong") }
	thunk := group[0]
	ctx.typeInfo.Functions[alias] = [] ch.FunctionReference { thunk }
	global_index, exists := ctx.addedDepMap[DepFunction {
		Module: mod_name,
		Name:   name,
		Index:  0,
	}]
	if !(exists) { panic("something went wrong") }
	ctx.addedDepMap[DepFunction {
		Module: mod_name,
		Name:   alias,
		Index:  0,
	}] = global_index
}

