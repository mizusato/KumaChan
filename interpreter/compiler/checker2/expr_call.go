package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/checked"
	"kumachan/interpreter/compiler/checker2/typsys"
)


func checkCall1(callee ast.Expr, arg ast.Expr, loc source.Location) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		var call_certain = func() checkResult {
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
		var call_overload = func(R FuncRefs) checkResult {
			if len(R.Functions) == 0 { panic("something went wrong") }
			var ctx = cc.exprContext
			var results = make([] checkResult, len(R.Functions))
			for i, f := range R.Functions {
				var params = f.Signature.TypeParameters
				var io = f.Signature.InputOutput
				var in_t = io.Input
				var out_t = io.Output
				var result = cc.infer(params, out_t, func(s0 *typsys.InferringState) (checked.ExprContent, *typsys.InferringState, *source.Error) {
					var arg_expr, s1, err1 = check(arg)(in_t, s0, ctx)
					if err1 != nil { return nil, nil, err1 }
					var f_expr, s2, err2 = makeFuncRef(f, s1, nil, loc, ctx)
					if err2 != nil { return nil, nil, err2 }
					return checked.Call {
						Callee:   f_expr,
						Argument: arg_expr,
					}, s2, nil
				})
				results[i] = result
			}
			var ok_result *checkResult
			for i, result := range results {
				if result.err == nil {
					if ok_result == nil {
						ok_result = &(results[i])
					} else {
						break
					}
				}
			}
			if ok_result != nil {
				return *ok_result
			} else {
				// TODO
			}
		}
		var ref_node, is_ref = getInlineRef(callee)
		if is_ref {
			var ref, err = cc.resolveInlineRef(ref_node, nil)
			if err != nil { return cc.propagate(err) }
			switch R := ref.(type) {
			case FuncRefs:
				return call_overload(R)
			case LocalRef:
				return call_certain()
			case LocalRefWithFuncRefs:
				return call_certain()
			default:
				panic("impossible branch")
			}
		} else {
			return call_certain()
		}
	})
}

func checkCall2(callee ast.Expr, arg ast.Expr, pivot *checked.Expr, loc source.Location) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		if pivot == nil { panic("something went wrong") }
		var ref, is_ref = getInlineRef(callee)
		if is_ref {
			// TODO (use pivot.Type to lookup name)
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


