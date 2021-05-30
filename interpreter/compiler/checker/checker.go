package checker

// TODO: split this long file, refactor bad code in this file
import (
	"strings"
	"kumachan/standalone/rpc"
	"kumachan/standalone/rpc/kmd"
	"kumachan/interpreter/def"
	. "kumachan/standalone/util/error"
	"kumachan/interpreter/compiler/loader"
	"kumachan/interpreter/parser/ast"
	"kumachan/stdlib"
)


type CheckedModule struct {
	Vendor     string
	Project    string
	Name       string
	RawModule  *loader.Module
	Imported   map[string] *CheckedModule
	Functions  map[string] ([] CheckedFunction)
	Effects    [] CheckedEffect
	Context    CheckContext
	CheckedModuleInfo
}
type CheckedModuleInfo struct {
	ExportedTypes  map[string] map[def.Symbol] bool
}
type CheckedFunction struct {
	Point    ErrorPoint
	Body     Body
	Implicit [] string
	FunctionKmdInfo
	CheckedFunctionInfo
}
type CheckedFunctionInfo struct {
	Section      string
	Public       bool
	Doc          string
	Params       [] TypeParam
	Bounds       TypeBounds
	Type         Type
	RawImplicit  [] Type
	AliasList    [] string
	IsSelfAlias  bool
	IsFromConst  bool
	FunctionGeneratorFlags
}
type FunctionGeneratorFlags struct {
	Exported        bool
	ConsideredThunk bool
	KmdRelated      bool
}
type CheckedEffect struct {
	Point  ErrorPoint
	Value  Expr
}
type FunctionKmdInfo struct {
	IsAdapter   bool
	AdapterId   kmd.AdapterId
	IsValidator bool
	ValidatorId kmd.ValidatorId
	KmdIn       def.Symbol
	KmdOut      def.Symbol
}

type Body interface { CheckerBody() }
func (impl BodyLambda) CheckerBody() {}
type BodyLambda struct {
	Info    ExprInfo
	Lambda  Lambda
}
func (impl BodyThunk) CheckerBody() {}
type BodyThunk struct {
	Value  Expr
}
func (impl BodyNative) CheckerBody() {}
type BodyNative struct {
	Name   string
	Point  ErrorPoint
}
func (impl BodyGenerated) CheckerBody() {}
type BodyGenerated struct {
	Value  interface{}
}
func (impl BodyRuntimeGenerated) CheckerBody() {}
type BodyRuntimeGenerated struct {
	Value  interface{}
}

type Index  map[string] *CheckedModule

type CheckContext struct {
	Types      TypeRegistry
	Functions  FunctionStore
	Mapping    KmdIdMapping
}
func (ctx CheckContext) CollectExportedTypes() (map[string] map[def.Symbol] bool) {
	var collect_symbols func(Type, string, (map[def.Symbol] bool))
	collect_symbols = func(t Type, mod string, exported (map[def.Symbol] bool)) {
		switch T := t.(type) {
		case *NamedType:
			var sym = T.Name
			if (sym.ModuleName != mod) {
				return
			}
			if (exported[sym]) {
				return
			}
			exported[sym] = true
			var g = ctx.Types[sym]
			switch D := g.Definition.(type) {
			case *Boxed:
				collect_symbols(D.InnerType, mod, exported)
			case *Enum:
				for _, case_info := range D.CaseTypes {
					var case_t = &NamedType { Name: case_info.Name }
					collect_symbols(case_t, mod, exported)
				}
			}
		case *AnonymousType:
			switch R := T.Repr.(type) {
			case Tuple:
				for _, el := range R.Elements {
					collect_symbols(el, mod, exported)
				}
			case Record:
				for _, field := range R.Fields {
					collect_symbols(field.Type, mod, exported)
				}
			case Func:
				collect_symbols(R.Input, mod, exported)
				collect_symbols(R.Output, mod, exported)
			}
		}
	}
	var all_exported = make(map[string] map[def.Symbol] bool)
	for mod, functions := range ctx.Functions {
		var mod_exported = make(map[def.Symbol] bool)
		all_exported[mod] = mod_exported
		for _, group := range functions {
			for _, f := range group {
				if f.IsImported || !(f.Function.Public) {
					continue
				}
				var t = &AnonymousType { Repr: f.Function.DeclaredType }
				collect_symbols(t, mod, mod_exported)
			}
		}
	}
	return all_exported
}

type ModuleInfo struct {
	Module     *loader.Module
	Types      TypeRegistry
	Functions  FunctionCollection
}

type ExprContext struct {
	ModuleInfo     ModuleInfo
	TypeParams     [] TypeParam
	TypeBounds     TypeBounds
	LocalValues    map[string] Type
	Inferring      TypeArgsInferringContext  // contains mutable part
}

type TypeArgsInferringContext struct {
	Enabled      bool
	Parameters   [] TypeParam
	Bounds       TypeBounds
	Arguments    map[uint] ActiveType  // mutable (interior)
}
type ActiveType struct {
	CurrentValue  Type
	Constraint    ActiveTypeConstraint
}
type ActiveTypeConstraint int
const (
	AT_Exact ActiveTypeConstraint = iota
	AT_ExactOrBigger
	AT_ExactOrSmaller
)
func (ctx TypeArgsInferringContext) GetPlainArgs() (map[uint] Type) {
	var arg_types = make(map[uint] Type)
	for i, arg := range ctx.Arguments {
		arg_types[i] = arg.CurrentValue
	}
	return arg_types
}
func (ctx TypeArgsInferringContext) MergeArgsFrom(another TypeArgsInferringContext) {
	if ctx.Enabled && another.Enabled {
		if ctx.Arguments == nil || another.Arguments == nil {
			panic("something went wrong")
		}
		for k, v := range another.Arguments {
			ctx.Arguments[k] = v
		}
	}
}

type Expr struct {
	Type   Type
	Value  ExprVal
	Info   ExprInfo
}
type ExprInfo struct {
	ErrorPoint  ErrorPoint
}
type ExprVal interface { ExprVal() }

type SemiExpr struct {
	Value  SemiExprVal
	Info   ExprInfo
}
type SemiExprVal interface { SemiExprVal() }
func (impl TypedExpr) SemiExprVal() {}
type TypedExpr Expr
func LiftTyped(expr Expr) SemiExpr {
	return SemiExpr {
		Info:  expr.Info,
		Value: TypedExpr(expr),
	}
}

type Sym interface { Sym() }
func (impl SymLocalValue) Sym() {}
type SymLocalValue struct {
	ValueType  Type
}
func (impl SymTypeParam) Sym() {}
type SymTypeParam struct {
	Index  uint
}
func (impl SymType) Sym() {}
type SymType struct {
	Type       *GenericType
	Name       def.Symbol
	ForceExact bool
}
func (impl SymFunctions) Sym() {}
type SymFunctions struct {
	Functions   [] SymFunctionReference
	Name        string
	TypeExists  bool
	TypeSym     SymType
}
type SymFunctionReference struct {
	Function  *GenericFunction
	Index     uint
}
func (impl SymLocalAndFunc) Sym() {}
type SymLocalAndFunc struct {
	Local  SymLocalValue
	Func   SymFunctions
}


func CreateExprContext(mod_info ModuleInfo, params ([] TypeParam), bounds TypeBounds) ExprContext {
	return ExprContext {
		ModuleInfo:   mod_info,
		TypeParams:   params,
		TypeBounds:   bounds,
		LocalValues:  make(map[string]Type),
		Inferring:    TypeArgsInferringContext {
			Enabled:    false,
		},
	}
}

func (ctx ExprContext) GetTypeContext() TypeContext {
	return TypeContext {
		TypeBoundsContext: TypeBoundsContext {
			TypeValidationContext: TypeValidationContext {
				TypeConstructContext: TypeConstructContext {
					Module:     ctx.ModuleInfo.Module,
					Parameters: ctx.TypeParams,
				},
				Registry: ctx.ModuleInfo.Types,
			},
			Bounds: ctx.TypeBounds,
		},
	}
}

func (ctx ExprContext) DescribeCertainType(t Type) string {
	var params = TypeParamsNames(ctx.TypeParams)
	var mod = ctx.GetModuleName()
	return DescribeTypeWithParams(t, params, mod)
}

func (ctx ExprContext) DescribeInferredType(t Type) string {
	if ctx.Inferring.Enabled {
		return DescribeType(t, TypeDescContext {
			ParamNames:    TypeParamsNames(ctx.TypeParams),
			InferredNames: TypeParamsNames(ctx.Inferring.Parameters),
			InferredTypes: ctx.Inferring.GetPlainArgs(),
			CurrentModule: ctx.GetModuleName(),
		})
	} else {
		return ctx.DescribeCertainType(t)
	}
}

func (ctx ExprContext) GetModuleName() string {
	return ctx.ModuleInfo.Module.Name
}

func (ctx ExprContext) GetTypeRegistry() TypeRegistry {
	return ctx.ModuleInfo.Types
}

func (ctx ExprContext) LookupSymbol(raw def.Symbol) (Sym, bool) {
	// TODO: rename the Sym type and refactor this function
	var self = ctx.ModuleInfo.Module.Name
	var lookup_type func(def.Symbol) (SymType, bool)
	lookup_type = func(sym def.Symbol) (SymType, bool) {
		if sym.ModuleName == "" {
			var core_sym = def.MakeSymbol(stdlib.Mod_core, sym.SymbolName)
			var g, exists = lookup_type(core_sym)
			if exists {
				return g, true
			}
			var self_sym = def.MakeSymbol(self, sym.SymbolName)
			return lookup_type(self_sym)
		}
		var g, exists = ctx.ModuleInfo.Types[sym]
		if exists {
			return SymType { Type: g, Name: sym }, true
		}
		if len(sym.SymbolName) > len(ForceExactSuffix) &&
			strings.HasSuffix(sym.SymbolName, ForceExactSuffix) {
			var sym_name_force = strings.TrimSuffix(sym.SymbolName, ForceExactSuffix)
			var sym_force = def.MakeSymbol(sym.ModuleName, sym_name_force)
			var g, exists = ctx.ModuleInfo.Types[sym_force]
			if exists {
				return SymType { Type: g, Name: sym_force, ForceExact: true }, true
			}
		}
		return SymType{}, false
	}
	var lookup_functions = func(sym_name string) (SymFunctions, bool) {
		f_refs, exists := ctx.ModuleInfo.Functions[sym_name]
		if exists {
			var functions = make([] SymFunctionReference, len(f_refs))
			for i, ref := range f_refs {
				functions[i] = SymFunctionReference {
					Function: ref.Function,
					Index:    uint(i),
				}
			}
			var g, exists = lookup_type(raw)
			return SymFunctions {
				Name:       sym_name,
				Functions:  functions,
				TypeExists: exists,
				TypeSym:    g,
			}, true
		}
		return SymFunctions{}, false
	}
	var mod_name = raw.ModuleName
	var sym_name = raw.SymbolName
	if mod_name == "" {
		if sym_name == IgnoreMark {
			return nil, false
		}
		local, exists := ctx.LocalValues[sym_name]
		if exists {
			functions, exists := lookup_functions(sym_name)
			if exists {
				return SymLocalAndFunc {
					Local: SymLocalValue { ValueType: local },
					Func:  functions,
				}, true
			} else {
				return SymLocalValue { ValueType: local }, true
			}
		}
		functions, exists := lookup_functions(sym_name)
		if exists {
			return functions, true
		}
		for index, param := range ctx.TypeParams {
			if param.Name == sym_name {
				return SymTypeParam { Index: uint(index) }, true
			}
		}
		g, exists := lookup_type(raw)
		if exists {
			return g, true
		}
		return nil, false
	} else {
		f_refs, exists := ctx.ModuleInfo.Functions[sym_name]
		if exists {
			var functions = make([] SymFunctionReference, 0)
			for i, ref := range f_refs {
				if ref.ModuleName == mod_name {
					functions = append(functions, SymFunctionReference {
						Function: ref.Function,
						Index:    uint(i),
					})
				}
			}
			if len(functions) > 0 {
				var g, exists = lookup_type(raw)
				return SymFunctions {
					Name:       raw.SymbolName,
					Functions:  functions,
					TypeExists: exists,
					TypeSym:    g,
				}, true
			}
		}
		g, exists := lookup_type(raw)
		if exists {
			return g, true
		}
		return nil, false
	}
}

func (ctx ExprContext) WithAddedLocalValues(added (map[string] Type)) ExprContext {
	var merged = make(map[string] Type)
	for name, t := range ctx.LocalValues {
		merged[name] = t
	}
	for name, t := range added {
		merged[name] = t
	}
	var new_ctx ExprContext
	*(&new_ctx) = ctx
	new_ctx.LocalValues = merged
	return new_ctx
}

func (ctx ExprContext) WithInferringEnabled(params ([] TypeParam), bounds TypeBounds) ExprContext {
	var new_ctx ExprContext
	*(&new_ctx) = ctx
	var bounds_copy = TypeBounds {
		Sub:   make(map[uint] Type),
		Super: make(map[uint] Type),
	}
	for i, t := range bounds.Super {
		bounds_copy.Super[i] = MarkParamsAsBeingInferred(t)
	}
	for i, t := range bounds.Sub {
		bounds_copy.Sub[i] = MarkParamsAsBeingInferred(t)
	}
	new_ctx.Inferring = TypeArgsInferringContext {
		Enabled:    true,
		Parameters: params,
		Bounds:     bounds_copy,
		Arguments:  make(map[uint] ActiveType),
	}
	return new_ctx
}

func (ctx ExprContext) WithInferringStateCloned() ExprContext {
	if ctx.Inferring.Enabled {
		var new_ctx ExprContext
		*(&new_ctx) = ctx
		var cloned_args = make(map[uint] ActiveType)
		for k, v := range new_ctx.Inferring.Arguments {
			cloned_args[k] = v
		}
		new_ctx.Inferring.Arguments = cloned_args
		return new_ctx
	} else {
		return ctx
	}
}

func (ctx ExprContext) GetExprInfo(node ast.Node) ExprInfo {
	return ExprInfo { ErrorPoint: ErrorPointFrom(node) }
}

func Check(expr ast.Expr, ctx ExprContext) (SemiExpr, *ExprError) {
	var base, err = CheckTerm(expr.Term, ctx)
	if err != nil { return SemiExpr{}, err }
	return CheckPipeline(base, expr.Pipeline, ctx)
}

func CheckPipeline(base SemiExpr, pipes ([] ast.VariousPipe), ctx ExprContext) (SemiExpr, *ExprError) {
	var current = base
	var err *ExprError
	for _, pipe := range pipes {
		var node = pipe.Node
		var info = ctx.GetExprInfo(node)
		switch p := pipe.Pipe.(type) {
		case ast.PipeFunc:
			var callee, err = Check(p.Callee, ctx)
			if err != nil { return SemiExpr{}, err }
			var pipe_arg_, exists = p.Argument.(ast.Expr)
			if exists {
				var pipe_arg, err = Check(pipe_arg_, ctx)
				if err != nil { return SemiExpr{}, err }
				var arg = SemiExpr {
					Value: SemiTypedTuple {
						Values: [] SemiExpr { current, pipe_arg },
					},
					Info:  ctx.GetExprInfo(node),
				}
				current, err = CheckDesugaredCall(callee, arg, node, ctx)
				if err != nil { return SemiExpr{}, err }
			} else {
				current, err = CheckDesugaredCall(callee, current, node, ctx)
				if err != nil { return SemiExpr{}, err }
			}
		case ast.PipeGet:
			current, err = CheckGet(current, p.Key, info, ctx)
			if err != nil { return SemiExpr{}, err }
		case ast.PipeRefField:
			current, err = CheckRefField(current, p.Key, info, ctx)
			if err != nil { return SemiExpr{}, err }
		case ast.PipeCast:
			current, err = CheckCast(current, p.Target, info, ctx)
			if err != nil { return SemiExpr{}, err }
		case ast.PipeSwitch:
			current, err = CheckPipeSwitch(current, p.Type, info, ctx)
			if err != nil { return SemiExpr{}, err }
		case ast.PipeRefBranch:
			current, err = CheckRefBranch(current, p.Type, info, ctx)
			if err != nil { return SemiExpr{}, err }
		default:
			panic("impossible branch")
		}
	}
	return current, nil
}

func CheckTerm(term ast.VariousTerm, ctx ExprContext) (SemiExpr, *ExprError) {
	switch t := term.Term.(type) {
	case ast.VariousCall:
		return CheckCall(t, ctx)
	case ast.Lambda:
		return CheckLambda(t, ctx)
	case ast.PipelineLambda:
		return CheckPipelineLambda(t, ctx)
	case ast.Switch:
		return CheckSwitch(t, ctx)
	case ast.MultiSwitch:
		return CheckMultiSwitch(t, ctx)
	case ast.If:
		return CheckIf(t, ctx)
	case ast.Block:
		return CheckBlock(t, ctx)
	case ast.Cps:
		return CheckCps(t, ctx)
	case ast.Tuple:
		return CheckTuple(t, ctx)
	case ast.Record:
		return CheckRecord(t, ctx)
	case ast.InlineRef:
		return CheckRef(t, ctx)
	case ast.Array:
		return CheckArray(t, ctx)
	case ast.IntegerLiteral:
		return CheckInteger(t, ctx)
	case ast.FloatLiteral:
		return CheckFloat(t, ctx)
	case ast.StringLiteral:
		return CheckString(t, ctx)
	case ast.Formatter:
		return CheckFormatter(t, ctx)
	case ast.CharLiteral:
		return CheckChar(t, ctx)
	default:
		panic("impossible branch")
	}
}


func TypeCheck(entry *loader.Module, raw_index loader.Index) (
	*CheckedModule, Index, kmd.SchemaTable, rpc.ServiceIndex, [] E,
) {
	var types, type_nodes, err1 = RegisterTypes(entry, raw_index)
	if err1 != nil {
		var type_errors = make([] E, len(err1))
		for i, e := range err1 {
			type_errors[i] = e
		}
		return nil, nil, nil, nil, type_errors
	}
	var functions = make(FunctionStore)
	var mapping, sch, inj, err2 = CollectKmdApi(types, type_nodes, raw_index)
	if err2 != nil { return nil, nil, nil, nil, [] E { err2 } }
	var _, err3 = CollectFunctions(entry, types, inj, functions)
	if err3 != nil { return nil, nil, nil, nil, [] E { err3 } }
	var serv, err4 = CollectServices(raw_index, functions, types, sch, mapping)
	if err4 != nil { return nil, nil, nil, nil, [] E { err4 } }
	var ctx = CheckContext {
		Types:     types,
		Functions: functions,
		Mapping:   mapping,
	}
	var checked_index = make(Index)
	var checked, errs1 = TypeCheckModule(entry, checked_index, ctx)
	if errs1 != nil { return nil, nil, nil, nil, errs1 }
	var errs2 = EnforceGoodKmdFunctions(types, checked_index)
	if errs2 != nil { return nil, nil, nil, nil, errs2 }
	return checked, checked_index, sch, serv, nil
}

func TypeCheckModule(mod *loader.Module, index Index, ctx CheckContext) (
	*CheckedModule, [] E,
) {
	var mod_name = mod.Name
	var existing, exists = index[mod_name]
	if exists {
		return existing, nil
	}
	var functions = ctx.Functions[mod_name]
	var mod_info = ModuleInfo {
		Module:    mod,
		Types:     ctx.Types,
		Functions: functions,
	}
	// TODO: throw an error if there is a name conflict
	//       between constants and functions
	var errors = make([] E, 0)
	var imported = make(map[string] *CheckedModule)
	for alias, imported_item := range mod.ImpMap {
		var checked, errs = TypeCheckModule(imported_item, index, ctx)
		if errs != nil {
			errors = append(errors, errs...)
			continue
		}
		imported[alias] = checked
	}
	var func_map = make(map[string] ([] CheckedFunction))
	for name, group := range functions {
		func_map[name] = make([] CheckedFunction, 0)
		var add = func(f *GenericFunction, body Body) {
			var t = f.DeclaredType
			var implicit_fields = make([] string, len(f.Implicit))
			for name, field := range f.Implicit {
				implicit_fields[field.Index] = name
			}
			var is_unit_input = (func() bool {
				switch T := t.Input.(type) {
				case *AnonymousType:
					switch T.Repr.(type) {
					case Unit:
						return true
					}
				}
				return false
			})()
			var considered_thunk = is_unit_input && !(f.Tags.ExplicitCall)
			var kmd_info = GetFunctionKmdInfo(name, t, ctx.Mapping)
			var _, is_kmd_api = f.Body.(ast.KmdApiFuncBody)
			var is_kmd_user_func = kmd_info.IsAdapter || kmd_info.IsValidator
			var is_kmd_related = is_kmd_api || is_kmd_user_func
			func_map[name] = append(func_map[name], CheckedFunction {
				Point:    ErrorPointFrom(f.Node),
				Body:     body,
				Implicit: implicit_fields,
				FunctionKmdInfo:     kmd_info,
				CheckedFunctionInfo: CheckedFunctionInfo {
					Section:     f.Section,
					Public:      f.Public,
					Doc:         f.Doc,
					Params:      f.TypeParams,
					Bounds:      f.TypeBounds,
					Type:        &AnonymousType { t },
					RawImplicit: f.RawImplicit,
					AliasList:   f.AliasList,
					IsSelfAlias: f.IsSelfAlias,
					IsFromConst: f.IsFromConst,
					FunctionGeneratorFlags: FunctionGeneratorFlags {
						Exported:        f.Public,
						ConsideredThunk: considered_thunk,
						KmdRelated:      is_kmd_related,
					},
				},
			})
		}
		for _, f_ref := range group {
			if f_ref.IsImported {
				continue
			}
			var f = f_ref.Function
			switch body := f.Body.(type) {
			case ast.Lambda:
				var implicit_types = make(map[string] Type)
				for name, field := range f.Implicit {
					implicit_types[name] = field.Type
				}
				var blank_ctx = CreateExprContext(mod_info, f.TypeParams, f.TypeBounds)
				var f_expr_ctx = blank_ctx.WithAddedLocalValues(implicit_types)
				var lambda_semi, err1 = CheckLambda(body, f_expr_ctx)
				if err1 != nil {
					errors = append(errors, err1)
					continue
				}
				var t = &AnonymousType { f.DeclaredType }
				var lambda_expr, err2 = AssignTo(t, lambda_semi, f_expr_ctx)
				if err2 != nil {
					errors = append(errors, err2)
					continue
				}
				add(f, BodyLambda {
					Info:   lambda_expr.Info,
					Lambda: lambda_expr.Value.(Lambda),
				})
			case ast.NativeRef:
				add(f, BodyNative {
					Name:  string(body.Id.Value),
					Point: ErrorPointFrom(body.Node),
				})
			case ast.PredefinedThunk:
				var stored = body.Value
				switch stored.(type) {
				case def.UiObjectThunk:
					add(f, BodyRuntimeGenerated { Value: stored })
				default:
					var v = def.ValNativeFun(func(_ def.Value, _ def.InteropContext) def.Value {
						return stored
					})
					add(f, BodyGenerated { Value: v })
				}
			case ast.KmdApiFuncBody:
				var v = def.CreateKmdApiFunction(body.Id)
				add(f, BodyGenerated { Value: v })
			case ast.ServiceMethodFuncBody:
				var v = def.CreateServiceMethodCaller(name)
				add(f, BodyGenerated { Value: v })
			case ast.ServiceCreateFuncBody:
				var names = mod.ServiceMethodNames
				var v = def.ValNativeFun(func(arg def.Value, h def.InteropContext) def.Value {
					var prod = arg.(def.TupleValue)
					var data = prod.Elements[0]
					var ctx = prod.Elements[1].(def.TupleValue)
					var dtor = ctx.Elements[0]
					var methods = ctx.Elements[1:]
					return def.CreateServiceInstance(data, dtor, methods, names, h)
				})
				add(f, BodyGenerated { Value: v })
			default:
				panic("impossible branch")
			}
		}
	}
	var expr_ctx = CreateExprContext(mod_info, __NoParams, __NoBounds)
	var do_effects = make([] CheckedEffect, 0)
	for _, cmd := range mod.AST.Statements {
		switch do := cmd.Statement.(type) {
		case ast.Do:
			var semi_expr, err = Check(do.Effect, expr_ctx)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			// TODO: enhance code
			var expr Expr
			var ok = true
			for _, t := range __DoTypes {
				expr, err = AssignTo(t, semi_expr, expr_ctx)
				if err != nil {
					errors = append(errors, err)
					ok = false
				} else {
					break
				}
			}
			if !(ok) {
				continue
			}
			do_effects = append(do_effects, CheckedEffect {
				Point: ErrorPointFrom(do.Node),
				Value: expr,
			})
		}
	}
	if len(errors) > 0 {
		index[mod_name] = nil
		return nil, errors
	} else {
		var checked = &CheckedModule {
			Vendor:    mod.Vendor,
			Project:   mod.Project,
			Name:      mod_name,
			RawModule: mod,
			Imported:  imported,
			Functions: func_map,
			Effects:   do_effects,
			Context:   ctx,
			CheckedModuleInfo: CheckedModuleInfo {
				ExportedTypes: ctx.CollectExportedTypes(),
			},
		}
		index[mod_name] = checked
		return checked, nil
	}
}
