package checker

import "kumachan/parser/ast"


func CheckCps(cps ast.Cps, ctx ExprContext) (SemiExpr, *ExprError)  {
	return CheckCall(DesugarCps(cps), ctx)
}

func DesugarCps(cps ast.Cps) ast.Call {
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
	return ast.Call {
		Node: cps.Node,
		Func: callee,
		Arg:  ast.Call {
			Node: cps.Node,
			Func: ast.VariousTerm {
				Node: cps.Node,
				Term: ast.Tuple {
					Node: cps.Node,
					Elements: []ast.Expr {
						input,
						cont,
					},
				},
			},
			Arg:  nil,
		},
	}
}

