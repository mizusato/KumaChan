package checker

import (
	"kumachan/transformer/node"
)


func (impl UntypedLambda) SemiExprVal() {}
type UntypedLambda struct {
	Input   node.VariousPattern
	Output  node.Expr
}

func (impl Lambda) ExprVal() {}
type Lambda struct {
	Input   Pattern
	Output  Expr
}


func CheckLambda(lambda node.Lambda, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(lambda.Node)
	return SemiExpr {
		Value: UntypedLambda {
			Input:  lambda.Input,
			Output: node.Expr {
				Node:     lambda.Node,
				Call:     lambda.Output,
				Pipeline: nil,
			},
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
			var input_t = func_repr.Input
			var output_t = func_repr.Output
			//
			var pattern, err1 = PatternFrom(lambda.Input, input_t, ctx)
			if err1 != nil { return Expr{}, err1 }
			//
			var inner_ctx = ctx.WithShadowingPatternMatching(pattern)
			var output_semi, err2 = Check(lambda.Output, inner_ctx)
			if err2 != nil { return Expr{}, err2 }
			//
			var output_typed, err3 = AssignTo(output_t, output_semi, inner_ctx)
			if err3 != nil { return Expr{}, err3 }
			return Expr {
				Type:  expected,
				Info:  info,
				Value: Lambda {
					Input:  pattern,
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


func CallUntypedLambda (
	input        SemiExpr,
	lambda       UntypedLambda,
	lambda_info  ExprInfo,
	call_info    ExprInfo,
	ctx          ExprContext,
) (SemiExpr, *ExprError) {
	var input_typed, input_is_typed = input.Value.(TypedExpr)
	if !input_is_typed {
		return SemiExpr{}, &ExprError {
			Point:    lambda_info.ErrorPoint,
			Concrete: E_ExplicitTypeRequired {},
		}
	}
	var pattern, err1 = PatternFrom(lambda.Input, input_typed.Type, ctx)
	if err1 != nil { return SemiExpr{}, err1 }
	var inner_ctx = ctx.WithShadowingPatternMatching(pattern)
	var output, err2 = Check(lambda.Output, inner_ctx)
	if err2 != nil { return SemiExpr{}, err2 }
	var output_typed, output_is_typed = output.Value.(TypedExpr)
	if !output_is_typed {
		return SemiExpr{}, &ExprError {
			Point:    lambda_info.ErrorPoint,
			Concrete: E_ExplicitTypeRequired {},
		}
	}
	var lambda_typed = Expr {
		Type:  AnonymousType { Func {
			Input:  input_typed.Type,
			Output: output_typed.Type,
		} },
		Value: Lambda {
			Input:  pattern,
			Output: Expr(output_typed),
		},
		Info:  lambda_info,
	}
	return LiftTyped(Expr{
		Type:  output_typed.Type,
		Value: Call {
			Function: lambda_typed,
			Argument: Expr(input_typed),
		},
		Info:  call_info,  // this is a little ambiguous
	}), nil
}
