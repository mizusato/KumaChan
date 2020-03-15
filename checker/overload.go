package checker

import (
	"fmt"
)


func OverloadedCall (
	functions  [] *GenericFunction,
	name       string,
	type_args  [] Type,
	arg        SemiExpr,
	f_info     ExprInfo,
	call_info  ExprInfo,
	ctx        ExprContext,
) (SemiExpr, *ExprError) {
	if len(functions) == 0 { panic("something went wrong") }
	if len(functions) == 1 {
		var f = functions[0]
		var unbox_count uint
		var expr, err = GenericFunctionCall (
			f, name, 0, type_args, arg, f_info, call_info, ctx,
			&unbox_count,
		)
		if err != nil { return SemiExpr{}, err }
		return LiftTyped(expr), nil
	} else {
		var options = make([]AvailableCall, 0)
		var candidates = make([]string, 0)
		for i, f := range functions {
			var index = uint(i)
			var unbox_count uint
			var expr, err = GenericFunctionCall (
				f, name, index, type_args, arg, f_info, call_info, ctx,
				&unbox_count,
			)
			if err != nil {
				candidates = append(candidates, DescribeCandidate (
					name, f, ctx,
				))
			} else {
				options = append(options, AvailableCall {
					Expr:       expr,
					UnboxCount: unbox_count,
				})
			}
		}
		var available_count = len(options)
		if available_count == 0 {
			return SemiExpr{}, &ExprError {
				Point:    call_info.ErrorPoint,
				Concrete: E_NoneOfFunctionsCallable {
					Candidates: candidates,
				},
			}
		} else if available_count == 1 {
			return LiftTyped(options[0].Expr), nil
		} else {
			var min_unbox = ^(uint(0))
			var min_quantity = 0
			var min = -1
			for i, opt := range options {
				if opt.UnboxCount < min_unbox {
					min_unbox = opt.UnboxCount
					min_quantity = 0
					min = i
				} else if opt.UnboxCount == min_unbox {
					min_quantity += 1
				}
			}
			if min == -1 { panic("something went wrong") }
			if min_quantity == 1 {
				return LiftTyped(options[min].Expr), nil
			} else {
				return SemiExpr {
					Value: UndecidedCall {
						Options:  options,
						FuncName: name,
					},
					Info: call_info,
				}, nil
			}
		}
	}
}

func OverloadedAssignTo (
	expected   Type,
	functions  [] *GenericFunction,
	name       string,
	type_args  [] Type,
	info       ExprInfo,
	ctx        ExprContext,
) (Expr, *ExprError) {
	if len(functions) == 0 { panic("something went wrong") }
	if len(functions) == 1 {
		var f = functions[0]
		return GenericFunctionAssignTo (
			expected, name, 0, f, type_args, info, ctx,
		)
	} else {
		var candidates = make([]string, 0)
		for i, f := range functions {
			var index = uint(i)
			var expr, err = GenericFunctionAssignTo (
				expected, name, index, f, type_args, info, ctx,
			)
			if err != nil {
				candidates = append(candidates, DescribeCandidate (
					name, f, ctx,
				))
			} else {
				return expr, nil
			}
		}
		if expected == nil {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_ExplicitTypeRequired {},
			}
		} else {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_NoneOfFunctionsAssignable {
					To:         ctx.DescribeType(expected),
					Candidates: candidates,
				},
			}
		}
	}
}


func DescribeCandidate(name string, f *GenericFunction, ctx ExprContext) string {
	return fmt.Sprintf (
		"%s: %s", name, ctx.DescribeIncompleteType (
			AnonymousType { f.DeclaredType },
		),
	)
}
