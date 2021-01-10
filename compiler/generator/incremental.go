package generator

import (
	"kumachan/compiler/loader"
	"kumachan/compiler/loader/parser/ast"
	. "kumachan/util/error"
	ch "kumachan/compiler/checker"
	"kumachan/lang"
)


type ExtraDepLocator = func(Dependency) (uint, bool)

type IncrementalCompiler struct {
	typeInfo       *ch.ModuleInfo
	baseLocator    DepLocator
	addedData      [] lang.DataValue
	addedClosures  [] FuncNode
	addedAmount    uint
	addedDepMap    map[Dependency] uint
}

func NewIncrementalCompiler(info *ch.ModuleInfo, base DepLocator) *IncrementalCompiler {
	return &IncrementalCompiler {
		typeInfo:      info,
		baseLocator:   base,
		addedData:     make([] lang.DataValue, 0),
		addedClosures: make([] FuncNode, 0),
		addedAmount:   0,
		addedDepMap:   make(map[Dependency] uint),
	}
}

func (ctx *IncrementalCompiler) AddConstant(id DepConstant, val ast.Expr) (
	lang.FunctionValue, ([] lang.Value), error,
) {
	var all_dep_values = make([] lang.Value, 0)
	var closure_deps = make(map[*lang.Function] ([] Dependency))
	var sym = loader.NewSymbol(id.Module, id.Name)
	var expr_ctx = ch.ExprContext {
		ModuleInfo: *(ctx.typeInfo),
	}
	semi, err := ch.Check(val, expr_ctx)
	if err != nil { return nil, nil, err }
	expr, err := ch.AssignTo(nil, semi, expr_ctx)
	if err != nil { return nil, nil, err }
	const_f, refs, errs := CompileConstant (
		ch.ExprExpr(expr),
		ctx.typeInfo.Module.Name,
		id.Name,
		ErrorPointFrom(val.Node),
	)
	if errs != nil { return nil, nil, MergeErrors(errs) }
	var const_deps = RefsToDeps(refs, &ctx.addedData, &ctx.addedClosures)
	var const_f_node = FuncNode {
		Underlying:   const_f,
		Dependencies: const_deps,
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
				all_dep_values = append(all_dep_values, &lang.ValFunc {
					Underlying: cl.Underlying,
				})
				closure_deps[cl.Underlying] = cl.Dependencies
				collect_all_deps(cl.Dependencies)
			default:
				continue
			}
		}
	}
	collect_all_deps(const_deps)
	for i, dep := range all_deps {
		ctx.addedDepMap[dep] = (ctx.addedAmount + uint(i))
	}
	// note: shadowing is ok
	var const_offset = ctx.addedAmount + uint(len(all_deps))
	ctx.addedDepMap[id] = const_offset
	ctx.addedAmount += (uint(len(all_deps)) + 1)
	ctx.typeInfo.Constants[sym] = &ch.Constant {
		Node:         val.Node,
		Public:       true,
		DeclaredType: expr.Type,
		Value:        val,
	}
	var extraLocator = func(dep Dependency) (uint, bool) {
		var offset, exists = ctx.addedDepMap[dep]
		return offset, exists
	}
	for _, v := range all_dep_values {
		var f, is_f = v.(lang.FunctionValue)
		if is_f {
			var f_node = FuncNode {
				Underlying:   f.Underlying,
				Dependencies: closure_deps[f.Underlying],
			}
			RelocateCode(&f_node, ctx.baseLocator, extraLocator)
		}
	}
	RelocateCode(&const_f_node, ctx.baseLocator, extraLocator)
	return &lang.ValFunc { Underlying: const_f }, all_dep_values, nil
}

func (ctx *IncrementalCompiler) SetConstantAlias(id DepConstant, alias DepConstant) {
	var sym = loader.NewSymbol(id.Module, id.Name)
	var alias_sym = loader.NewSymbol(alias.Module, alias.Name)
	constant, exists := ctx.typeInfo.Constants[sym]
	if !(exists) { panic("something went wrong") }
	ctx.typeInfo.Constants[alias_sym] = constant
	index, exists := ctx.addedDepMap[id]
	if !(exists) { panic("something went wrong") }
	ctx.addedDepMap[alias] = index
}

