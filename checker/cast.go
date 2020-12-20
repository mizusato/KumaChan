package checker

import "kumachan/parser/ast"


func CheckCast(cast ast.Cast, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(cast.Node)
	var type_ctx = ctx.GetTypeContext()
	// TODO: (:super: value)
	var target, err1 = TypeFrom(cast.Target, type_ctx)
	if err1 != nil { return SemiExpr{}, &ExprError {
		Point:    err1.Point,
		Concrete: E_TypeErrorInExpr { err1 },
	} }
	var semi, err2 = Check(cast.Object, ctx)
	if err2 != nil { return SemiExpr{}, err2 }
	var typed, err3 = AssignTo(target, semi, ctx)
	if err3 != nil { return SemiExpr{}, err3 }
	if !(AreTypesEqualInSameCtx(typed.Type, target)) {
		panic("something went wrong")
	}
	return LiftTyped(Expr {
		Type:  typed.Type,
		Value: typed.Value,
		Info:  info,
	}), nil
}
