package checker

import "kumachan/transformer/node"


func (impl Call) ExprVal() {}
type Call struct {
	Function  Expr
	Argument  Expr
}


func CheckCall(call node.Terms, ctx ExprContext) (SemiExpr, *ExprError) {
	var L = len(call.Terms)
	if L == 0 { panic("something went wrong") }
	if L == 1 {
		return CheckTerm(call.Terms[0], ctx)
	} else {
		var semi, err = CheckTerm(call.Terms[0], ctx)
		if err != nil { return SemiExpr{}, err }
		return ReduceCall(semi, call.Terms[:L-1], ctx)
	}
}

func ReduceCall(arg SemiExpr, terms []node.VariousTerm, ctx ExprContext) (SemiExpr, *ExprError) {
	var L = len(terms)
	if L == 0 {
		return arg, nil
	} else {
		var f_node = terms[L-1]
		var f, err1 = CheckTerm(f_node, ctx)
		if err1 != nil { return SemiExpr{}, err1 }
		var next_arg, err2 = CheckSingleCall(f, arg)
		if err2 != nil { return SemiExpr{}, err2 }
		return ReduceCall(next_arg, terms[:L-1], ctx)
	}
}

func CheckSingleCall(f SemiExpr, arg SemiExpr) (SemiExpr, *ExprError) {
	panic("not implemented")  // TODO
}


func DesugarExpr(expr node.Expr) node.Terms {
	return DesugarPipeline(expr.Call, expr.Pipeline)
}

func DesugarPipeline(left node.Terms, p node.MaybePipeline) node.Terms {
	var pipeline, ok = p.(node.Pipeline)
	if !ok {
		return left
	}
	var f = pipeline.Func
	var maybe_right = pipeline.Args
	var right, exists = maybe_right.(node.Terms)
	var arg node.Tuple
	if exists {
		arg = node.Tuple {
			Node:     pipeline.Operator.Node,
			Elements: []node.Expr {
				node.Expr {
					Node:     left.Node,
					Call:     left,
					Pipeline: nil,
				},
				node.Expr {
					Node:     right.Node,
					Call:     right,
					Pipeline: nil,
				},
			},
		}
	} else {
		arg = node.Tuple {
			Node:     left.Node,
			Elements: []node.Expr { {
				Node:     left.Node,
				Call:     left,
				Pipeline: nil,
			} },
		}
	}
	var current = node.Terms {
		Node:  pipeline.Node,
		Terms: []node.VariousTerm {
			f,
			node.VariousTerm {
				Node: arg.Node,
				Term: arg,
			},
		},
	}
	return DesugarPipeline(current, pipeline.Next)
}
