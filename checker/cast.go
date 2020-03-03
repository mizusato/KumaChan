package checker

import "kumachan/transformer/node"


func CheckCast(cast node.Cast, ctx ExprContext) (Expr, *ExprError) {
	var type_ctx = ctx.GetTypeContext()
	var target, err1 = TypeFrom(cast.Target.Type, type_ctx)
	if err1 != nil { return Expr{}, &ExprError {
		Point:    ctx.GetErrorPoint(cast.Target.Node),
		Concrete: E_TypeErrorInExpr { err1 },
	} }
	var semi, err2 = Check(cast.Expr, ctx)
	if err2 != nil { return Expr{}, err2 }
	var typed, err3 = AssignTo(target, semi, ctx)
	if err3 != nil { return Expr{}, err3 }
	return typed, nil
}
