package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
	"kumachan/interpreter/lang/common/source"
)


func getLambda(t typsys.Type) (typsys.Lambda, bool) {
	var nested, is_nested = t.(*typsys.NestedType)
	if !(is_nested) { return typsys.Lambda {}, false }
	var lambda, is_tuple = nested.Content.(typsys.Lambda)
	return lambda, is_tuple
}
func unboxLambda(t typsys.Type, mod string) (typsys.Lambda, bool) {
	var lambda, is_tuple = getLambda(t)
	if is_tuple {
		return lambda, true
	} else {
		var inner, _, exists = typsys.Unbox(t, mod)
		if exists {
			return unboxLambda(inner, mod)
		} else {
			return typsys.Lambda {}, false
		}
	}
}

func checkLambda(lambda ast.Lambda) ExprChecker {
	return ExprChecker(func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var lm = make(localBindingMap)
		var cc = makeCheckContext(lambda.Location, &s, ctx, lm)
		if expected == nil {
			return cc.error(E_ExplicitTypeRequired {})
		} else {
			var io, ok = getLambda(expected)
			if !(ok) {
				return cc.error(E_LambdaAssignedToIncompatible {
					TypeName: cc.describeType(expected),
				})
			}
			var in, err1 = cc.productPatternMatch(lambda.Input, io.Input)
			if err1 != nil { return cc.propagate(err1) }
			var out, err2 = cc.checkChildExpr(io.Output, lambda.Output)
			if err2 != nil { return cc.propagate(err2) }
			var lambda_t = &typsys.NestedType { Content: io }
			return cc.ok(lambda_t, checked.Lambda { In: in, Out: out })
		}
	})
}


