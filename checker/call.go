package checker

import "kumachan/transformer/node"


func (impl Call) ExprVal() {}
type Call struct {
	Function  Expr
	Argument  Expr
}


func CheckCall(call node.Call, ctx ExprContext) (SemiExpr, *ExprError) {
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


func DesugarExpr(expr node.Expr) node.Call {
	return DesugarPipeline(expr.Call, expr.Pipeline)
}

func DesugarPipeline(input node.Call, p node.MaybePipeline) node.Call {
	var pipeline, ok = p.(node.Pipeline)
	if !ok {
		return input
	}
	var right_terms = pipeline.Call.Terms
	if len(right_terms) == 0 { panic("something went wrong") }
	var f = right_terms[0]
	var rest_args = right_terms[1:]
	var L = (1 + len(rest_args))
	var args = make([]node.Expr, L)
	for i := 0; i < L; i += 1 {
		if i == 0 {
			args[i] = node.Expr {
				Node:     input.Node,
				Call:     input,
				Pipeline: nil,
			}
		} else {
			var rest_arg = rest_args[i-1]
			args[i] = node.Expr {
				Node:     rest_arg.Node,
				Call:     node.Call {
					Terms: []node.VariousTerm { rest_arg },
				},
				Pipeline: nil,
			}
		}
	}
	var args_tuple = node.Tuple {
		Node:     pipeline.Operator.Node,
		Elements: args,
	}
	var current = node.Call {
		Node:  pipeline.Node,
		Terms: []node.VariousTerm {
			f,
			node.VariousTerm {
				Node: args_tuple.Node,
				Term: args_tuple,
			},
		},
	}
	return DesugarPipeline(current, pipeline.Next)
}
