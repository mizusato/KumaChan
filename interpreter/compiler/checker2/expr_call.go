package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/checked"
)


func checkCall(callee ast.Expr, arg ast.Expr, loc source.Location) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		var ref, is_ref = getInlineRef(callee)
		if is_ref {
			// TODO
		} else {
			var callee_expr, err1 = cc.checkChildExpr(nil, callee)
			if err1 != nil { return cc.propagate(err1) }
			var io, callable = cc.unboxLambda(callee_expr)
			if !(callable) {
				// TODO
			}
			var arg_expr, err2 = cc.checkChildExpr(io.Input, arg)
			if err2 != nil { return cc.propagate(err2) }
			return cc.assign(io.Output, checked.Call {
				Callee:   callee_expr,
				Argument: arg_expr,
			})
		}
	})
}

func getInlineRef(expr ast.Expr) (ast.InlineRef, bool) {
	if expr.Pipeline == nil {
		var inline_ref, ok = expr.Term.Term.(ast.InlineRef)
		return inline_ref, ok
	} else {
		return ast.InlineRef {}, false
	}
}


