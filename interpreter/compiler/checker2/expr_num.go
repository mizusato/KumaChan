package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
	"kumachan/standalone/util"
)


func checkChar(C ast.CharLiteral) ExprChecker {
	return ExprChecker(func(expected typsys.Type, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var loc = C.Location
		var info = checked.ExprInfo { Location: loc }
		var value, ok = util.ParseRune(C.Value)
		if !(ok) {
			return nil, nil, source.MakeError(loc,
				E_InvalidChar { Content: string(C.Value) })
		}
		var expr = &checked.Expr {
			Type: coreChar(ctx.Types),
			Info: info,
			Expr: checked.IntegerLiteral { Value: value },
		}
		// TODO: extract common logic for nil expected type
		if expected == nil {
			return expr, nil, nil
		} else {
			var assign_ctx = ctx.AssignContext()
			var ok, s = typsys.Assign(expected, expr.Type, assign_ctx)
			if ok {
				return expr, s, nil
			} else {
				return nil, nil, source.MakeError(loc,
					E_CharAssignedToIncompatibleType {
						TypeName: typsys.DescribeType(expected, nil),
					})
			}
		}
	})
}

func checkFloat(F ast.FloatLiteral) ExprChecker {
	return ExprChecker(func(expected typsys.Type, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var loc = F.Location
		var info = checked.ExprInfo { Location: loc }
		var value, ok = util.ParseDouble(F.Value)
		if !(ok) {
			return nil, nil, source.MakeError(loc,
				E_FloatOverflowUnderflow {})
		}
		if !(util.IsNormalFloat(value)) {
			panic("invalid float literal got from parser")
		}
		var expr = &checked.Expr {
			Type: coreNormalFloat(ctx.Types),
			Info: info,
			Expr: checked.FloatLiteral { Value: value },
		}
		if expected == nil {
			return expr, nil, nil
		} else {
			var assign_ctx = ctx.AssignContext()
			var ok, s = typsys.Assign(expected, expr.Type, assign_ctx)
			if ok {
				return expr, s, nil
			} else {
				return nil, nil, source.MakeError(loc,
					E_FloatAssignedToIncompatibleType {
						TypeName: typsys.DescribeType(expected, nil),
					})
			}
		}
	})
}

func checkInteger(I ast.IntegerLiteral) ExprChecker {
	return ExprChecker(func(expected typsys.Type, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var loc = I.Location
		var info = checked.ExprInfo { Location: loc }
		var value, ok = util.WellBehavedParseInteger(I.Value)
		if !(ok) { panic("something went wrong") }
		var get_big_min_t = func()(typsys.Type) {
			if util.IsNonNegative(value) {
				return coreNumber(ctx.Types)
			} else {
				return coreInteger(ctx.Types)
			}
		}
		if expected == nil {
			var big_min_t = get_big_min_t()
			return &checked.Expr {
				Type: big_min_t,
				Info: info,
				Expr: checked.IntegerLiteral { Value: value },
			}, nil, nil
		} else {
			for _, float_t := range floatTypes {
				if typeEqual(expected, float_t, ctx.Types) {
					var float_value, ok = util.IntegerToDouble(value)
					if ok {
						return &checked.Expr {
							Type: expected,
							Info: info,
							Expr: checked.FloatLiteral { Value: float_value },
						}, nil, nil
					} else {
						return nil, nil, source.MakeError(loc,
							E_IntegerNotRepresentableByFloatType {})
					}
				}
			}
			for _, int_t := range integerTypes {
				if typeEqual(expected, int_t.which, ctx.Types) {
					var adapted, ok = int_t.adapt(value)
					if ok {
						return &checked.Expr {
							Type: expected,
							Info: info,
							Expr: checked.IntegerLiteral { Value: adapted },
						}, nil, nil
					} else {
						return nil, nil, source.MakeError(loc,
							E_IntegerOverflowUnderflow {
								TypeName: typsys.DescribeType(expected, nil),
							})
					}
				} else {
					// continue
				}
			}
			var assign_ctx = ctx.AssignContext()
			var big_min_t = get_big_min_t()
			var ok, s = typsys.Assign(expected, big_min_t, assign_ctx)
			if ok {
				return &checked.Expr {
					Type: big_min_t,
					Info: info,
					Expr: checked.IntegerLiteral { Value: value },
				}, s, nil
			} else {
				return nil, nil, source.MakeError(loc,
					E_IntegerAssignedToIncompatibleType {
						TypeName: ctx.DescribeType(expected, nil),
					})
			}
		}
	})
}


