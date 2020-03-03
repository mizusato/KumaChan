package checker

import "kumachan/transformer/node"


func (impl UntypedLambda) SemiExprVal() {}
type UntypedLambda struct {
	Input   Pattern
	Output  SemiExpr
}

func (impl Lambda) ExprVal() {}
type Lambda struct {
	Input   Pattern
	Output  Expr
}


func CheckLambda(lambda node.Lambda, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ExprInfo { ErrorPoint: ctx.GetErrorPoint(lambda.Node) }
	var output_expr = node.Expr {
		Node:  lambda.Node,
		Pipes: []node.Pipe { lambda.Output },
	}
	var input = PatternFrom(lambda.Input, ctx)
	var output_semi, err = Check(output_expr, ctx)
	if err != nil { return SemiExpr{}, nil }
	return SemiExpr {
		Value: UntypedLambda {
			Input:  input,
			Output: output_semi,
		},
		Info: info,
	}, nil
}


func AssignLambdaTo(expected Type, lambda UntypedLambda, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var _, err = RequireExplicitType(expected, info)
	if err != nil { return Expr{}, err }
	switch E := expected.(type) {
	case AnonymousType:
		switch func_repr := E.Repr.(type) {
		case Func:
			var input = func_repr.Input
			var output = func_repr.Output
			var inner_ctx, err1 = ctx.WithPatternMatching (
				input, lambda.Input, false,
			)
			if err1 != nil { return Expr{}, err1 }
			var output_typed, err2 = AssignTo (
				output, lambda.Output, inner_ctx,
			)
			if err2 != nil { return Expr{}, err2 }
			return Expr {
				Type:  expected,
				Info:  info,
				Value: Lambda {
					Input:  lambda.Input,
					Output: output_typed,
				},
			}, nil
		}
	}
	return Expr{}, &ExprError {
		Point:    info.ErrorPoint,
		Concrete: E_LambdaAssignedToNonFuncType {
			NonFuncType: ctx.DescribeType(expected),
		},
	}
}