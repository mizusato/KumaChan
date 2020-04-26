package checker

import "kumachan/transformer/ast"


func CheckCps(cps ast.Cps, ctx ExprContext) (SemiExpr, *ExprError)  {
	return CheckCall(DesugarCps(cps), ctx)
}

func DesugarCps(cps ast.Cps) ast.Call {
	var input = cps.Input
	var output = WrapExprAsTerm(cps.Output)
	var callee = ast.VariousTerm {
		Node: cps.Callee.Node,
		Term: cps.Callee,
	}
	var binding, exists = cps.Binding.(ast.CpsBinding)
	var lambda ast.Expr
	if exists {
		lambda = WrapTermAsExpr(ast.Lambda {
			Node:   binding.Node,
			Input:  binding.Pattern,
			Output: output,
		}, binding.Node)
	} else {
		lambda = WrapTermAsExpr(ast.Lambda {
			Node:   output.Node,
			Input:  ast.VariousPattern {
				Node:    input.Node,
				Pattern: ast.PatternTrivial {
					Node: input.Node,
					Name: ast.Identifier {
						Node: input.Node,
						Name: ([] rune)(IgnoreMark),
					},
				},
			},
			Output: output,
		}, output.Node)
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
						lambda,
					},
				},
			},
			Arg:  nil,
		},
	}
}

func WrapTermAsExpr(term ast.Term, node ast.Node) ast.Expr {
	return ast.Expr {
		Node:     node,
		Call:     ast.Call {
			Node: node,
			Func: ast.VariousTerm {
				Node: node,
				Term: term,
			},
			Arg:  nil,
		},
		Pipeline: nil,
	}
}

func WrapExprAsTerm(expr ast.Expr) ast.VariousTerm {
	return ast.VariousTerm {
		Node: expr.Node,
		Term: ast.Tuple {
			Node:     expr.Node,
			Elements: [] ast.Expr { expr },
		},
	}
}
