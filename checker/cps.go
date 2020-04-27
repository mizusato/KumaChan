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
	var cont ast.VariousTerm
	if exists {
		cont = ast.VariousTerm {
			Node: binding.Node,
			Term: ast.Lambda {
				Node:   binding.Node,
				Input:  binding.Pattern,
				Output: output,
			},
		}
	} else {
		cont = output
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
						WrapTermAsExpr(cont),
					},
				},
			},
			Arg:  nil,
		},
	}
}

func WrapTermAsExpr(term ast.VariousTerm) ast.Expr {
	return ast.Expr {
		Node:     term.Node,
		Call:     ast.Call {
			Node: term.Node,
			Func: term,
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
