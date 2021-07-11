package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/checked"
	"kumachan/interpreter/compiler/checker2/typsys"
)


func checkCall1(callee ast.Expr, arg ast.Expr, loc source.Location) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		var ref, is_ref = getInlineRef(callee)
		if is_ref {
			// TODO
		} else {
			var callee_expr, err1 = cc.checkChildExpr(nil, callee)
			if err1 != nil { return cc.propagate(err1) }
			var io, callable = cc.unboxLambda(callee_expr)
			if !(callable) {
				return cc.error(
					E_TypeNotCallable {})
			}
			var in = io.Input
			var out = io.Output
			var arg_expr, err2 = cc.checkChildExpr(in, arg)
			if err2 != nil { return cc.propagate(err2) }
			return cc.assign(out, checked.Call {
				Callee:   callee_expr,
				Argument: arg_expr,
			})
		}
	})
}

func checkCall2(callee ast.Expr, arg ast.Expr, pivot *checked.Expr, loc source.Location) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		if pivot == nil { panic("something went wrong") }
		var ref, is_ref = getInlineRef(callee)
		if is_ref {
			// TODO
		} else {
			var callee_expr, err1 = cc.checkChildExpr(nil, callee)
			if err1 != nil { return cc.propagate(err1) }
			var io, callable = cc.unboxLambda(callee_expr)
			if !(callable) {
				return cc.error(
					E_TypeNotCallable {})
			}
			var in = io.Input
			var out = io.Output
			var tuple_exp, ok = getTuple(cc.expected)
			if !(ok) {
				// TODO
			}
			if len(tuple_exp.Elements) != 2 {
				// TODO
			}
			var pivot_exp = tuple_exp.Elements[0]
			var arg_exp = tuple_exp.Elements[1]
			var err2 = cc.assignType(pivot_exp, pivot.Type)
			if err2 != nil { return cc.propagate(err2) }
			var arg_expr, err3 = cc.checkChildExpr(arg_exp, arg)
			if err3 != nil { return cc.propagate(err3) }
			var pair_t = &typsys.NestedType {
				Content: typsys.Tuple { Elements: [] typsys.Type {
					pivot.Type,
					arg_expr.Type,
				} },
			}
			var pair = &checked.Expr {
				Type:    pair_t,
				Info:    checked.ExprInfoFrom(loc),
				Content: checked.Tuple { Elements: [] *checked.Expr {
					pivot,
					arg_expr,
				} },
			}
			var err4 = cc.assignType(in, pair_t)
			if err4 != nil { return cc.propagate(err4) }
			return cc.assign(out, checked.Call {
				Callee:   callee_expr,
				Argument: pair,
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


