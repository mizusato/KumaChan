package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
)


func checkLambda(L ast.Lambda) ExprChecker {
	return makeExprCheckerWithLocalScope(L.Location, func(cc *checkContextWithLocalScope) checkResult {
		if cc.expected == nil {
			return cc.error(E_ExplicitTypeRequired {})
		} else {
			var io, ok = getLambda(cc.expected)
			if !(ok) {
				return cc.error(E_LambdaAssignedToIncompatible {
					TypeName: cc.describeType(cc.expected),
				})
			}
			var in, err1 = cc.productPatternMatch(L.Input, io.Input)
			if err1 != nil { return cc.propagate(err1) }
			var out, err2 = cc.checkChildExpr(io.Output, L.Output)
			if err2 != nil { return cc.propagate(err2) }
			var lambda_t = &typsys.NestedType { Content: io }
			return cc.assign(lambda_t, checked.Lambda { In: in, Out: out })
		}
	})
}

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


