package checker

import (
	"strings"
	. "kumachan/util/error"
	"kumachan/compiler/loader"
	"kumachan/compiler/loader/parser/ast"
	"kumachan/rpc/kmd"
)


type CheckedModule struct {
	Vendor     string
	Project    string
	Name       string
	RawModule  *loader.Module
	Imported   map[string] *CheckedModule
	Constants  map[string] CheckedConstant
	Functions  map[string] ([] CheckedFunction)
	Effects    [] CheckedEffect
	Context    CheckContext
}
type CheckedConstant struct {
	Point  ErrorPoint
	Value  ExprLike
}
type CheckedFunction struct {
	Point     ErrorPoint
	Body      ExprLike
	Implicit  [] string
	FunctionKmdInfo
}
type CheckedEffect struct {
	Point  ErrorPoint
	Value  Expr
}
type FunctionKmdInfo struct {
	IsAdapter    bool
	AdapterId    kmd.AdapterId
	IsValidator  bool
	ValidatorId  kmd.ValidatorId
}

type ExprLike interface { ExprLike() }
func (impl ExprNative) ExprLike() {}
type ExprNative struct {
	Name   string
	Point  ErrorPoint
}
func (impl ExprPredefinedValue) ExprLike() {}
type ExprPredefinedValue struct {
	Value  interface{}
}
func (impl ExprKmdApi) ExprLike() {}
type ExprKmdApi struct {
	Id  kmd.TransformerPartId
}
func (impl ExprExpr) ExprLike() {}
type ExprExpr Expr

type Index  map[string] *CheckedModule

type CheckContext struct {
	Types      TypeRegistry
	Functions  FunctionStore
	Constants  ConstantStore
	Mapping    KmdIdMapping
}

type ModuleInfo struct {
	Module     *loader.Module
	Types      TypeRegistry
	Constants  ConstantCollection
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
	Arguments    map[uint] Type  // mutable (interior)
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
type SymLocalValue struct { ValueType Type }
func (impl SymConst) Sym() {}
type SymConst struct { Const *Constant; Name loader.Symbol }
func (impl SymTypeParam) Sym() {}
type SymTypeParam struct { Index uint }
func (impl SymType) Sym() {}
type SymType struct { Type *GenericType; Name loader.Symbol; ForceExact bool }
func (impl SymFunctions) Sym() {}
type SymFunctions struct { Functions []*GenericFunction; Name string }
func (impl SymLocalAndFunc) Sym() {}
type SymLocalAndFunc struct { Local SymLocalValue; Func SymFunctions }


func CreateExprContext(mod_info ModuleInfo, params ([] TypeParam), bounds TypeBounds) ExprContext {
	return ExprContext {
		ModuleInfo:   mod_info,
		TypeParams:   params,
		TypeBounds:   bounds,
		LocalValues:  make(map[string]Type),
		Inferring:    TypeArgsInferringContext {
			Enabled:    false,
			Parameters: nil,
			Arguments:  nil,
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
			InferredTypes: ctx.Inferring.Arguments,
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

func (ctx ExprContext) LookupSymbol(raw loader.Symbol) (Sym, bool) {
	var lookup_type = func(sym loader.Symbol) (Sym, bool) {
		g, exists := ctx.ModuleInfo.Types[sym]
		if exists {
			return SymType { Type: g, Name: sym }, true
		}
		if len(sym.SymbolName) > len(ForceExactSuffix) &&
			strings.HasSuffix(sym.SymbolName, ForceExactSuffix) {
			var sym_name_force = strings.TrimSuffix(sym.SymbolName, ForceExactSuffix)
			var sym_force = loader.NewSymbol(sym.ModuleName, sym_name_force)
			g, exists := ctx.ModuleInfo.Types[sym_force]
			if exists {
				return SymType { Type: g, Name: sym_force, ForceExact: true }, true
			}
		}
		return nil, false
	}
	var lookup_functions = func(sym_name string) (SymFunctions, bool) {
		var real_sym_name = sym_name
		f_refs, exists := ctx.ModuleInfo.Functions[sym_name]
		if !exists &&
			len(sym_name) > len(FuncSuffix) &&
			strings.HasSuffix(sym_name, FuncSuffix) {
			var func_sym_name = strings.TrimSuffix(sym_name, FuncSuffix)
			f_refs, exists = ctx.ModuleInfo.Functions[func_sym_name]
			if exists {
				real_sym_name = func_sym_name
			}
		}
		if exists {
			var functions = make([] *GenericFunction, len(f_refs))
			for i, ref := range f_refs {
				functions[i] = ref.Function
			}
			return SymFunctions {
				Name:      real_sym_name,
				Functions: functions,
			}, true
		}
		return SymFunctions{}, false
	}
	var mod_name = raw.ModuleName
	var sym_name = raw.SymbolName
	if mod_name == "" {
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
		for index, param := range ctx.TypeParams {
			if param.Name == sym_name {
				return SymTypeParam { Index: uint(index) }, true
			}
		}
		functions, exists := lookup_functions(sym_name)
		if exists {
			return functions, true
		}
		var self = ctx.ModuleInfo.Module.Name
		var sym_this_mod = loader.NewSymbol(self, sym_name)
		g, exists := lookup_type(sym_this_mod)
		if exists {
			return g, true
		}
		constant, exists := ctx.ModuleInfo.Constants[sym_this_mod]
		if exists {
			return SymConst { Const: constant, Name: sym_this_mod }, true
		}
		return nil, false
	} else {
		g, exists := lookup_type(raw)
		if exists {
			return g, true
		}
		constant, exists := ctx.ModuleInfo.Constants[raw]
		if exists {
			return SymConst { Const: constant, Name: raw }, true
		}
		f_refs, exists := ctx.ModuleInfo.Functions[raw.SymbolName]
		if exists {
			var functions = make([] *GenericFunction, 0)
			for _, ref := range f_refs {
				functions = append(functions, ref.Function)
			}
			return SymFunctions {
				Name:      raw.SymbolName,
				Functions: functions,
			}, true
		}
		return nil, false
	}
}

func (ctx ExprContext) WithAddedLocalValues(added (map[string] Type)) (ExprContext, string) {
	var merged = make(map[string] Type)
	var shadowed = ""
	for name, t := range ctx.LocalValues {
		var _, exists = added[name]
		if exists {
			shadowed = name
		}
		merged[name] = t
	}
	for name, t := range added {
		merged[name] = t
	}
	var new_ctx ExprContext
	*(&new_ctx) = ctx
	new_ctx.LocalValues = merged
	return new_ctx, shadowed
}

func (ctx ExprContext) WithInferringEnabled(params ([] TypeParam)) ExprContext {
	var new_ctx ExprContext
	*(&new_ctx) = ctx
	new_ctx.Inferring = TypeArgsInferringContext {
		Enabled:    true,
		Parameters: params,
		Arguments:  make(map[uint] Type),
	}
	return new_ctx
}

func (ctx ExprContext) WithInferringStateCloned() ExprContext {
	if ctx.Inferring.Enabled {
		var new_ctx ExprContext
		*(&new_ctx) = ctx
		var cloned_args = make(map[uint] Type)
		for k, v := range new_ctx.Inferring.Arguments {
			cloned_args[k] = v
		}
		new_ctx.Inferring.Arguments = cloned_args
		return new_ctx
	} else {
		return ctx
	}
}

func (ctx ExprContext) WithInferringStateSaved() (struct{}, func()) {
	if ctx.Inferring.Enabled {
		var args = ctx.Inferring.Arguments
		var saved = make(map[uint] Type)
		for k, v := range args {
			saved[k] = v
		}
		return struct{}{}, func() {
			for k, _ := range args {
				delete(args, k)
			}
			for k, v := range saved {
				args[k] = v
			}
		}
	} else {
		return struct{}{}, func() {}
	}
}

func (ctx ExprContext) GetExprInfo(node ast.Node) ExprInfo {
	return ExprInfo { ErrorPoint: ErrorPointFrom(node) }
}


func Check(expr ast.Expr, ctx ExprContext) (SemiExpr, *ExprError) {
	return CheckCall(DesugarExpr(expr), ctx)
}

func CheckTerm(term ast.VariousTerm, ctx ExprContext) (SemiExpr, *ExprError) {
	switch t := term.Term.(type) {
	case ast.Call:
		return CheckCall(t, ctx)
	case ast.Cast:
		return CheckCast(t, ctx)
	case ast.Lambda:
		return CheckLambda(t, ctx)
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
	case ast.Bundle:
		return CheckBundle(t, ctx)
	case ast.Get:
		return CheckGet(t, ctx)
	case ast.Infix:
		return CheckInfix(t, ctx)
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
	*CheckedModule, Index, kmd.SchemaTable, [] E,
) {
	var types, type_nodes, err1 = RegisterTypes(entry, raw_index)
	if err1 != nil {
		var type_errors = make([] E, len(err1))
		for i, e := range err1 {
			type_errors[i] = e
		}
		return nil, nil, nil, type_errors
	}
	var constants = make(ConstantStore)
	var functions = make(FunctionStore)
	var _, err2 = CollectConstants(entry, types, constants)
	if err2 != nil { return nil, nil, nil, [] E { err2 } }
	var mapping, sch, inj, err3 = CollectKmdApi(types, type_nodes, raw_index)
	if err3 != nil { return nil, nil, nil, [] E { err3 } }
	var _, err4 = CollectFunctions(entry, types, inj, functions)
	if err4 != nil { return nil, nil, nil, [] E { err4 } }
	var ctx = CheckContext {
		Types:     types,
		Functions: functions,
		Constants: constants,
		Mapping:   mapping,
	}
	var checked_index = make(Index)
	var checked, errs = TypeCheckModule(entry, checked_index, ctx)
	if errs != nil { return nil, nil, nil, errs }
	return checked, checked_index, sch, nil
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
	var constants = ctx.Constants[mod_name]
	var mod_info = ModuleInfo {
		Module:    mod,
		Types:     ctx.Types,
		Constants: constants,
		Functions: functions,
	}
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
		var add = func(t Func, body ExprLike, imp ([] string), node ast.Node) {
			func_map[name] = append(func_map[name], CheckedFunction {
				Point:    ErrorPointFrom(node),
				Body:     body,
				Implicit: imp,
				FunctionKmdInfo: GetFunctionKmdInfo(name, t, ctx.Mapping),
			})
		}
		for _, f_ref := range group {
			if f_ref.IsImported {
				continue
			}
			var f = f_ref.Function
			switch body := f.Body.(type) {
			case ast.Lambda:
				var implicit_fields = make([] string, len(f.Implicit))
				var implicit_types = make(map[string] Type)
				for name, field := range f.Implicit {
					implicit_fields[field.Index] = name
					implicit_types[name] = field.Type
				}
				var blank_ctx = CreateExprContext(mod_info, f.TypeParams, f.TypeBounds)
				var f_expr_ctx, _ = blank_ctx.WithAddedLocalValues(implicit_types)
				var lambda, err1 = CheckLambda(body, f_expr_ctx)
				if err1 != nil {
					errors = append(errors, err1)
					continue
				}
				var t = &AnonymousType { f.DeclaredType }
				var body_expr, err2 = AssignTo(t, lambda, f_expr_ctx)
				if err2 != nil {
					errors = append(errors, err2)
					continue
				}
				add(f.DeclaredType, ExprExpr(body_expr),
					implicit_fields, f.Node)
			case ast.NativeRef:
				add(f.DeclaredType, ExprNative {
					Name:  string(body.Id.Value),
					Point: ErrorPointFrom(body.Node),
				}, make([] string, 0), f.Node)
			case ast.KmdApiFuncBody:
				add(f.DeclaredType, ExprKmdApi { Id: body.Id },
					make([] string, 0), f.Node)
			default:
				panic("impossible branch")
			}
		}
	}
	var expr_ctx = CreateExprContext(mod_info, __NoParams, __NoBounds)
	var const_map = make(map[string] CheckedConstant)
	for sym, constant := range constants {
		if sym.ModuleName != mod_name {
			continue
		}
		var name = sym.SymbolName
		switch val := constant.Value.(type) {
		case ast.Expr:
			var semi_expr, err1 = Check(val, expr_ctx)
			if err1 != nil {
				errors = append(errors, err1)
				continue
			}
			var t = constant.DeclaredType
			var expr, err2 = AssignTo(t, semi_expr, expr_ctx)
			if err2 != nil {
				errors = append(errors, err2)
				continue
			}
			const_map[name] = CheckedConstant {
				Point: ErrorPointFrom(constant.Node),
				Value: ExprExpr(expr),
			}
		case ast.NativeRef:
			const_map[name] = CheckedConstant {
				Point: ErrorPointFrom(constant.Node),
				Value: ExprNative {
					Name:  string(val.Id.Value),
					Point: ErrorPointFrom(val.Node),
				},
			}
		case ast.PredefinedValue:
			const_map[name] = CheckedConstant {
				Point: ErrorPointFrom(constant.Node),
				Value: ExprPredefinedValue {
					Value: val.Value,
				},
			}
		default:
			panic("impossible branch")
		}
	}
	var do_effects = make([] CheckedEffect, 0)
	for _, cmd := range mod.AST.Statements {
		switch do := cmd.Statement.(type) {
		case ast.Do:
			var semi_expr, err1 = Check(do.Effect, expr_ctx)
			if err1 != nil {
				errors = append(errors, err1)
				continue
			}
			var expr, err2 = AssignTo(__DoType, semi_expr, expr_ctx)
			if err2 != nil {
				errors = append(errors, err2)
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
			Constants: const_map,
			Functions: func_map,
			Effects:   do_effects,
			Context:   ctx,
		}
		index[mod_name] = checked
		return checked, nil
	}
}
