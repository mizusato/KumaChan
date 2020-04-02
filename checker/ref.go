package checker

import (
	"kumachan/loader"
	"kumachan/transformer/ast"
)


func (impl UntypedRef) SemiExprVal() {}
type UntypedRef struct {
	RefBody   UntypedRefBody
	TypeArgs  [] Type
}

type UntypedRefBody interface { UntypedRefBody() }
func (impl UntypedRefToType) UntypedRefBody() {}
type UntypedRefToType struct {
	TypeName  loader.Symbol
	Type      *GenericType
}
func (impl UntypedRefToFunctions) UntypedRefBody() {}
type UntypedRefToFunctions struct {
	FuncName   string
	Functions  [] *GenericFunction
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


func CheckRef(ref ast.Ref, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(ref.Node)
	var maybe_symbol = ctx.ModuleInfo.Module.SymbolFromRef(ref)
	var symbol, ok = maybe_symbol.(loader.Symbol)
	if !ok { return SemiExpr{}, &ExprError {
		Point:    ctx.GetErrorPoint(ref.Module.Node),
		Concrete: E_ModuleNotFound { loader.Id2String(ref.Module) },
	} }
	var sym_concrete, exists = ctx.LookupSymbol(symbol)
	if !exists { return SemiExpr{}, &ExprError {
		Point:    ctx.GetErrorPoint(ref.Id.Node),
		Concrete: E_TypeOrValueNotFound { symbol },
	} }
	var type_ctx = ctx.GetTypeContext()
	var type_args = make([]Type, len(ref.TypeArgs))
	for i, arg_node := range ref.TypeArgs {
		var t, err = TypeFrom(arg_node.Type, type_ctx)
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
			Value: RefConstant { symbol },
			Info:  info,
		}), nil
	case SymTypeParam:
		return SemiExpr{}, &ExprError {
			Point:    ctx.GetErrorPoint(ref.Id.Node),
			Concrete: E_TypeParamInExpr { symbol.SymbolName },
		}
	case SymType:
		return SemiExpr {
			Value: UntypedRef {
				RefBody:  UntypedRefToType {
					TypeName: symbol,
					Type:     s.Type,
				},
				TypeArgs: type_args,
			},
			Info:  info,
		}, nil
	case SymFunctions:
		return SemiExpr {
			Value: UntypedRef {
				RefBody:  UntypedRefToFunctions {
					FuncName:  symbol.SymbolName,
					Functions: s.Functions,
				},
				TypeArgs: type_args,
			},
			Info:  info,
		}, nil
	default:
		panic("impossible branch")
	}
}


func AssignRefTo(expected Type, ref UntypedRef, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	switch r := ref.RefBody.(type) {
	case UntypedRefToType:
		return Expr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_TypeUsedAsValue { r.TypeName },
		}
	case UntypedRefToFunctions:
		var name = r.FuncName
		var functions = r.Functions
		var type_args = ref.TypeArgs
		return OverloadedAssignTo (
			expected, functions, name, type_args, info, ctx,
		)
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
		var type_args = ref.TypeArgs
		var expr, err = Box (
			arg, g, g_name, ref_info, type_args, call_info, ctx,
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
	default:
		panic("impossible branch")
	}
}
