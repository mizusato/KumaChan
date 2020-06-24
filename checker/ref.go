package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/parser/ast"
)


func (impl UntypedRef) SemiExprVal() {}
type UntypedRef struct {
	RefBody   UntypedRefBody
	TypeArgs  [] Type
}

type UntypedRefBody interface { UntypedRefBody() }
func (impl UntypedRefToType) UntypedRefBody() {}
type UntypedRefToType struct {
	TypeName    loader.Symbol
	Type        *GenericType
	ForceExact  bool
}
func (impl UntypedRefToFunctions) UntypedRefBody() {}
type UntypedRefToFunctions struct {
	FuncName   string
	Functions  [] *GenericFunction
}
func (impl UntypedRefToMacro) UntypedRefBody() {}
type UntypedRefToMacro struct {
	MacroName  string
	Macro      *Macro
}

func (impl RefConstant) ExprVal() {}
type RefConstant struct {
	Name  loader.Symbol
}

func (impl RefFunction) ExprVal() {}
type RefFunction struct {
	Name    string
	Index   uint
	AbsRef  AbsRefFunction
}
type AbsRefFunction struct {
	Module  string
	Name    string
	Index   uint
}
func MakeRefFunction(name string, index uint, ctx ExprContext) RefFunction {
	var raw_ref = ctx.ModuleInfo.Functions[name][index]
	var abs_ref = AbsRefFunction {
		Module: raw_ref.ModuleName,
		Name:   name,
		Index:  raw_ref.Index,
	}
	return RefFunction {
		Name:   name,
		Index:  index,
		AbsRef: abs_ref,
	}
}

func (impl RefLocal) ExprVal() {}
type RefLocal struct {
	Name  string
}


func CheckRef(ref ast.InlineRef, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(ref.Node)
	var maybe_symbol = ctx.ModuleInfo.Module.SymbolFromInlineRef(ref)
	var symbol, ok = maybe_symbol.(loader.Symbol)
	if !ok { return SemiExpr{}, &ExprError {
		Point:    ErrorPointFrom(ref.Module.Node),
		Concrete: E_ModuleNotFound { loader.Id2String(ref.Module) },
	} }
	if len(ctx.MacroPath) > 0 && symbol.ModuleName == "" {
		var arg_name = symbol.SymbolName
		var arg, unwound_ctx, exists = ctx.FindMacroArg(arg_name)
		if exists {
			var semi, err = Check(arg, unwound_ctx)
			if err != nil { return SemiExpr{}, err }
			switch content := semi.Value.(type) {
			case TypedExpr:
				return LiftTyped(Expr {
					Type:  content.Type,
					Value: content.Value,
					Info:  info,
				}), nil
			default:
				return SemiExpr {
					Value: semi.Value,
					Info:  info,
				}, nil
			}
		}
	}
	var sym_concrete, exists = ctx.LookupSymbol(symbol)
	if !exists { return SemiExpr{}, &ExprError {
		Point:    ErrorPointFrom(ref.Id.Node),
		Concrete: E_TypeOrValueNotFound { symbol },
	} }
	var type_ctx = ctx.GetTypeContext()
	var type_args = make([]Type, len(ref.TypeArgs))
	for i, arg_node := range ref.TypeArgs {
		var t, _, err = TypeFrom(arg_node.Type, type_ctx)
		if err != nil { return SemiExpr{}, &ExprError {
			Point:    err.Point,
			Concrete: E_TypeErrorInExpr { err },
		} }
		type_args[i] = t
	}
	switch s := sym_concrete.(type) {
	case SymLocalValue:
		return LiftTyped(Expr {
			Type:  s.ValueType,
			Value: RefLocal { symbol.SymbolName },
			Info:  info,
		}), nil
	case SymConst:
		return LiftTyped(Expr {
			Type:  s.Const.DeclaredType,
			Value: RefConstant { s.Name },
			Info:  info,
		}), nil
	case SymTypeParam:
		return SemiExpr{}, &ExprError {
			Point:    ErrorPointFrom(ref.Id.Node),
			Concrete: E_TypeParamInExpr { symbol.SymbolName },
		}
	case SymType:
		return SemiExpr {
			Value: UntypedRef {
				RefBody:  UntypedRefToType {
					TypeName:   s.Name,
					Type:       s.Type,
					ForceExact: s.ForceExact,
				},
				TypeArgs: type_args,
			},
			Info:  info,
		}, nil
	case SymFunctions:
		return SemiExpr {
			Value: UntypedRef {
				RefBody:  UntypedRefToFunctions {
					FuncName:  s.Name,
					Functions: s.Functions,
				},
				TypeArgs: type_args,
			},
			Info:  info,
		}, nil
	case SymMacro:
		if len(type_args) > 0 {
			return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(ref.Node),
				Concrete: E_TypeParamsOnMacro {},
			}
		} else {
			return SemiExpr {
				Value: UntypedRef {
					RefBody: UntypedRefToMacro {
						MacroName: s.Name,
						Macro:     s.Macro,
					},
					TypeArgs: nil,
				},
				Info: info,
			}, nil
		}
	default:
		panic("impossible branch")
	}
}


func AssignRefTo(expected Type, ref UntypedRef, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	switch r := ref.RefBody.(type) {
	case UntypedRefToType:
		var g = r.Type
		var g_name = r.TypeName
		var force_exact = r.ForceExact
		var type_args = ref.TypeArgs
		var unit = LiftTyped(Expr {
			Type:  AnonymousType { Unit {} },
			Value: UnitValue {},
			Info:  info,
		})
		var boxed_unit, err = Box (
			unit, g, g_name, info, type_args,
			force_exact, info, ctx,
		)
		if err != nil { return Expr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_TypeUsedAsValue { r.TypeName },
		} }
		return TypedAssignTo(expected, boxed_unit, ctx)
	case UntypedRefToFunctions:
		var name = r.FuncName
		var functions = r.Functions
		var type_args = ref.TypeArgs
		return OverloadedAssignTo (
			expected, functions, name, type_args, info, ctx,
		)
	case UntypedRefToMacro:
		return Expr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_MacroUsedAsValue {},
		}
	default:
		panic("impossible branch")
	}
}


func CallUntypedRef (
	arg        SemiExpr,
	ref        UntypedRef,
	ref_info   ExprInfo,
	call_info  ExprInfo,
	ctx        ExprContext,
) (SemiExpr, *ExprError) {
	switch ref_body := ref.RefBody.(type) {
	case UntypedRefToType:
		var g = ref_body.Type
		var g_name = ref_body.TypeName
		var force_exact = ref_body.ForceExact
		var type_args = ref.TypeArgs
		var expr, err = Box (
			arg, g, g_name, ref_info, type_args,
			force_exact, call_info, ctx,
		)
		if err != nil { return SemiExpr{}, err }
		return LiftTyped(expr), nil
	case UntypedRefToFunctions:
		var functions = ref_body.Functions
		var name = ref_body.FuncName
		var type_args = ref.TypeArgs
		return OverloadedCall (
			functions, name, type_args, arg, ref_info, call_info, ctx,
		)
	case UntypedRefToMacro:
		panic("this branch should have been processed in CheckCall()")
	default:
		panic("impossible branch")
	}
}
