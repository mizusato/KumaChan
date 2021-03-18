package checker

import (
	"kumachan/lang/parser/ast"
)


type UntypedCall struct {
	Callee    SemiExpr
	Argument  ast.Call
	Context   ExprContext
}

func (impl Call) ExprVal() {}
type Call struct {
	Function  Expr
	Argument  Expr
}

func CheckCall(call ast.VariousCall, ctx ExprContext) (SemiExpr, *ExprError) {
	switch c := call.Call.(type) {
	case ast.CallPrefix:
		var callee, err1 = Check(c.Callee, ctx)
		if err1 != nil { return SemiExpr{}, err1 }
		var arg, err2 = Check(c.Argument, ctx)
		if err2 != nil { return SemiExpr{}, err2 }
		return CheckDesugaredCall(callee, arg, call.Node, ctx)
	case ast.CallInfix:
		var callee, err1 = Check(c.Operator, ctx)
		if err1 != nil { return SemiExpr{}, err1 }
		var left, err2 = Check(c.Left, ctx)
		if err2 != nil { return SemiExpr{}, err2 }
		var right, err3 = Check(c.Right, ctx)
		if err3 != nil { return SemiExpr{}, err3 }
		var arg = SemiExpr {
			Value: SemiTypedTuple {
				Values: [] SemiExpr { left, right },
			},
			Info:  ctx.GetExprInfo(call.Node),
		}
		return CheckDesugaredCall(callee, arg, call.Node, ctx)
	default:
		panic("impossible branch")
	}
}

func CheckDesugaredCall(callee SemiExpr, arg SemiExpr, node ast.Node, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(node)
	var f_info = callee.Info
	switch f := callee.Value.(type) {
	case TypedExpr:
		var expr, err = CallTyped(Expr(f), arg, info, ctx)
		if err != nil { return SemiExpr{}, err }
		return LiftTyped(expr), nil
	case UntypedLambda:
		var typed, err = CallUntypedLambda(arg, f, f_info, info, ctx)
		if err != nil { return SemiExpr{}, err }
		return LiftTyped(typed), nil
	case UntypedRef:
		return CallUntypedRef(arg, f, f_info, info, ctx)
	case SemiTypedSwitch,
		SemiTypedBlock:
		return SemiExpr{}, &ExprError {
			Point:    f_info.ErrorPoint,
			Concrete: E_ExplicitTypeRequired {},
		}
	default:
		return SemiExpr{}, &ExprError {
			Point:    f_info.ErrorPoint,
			Concrete: E_ExprNotCallable {},
		}
	}
}

func CallTyped(f Expr, arg SemiExpr, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var r, ok = UnboxFunc(f.Type, ctx).(Func)
	if ok {
		var arg_typed, err = AssignTo(r.Input, arg, ctx)
		if err != nil { return Expr{}, err }
		var typed = Expr {
			Type:  r.Output,
			Value: Call {
				Function: Expr(f),
				Argument: arg_typed,
			},
			Info:  info,
		}
		return typed, nil
	} else {
		return Expr{}, &ExprError {
			Point:    f.Info.ErrorPoint,
			Concrete: E_ExprTypeNotCallable {
				Type: ctx.DescribeCertainType(f.Type),
			},
		}
	}
}

func CraftAstCallExpr(f ast.VariousTerm, arg ast.VariousTerm, node ast.Node) ast.Expr {
	return ast.Expr {
		Node:     node,
		Term:     ast.VariousTerm {
			Node: node,
			Term: ast.VariousCall {
				Node: node,
				Call: ast.CallPrefix {
					Node:     node,
					Callee:   ast.WrapTermAsExpr(f),
					Argument: ast.WrapTermAsExpr(arg),
				},
			},
		},
		Pipeline: nil,
	}
}

