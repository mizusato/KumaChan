package checker

import "kumachan/interpreter/lang/ast"


func CheckCps(cps ast.Cps, ctx ExprContext) (SemiExpr, *ExprError)  {
	return Check(DesugarCps(cps), ctx)
}

func DesugarCps(cps ast.Cps) ast.Expr {
	var input = cps.Input
	var output = cps.Output
	var callee = ast.VariousTerm {
		Node: cps.Callee.Node,
		Term: cps.Callee,
	}
	var binding, exists = cps.Binding.(ast.CpsBinding)
	var cont ast.Expr
	if exists {
		cont = ast.WrapTermAsExpr(ast.VariousTerm {
			Node: binding.Node,
			Term: ast.Lambda {
				Node:   binding.Node,
				Input:  binding.Pattern,
				Output: output,
			},
		})
	} else {
		cont = ast.WrapTermAsExpr(ast.VariousTerm {
			Node: cps.Output.Node,
			Term: ast.Lambda {
				Node:   cps.Output.Node,
				Input:  ast.VariousPattern {
					Node:    cps.Output.Node,
					Pattern: ast.PatternTrivial {
						Node: cps.Output.Node,
						Name: ast.Identifier {
							Node: cps.Output.Node,
							Name: ([] rune)(IgnoreMark),
						},
					},
				},
				Output: output,
			},
		})
	}
	var arg = ast.VariousTerm {
		Node: cps.Node,
		Term: ast.Tuple {
			Node: cps.Node,
			Elements: [] ast.Expr {
				input,
				cont,
			},
		},
	}
	return ast.WrapTermAsExpr(ast.VariousTerm {
		Node: cps.Node,
		Term: ast.VariousCall {
			Node: cps.Node,
			Call: ast.CallPrefix {
				Node:     cps.Node,
				Callee:   ast.WrapTermAsExpr(callee),
				Argument: ast.WrapTermAsExpr(arg),
			},
		},
	})
}

