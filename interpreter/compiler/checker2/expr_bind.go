package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
	"kumachan/interpreter/lang/common/source"
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

func checkBlock(B ast.Block) ExprChecker {
	return makeExprCheckerWithLocalScope(B.Location, func(cc *checkContextWithLocalScope) checkResult {
		var cons_ctx = cc.typeConsContext()
		if len(B.Bindings) == 0 { panic("something went wrong") }
		var let_list = make([] checked.Let, len(B.Bindings))
		for i, binding := range B.Bindings {
			var declared_t, err = (func() (typsys.Type, *source.Error) {
				var type_node, type_declared = binding.Type.(ast.VariousType)
				if type_declared {
					return newType(type_node, cons_ctx)
				} else {
					return nil, nil
				}
			})()
			if err != nil { return cc.propagate(err) }
			var recursive = binding.Recursive
			var pattern_node = binding.Pattern
			var value_node = binding.Value
			if declared_t != nil && recursive {
				if !(isLambdaExprNode(value_node)) {
					return cc.propagate(source.MakeError(value_node.Location,
						E_NonLambdaRecursive {}))
				}
				var t = declared_t
				var pattern, err1 = cc.productPatternMatch(pattern_node, t)
				if err1 != nil { return cc.propagate(err1) }
				var expr, err2 = cc.checkChildExpr(t, value_node)
				if err2 != nil { return cc.propagate(err2) }
				let_list[i] = checked.Let {
					Pattern: pattern,
					Value:   expr,
				}
			} else {
				var expr, err1 = cc.checkChildExpr(declared_t, value_node)
				if err1 != nil { return cc.propagate(err1) }
				var t = (func() typsys.Type {
					if declared_t == nil {
						return expr.Type
					} else {
						return declared_t
					}
				})()
				var pattern, err2 = cc.productPatternMatch(pattern_node, t)
				if err2 != nil { return cc.propagate(err2) }
				let_list[i] = checked.Let {
					Pattern: pattern,
					Value:   expr,
				}
			}
		}
		return cc.assignFinalExpr(B.Return, func(ret *checked.Expr) checked.ExprContent {
			return checked.Block {
				LetList: let_list,
				Return:  ret,
			}
		})
	})
}

func isLambdaExprNode(expr ast.Expr) bool {
	if expr.Pipeline != nil {
		return false
	}
	var _, is_lambda = expr.Term.Term.(ast.Lambda)
	return is_lambda
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


