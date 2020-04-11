package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/transformer/ast"
)


type CheckedModule struct {
	Name       string
	RawModule  *loader.Module
	Imported   map[string] *CheckedModule
	Constants  map[string] CheckedConstant
	Functions  map[string] ([] CheckedFunction)
	Effects    [] CheckedEffect
}
type CheckedConstant struct {
	Point  ErrorPoint
	Value  ExprLike
}
type CheckedFunction struct {
	Point  ErrorPoint
	Body   ExprLike
}
type CheckedEffect struct {
	Point  ErrorPoint
	Value  Expr
}

type ExprLike interface { ExprLike() }
func (impl ExprNative) ExprLike() {}
type ExprNative struct {
	Name   string
	Point  ErrorPoint
}
func (impl ExprExpr) ExprLike() {}
type ExprExpr Expr

type Index  map[string] *CheckedModule

type CheckContext struct {
	Types      TypeRegistry
	Functions  FunctionStore
	Constants  ConstantStore
	Macros     MacroStore
}

type ModuleInfo struct {
	Module     *loader.Module
	Types      TypeRegistry
	Constants  ConstantCollection
	Functions  FunctionCollection
	Macros     MacroCollection
}

type ExprContext struct {
	ModuleInfo     ModuleInfo
	TypeParams     [] string
	LocalValues    map[string] Type
	InferTypeArgs  bool
	InferredNames  [] string
	Inferred       map[uint] Type  // mutable
	MacroPath      [] MacroExpanding
}

type MacroExpanding struct {
	Name   string
	Macro  *Macro
	Point  ErrorPoint
	Args   map[string] ast.Expr
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
type SymConst struct { Const *Constant }
func (impl SymTypeParam) Sym() {}
type SymTypeParam struct { Index uint }
func (impl SymType) Sym() {}
type SymType struct { Type *GenericType }
func (impl SymFunctions) Sym() {}
type SymFunctions struct { Functions []*GenericFunction }
func (impl SymMacro) Sym() {}
type SymMacro struct { Macro *Macro }


func CreateExprContext(mod_info ModuleInfo, params []string) ExprContext {
	return ExprContext {
		ModuleInfo:    mod_info,
		TypeParams:    params,
		LocalValues:   make(map[string]Type),
		InferTypeArgs: false,
	}
}

func (ctx ExprContext) GetTypeContext() TypeContext {
	return TypeContext {
		Module: ctx.ModuleInfo.Module,
		Params: ctx.TypeParams,
		Ireg:   ctx.ModuleInfo.Types,
	}
}

func (ctx ExprContext) DescribeType(t Type) string {
	return DescribeTypeWithParams(t, ctx.TypeParams)
}

func (ctx ExprContext) DescribeExpectedType(t Type) string {
	if ctx.InferTypeArgs {
		return DescribeType(t, TypeDescContext {
			ParamNames:    ctx.TypeParams,
			UseInferred:   ctx.InferTypeArgs,
			InferredNames: ctx.InferredNames,
			InferredTypes: ctx.Inferred,
		})
	} else {
		return ctx.DescribeType(t)
	}
}

func (ctx ExprContext) GetModuleName() string {
	return loader.Id2String(ctx.ModuleInfo.Module.Node.Name)
}

func (ctx ExprContext) LookupSymbol(raw loader.Symbol) (Sym, bool) {
	var mod_name = raw.ModuleName
	var sym_name = raw.SymbolName
	if mod_name == "" {
		local, exists := ctx.LocalValues[sym_name]
		if exists {
			return SymLocalValue { ValueType: local }, true
		}
		for index, param_name := range ctx.TypeParams {
			if param_name == sym_name {
				return SymTypeParam { Index: uint(index) }, true
			}
		}
		f_refs, exists := ctx.ModuleInfo.Functions[sym_name]
		if exists {
			var functions = make([]*GenericFunction, len(f_refs))
			for i, ref := range f_refs {
				functions[i] = ref.Function
			}
			return SymFunctions { Functions: functions }, true
		}
		m_ref, exists := ctx.ModuleInfo.Macros[sym_name]
		if exists {
			return SymMacro { m_ref.Macro }, true
		}
		var self = ctx.ModuleInfo.Module.Name
		var sym_self = loader.NewSymbol(self, sym_name)
		g, exists := ctx.ModuleInfo.Types[sym_self]
		if exists {
			return SymType { Type: g }, true
		}
		constant, exists := ctx.ModuleInfo.Constants[sym_self]
		if exists {
			return SymConst { Const: constant }, true
		}
		return nil, false
	} else {
		var g, exists = ctx.ModuleInfo.Types[raw]
		if exists {
			return SymType { Type: g }, true
		}
		constant, exists := ctx.ModuleInfo.Constants[raw]
		if exists {
			return SymConst { Const: constant }, true
		}
		return nil, false
	}
}

func (ctx ExprContext) WithAddedLocalValues(added map[string]Type) (ExprContext, string) {
	var merged = make(map[string]Type)
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

func (ctx ExprContext) WithTypeArgsInferringEnabled(names []string) ExprContext {
	var new_ctx ExprContext
	*(&new_ctx) = ctx
	new_ctx.InferTypeArgs = true
	new_ctx.InferredNames = names
	new_ctx.Inferred = make(map[uint] Type)
	return new_ctx
}

func (ctx ExprContext) GetExprInfo(node ast.Node) ExprInfo {
	return ExprInfo { ErrorPoint: ErrorPointFrom(node) }
}


func Check(expr ast.Expr, ctx ExprContext) (SemiExpr, *ExprError) {
	return CheckCall(DesugarExpr(expr), ctx)
}

func CheckTerm(term ast.VariousTerm, ctx ExprContext) (SemiExpr, *ExprError) {
	switch t := term.Term.(type) {
	case ast.Cast:
		return CheckCast(t, ctx)
	case ast.Lambda:
		return CheckLambda(t, ctx)
	case ast.Switch:
		return CheckSwitch(t, ctx)
	case ast.If:
		return CheckIf(t, ctx)
	case ast.Block:
		return CheckBlock(t, ctx)
	case ast.Tuple:
		return CheckTuple(t, ctx)
	case ast.Bundle:
		return CheckBundle(t, ctx)
	case ast.Get:
		return CheckGet(t, ctx)
	case ast.Array:
		return CheckArray(t, ctx)
	case ast.Text:
		return CheckText(t, ctx)
	case ast.VariousLiteral:
		switch l := t.Literal.(type) {
		case ast.IntegerLiteral:
			return CheckInteger(l, ctx)
		case ast.FloatLiteral:
			return CheckFloat(l, ctx)
		case ast.StringLiteral:
			return CheckString(l, ctx)
		case ast.CharLiteral:
			return CheckChar(l, ctx)
		default:
			panic("impossible branch")
		}
	case ast.Ref:
		return CheckRef(t, ctx)
	case ast.Infix:
		return CheckInfix(t, ctx)
	default:
		panic("impossible branch")
	}
}


func TypeCheck(entry *loader.Module, raw_index loader.Index) (
	*CheckedModule, Index, []E,
) {
	var types, err1 = RegisterTypes(entry, raw_index)
	if err1 != nil { return nil, nil, []E { err1 } }
	var constants = make(ConstantStore)
	var _, err2 = CollectConstants(entry, types, constants)
	if err2 != nil { return nil, nil, []E { err2 } }
	var functions = make(FunctionStore)
	var _, err3 = CollectFunctions(entry, types, functions)
	if err3 != nil { return nil, nil, []E { err3 } }
	var macros = make(MacroStore)
	var _, err4 = CollectMacros(entry, macros)
	if err4 != nil { return nil, nil, []E { err4 } }
	for mod_name, _ := range raw_index {
		for name, f := range functions[mod_name] {
			var macro, exists = macros[mod_name][name]
			if exists { return nil, nil, []E { &FunctionError {
				Point:    ErrorPointFrom(f[0].Function.Node),
				Concrete: E_FunctionConflictWithMacro {
					Name:   name,
					Module: macro.ModuleName,
				},
			} } }
		}
	}
	var ctx = CheckContext {
		Types:     types,
		Functions: functions,
		Constants: constants,
		Macros:    macros,
	}
	var checked_index = make(Index)
	var checked, errs = TypeCheckModule(entry, checked_index, ctx)
	if errs != nil { return nil, nil, errs }
	return checked, checked_index, nil
}

func TypeCheckModule(mod *loader.Module, index Index, ctx CheckContext) (
	*CheckedModule, []E,
) {
	var mod_name = mod.Name
	var existing, exists = index[mod_name]
	if exists {
		return existing, nil
	}
	var functions = ctx.Functions[mod_name]
	var constants = ctx.Constants[mod_name]
	var macros = ctx.Macros[mod_name]
	var mod_info = ModuleInfo {
		Module:    mod,
		Types:     ctx.Types,
		Constants: constants,
		Functions: functions,
		Macros:    macros,
	}
	var errors = make([]E, 0)
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
		func_map[name] = make([]CheckedFunction, 0)
		var add = func(body ExprLike, node ast.Node) {
			func_map[name] = append(func_map[name], CheckedFunction {
				Point: ErrorPointFrom(node),
				Body:  body,
			})
		}
		for _, f_ref := range group {
			if f_ref.IsImported {
				continue
			}
			var f = f_ref.Function
			switch body := f.Body.(type) {
			case ast.Lambda:
				var f_expr_ctx = CreateExprContext(mod_info, f.TypeParams)
				var lambda, err1 = CheckLambda(body, f_expr_ctx)
				if err1 != nil {
					errors = append(errors, err1)
					continue
				}
				var t = AnonymousType { f.DeclaredType }
				var body_expr, err2 = AssignTo(t, lambda, f_expr_ctx)
				if err2 != nil {
					errors = append(errors, err2)
					continue
				}
				add(ExprExpr(body_expr), f.Node)
			case ast.NativeRef:
				add(ExprNative {
					Name:  string(body.Id.Value),
					Point: ErrorPointFrom(body.Node),
				}, f.Node)
			default:
				panic("impossible branch")
			}
		}
	}
	var expr_ctx = CreateExprContext(mod_info, make([]string, 0))
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
		default:
			panic("impossible branch")
		}
	}
	var do_effects = make([] CheckedEffect, 0)
	for _, cmd := range mod.Node.Commands {
		switch do := cmd.Command.(type) {
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
			Name:      mod_name,
			RawModule: mod,
			Imported:  imported,
			Constants: const_map,
			Functions: func_map,
			Effects:   do_effects,
		}
		index[mod_name] = checked
		return checked, nil
	}
}
