package checker

import "kumachan/parser/ast"


func (impl Panic) ExprVal() {}
type Panic struct {
	Message  Expr
}

func CheckPanic(p ast.Panic, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(p.Node)
	msg_semi, err := Check(p.Object, ctx)
	if err != nil { return SemiExpr{}, err }
	msg, err := AssignTo(__T_String, msg_semi, ctx)
	if err != nil { return SemiExpr{}, err }
	return LiftTyped(Expr {
		Type:  &WildcardRhsType {},
		Value: Panic { Message: msg },
		Info:  info,
	}), nil
}
