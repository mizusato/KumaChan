package checker

import (
	"kumachan/parser/scanner"
	"kumachan/parser/ast"
)


type UntypedCall struct {
	Callee    SemiExpr
	Argument  ast.Call
	Context   ExprContext
}

func (impl Call) ExprVal() {}
type Call struct {
	Function  Expr
	Argument  Expr
}


func CheckCall(call ast.Call, ctx ExprContext) (SemiExpr, *ExprError) {
	var arg_node, has_arg = call.Arg.(ast.Call)
	if has_arg {
		arg, err := CheckCall(arg_node, ctx)
		if err != nil { return SemiExpr{}, err }
		callee, err := CheckTerm(call.Func, ctx)
		if err != nil { return SemiExpr{}, err }
		var info = ctx.GetExprInfo(call.Node)
		var f_info = callee.Info
		switch f := callee.Value.(type) {
		case TypedExpr:
			var t = f.Type
			var r, ok = UnboxFunc(t, ctx).(Func)
			if ok {
				var arg_typed, err = AssignTo(r.Input, arg, ctx)
				if err != nil { return SemiExpr{}, err }
				var typed = Expr {
					Type:  r.Output,
					Value: Call {
						Function: Expr(f),
						Argument: arg_typed,
					},
					Info:  info,
				}
				return LiftTyped(typed), nil
			} else {
				return SemiExpr{}, &ExprError {
					Point:    f_info.ErrorPoint,
					Concrete: E_ExprTypeNotCallable {
						Type: ctx.DescribeType(t),
					},
				}
			}
		case UntypedLambda:
			var typed, err = CallUntypedLambda(arg, f, f_info, info, ctx)
			if err != nil { return SemiExpr{}, err }
			return LiftTyped(typed), nil
		case UntypedRef:
			return CallUntypedRef(arg, f, f_info, info, ctx)
		case SemiTypedSwitch,
			SemiTypedBlock:
			return SemiExpr{}, &ExprError {
				Point:    f_info.ErrorPoint,
				Concrete: E_ExplicitTypeRequired {},
			}
		default:
			return SemiExpr{}, &ExprError {
				Point:    f_info.ErrorPoint,
				Concrete: E_ExprNotCallable {},
			}
		}
	} else {
		return CheckTerm(call.Func, ctx)
	}
}

func CheckInfix(infix ast.Infix, ctx ExprContext) (SemiExpr, *ExprError) {
	return CheckCall(DesugarInfix(infix), ctx)
}


func CraftAstCallExpr(f ast.VariousTerm, arg ast.VariousTerm, node ast.Node) ast.Expr {
	var call = ast.Call {
		Node: node,
		Func: f,
		Arg:  ast.Call {
			Node: arg.Node,
			Func: arg,
			Arg:  nil,
		},
	}
	return ast.WrapCallAsExpr(call)
}

func DesugarExpr(expr ast.Expr) ast.Call {
	return DesugarPipeline(DesugarTerms(expr.Terms), expr.Pipeline)
}

func DesugarTerms(terms ast.Terms) ast.Call {
	if len(terms.Terms) == 0 { panic("something went wrong") }
	var callee = terms.Terms[0]
	var args = terms.Terms[1:]
	if len(args) == 0 {
		return ast.Call {
			Node: terms.Node,
			Func: callee,
			Arg:  nil,
		}
	} else if len(args) == 1 {
		return ast.Call {
			Node: terms.Node,
			Func: callee,
			Arg:  ast.Call {
				Node: terms.Node,
				Func: args[0],
				Arg:  nil,
			},
		}
	} else {
		var elements = make([] ast.Expr, len(args))
		for i, arg := range args {
			elements[i] = ast.WrapCallAsExpr(ast.Call {
				Node: arg.Node,
				Func: arg,
				Arg:  nil,
			})
		}
		return ast.Call {
			Node: terms.Node,
			Func: callee,
			Arg:  ast.Call {
				Node: terms.Node,
				Func: ast.VariousTerm {
					Node: terms.Node,
					Term: ast.Tuple {
						Node:     terms.Node,
						Elements: elements,
					},
				},
				Arg:  nil,
			},
		}
	}
}

func DesugarPipeline(left ast.Call, p ast.MaybePipeline) ast.Call {
	var pipeline, ok = p.(ast.Pipeline)
	if !ok {
		return left
	}
	var f = pipeline.Func
	var maybe_right = pipeline.Arg
	var right, exists = maybe_right.(ast.Terms)
	var arg ast.VariousTerm
	var current_node ast.Node
	if exists {
		var elements = make([] ast.Expr, 0, (1 + len(right.Terms)))
		elements = append(elements, ast.WrapCallAsExpr(left))
		for _, r := range right.Terms {
			elements = append(elements, ast.WrapTermAsExpr(r))
		}
		arg = ast.VariousTerm {
			Node: pipeline.Operator.Node,
			Term: ast.Tuple {
				Node:     pipeline.Operator.Node,
				Elements: elements,
			},
		}
		current_node = ast.Node {
			CST:   pipeline.Node.CST,
			Point: pipeline.Node.Point,
			Span:  scanner.Span {
				Start: pipeline.Node.Span.Start,
				End:   right.Span.End,
			},
		}
	} else {
		arg = ast.WrapCallAsTerm(left)
		current_node = ast.Node {
			CST:   pipeline.Node.CST,
			Point: pipeline.Node.Point,
			Span:  scanner.Span {
				Start: pipeline.Node.Span.Start,
				End:   pipeline.Func.Span.End,
			},
		}
	}
	var current = ast.Call {
		Node:  current_node,
		Func:  f,
		Arg:   ast.Call {
			Node: arg.Node,
			Func: arg,
			Arg:  nil,
		},
	}
	return DesugarPipeline(current, pipeline.Next)
}

func DesugarInfix(infix ast.Infix) ast.Call {
	return ast.Call {
		Node: infix.Node,
		Func: infix.Operator,
		Arg:  ast.Call {
			Node: infix.Node,
			Func: ast.VariousTerm {
				Node: infix.Node,
				Term: ast.Tuple {
					Node:     infix.Node,
					Elements: []ast.Expr {
						ast.WrapCallAsExpr(ast.Call {
							Node: infix.Operand1.Node,
							Func: infix.Operand1,
							Arg:  nil,
						}),
						ast.WrapCallAsExpr(ast.Call {
							Node: infix.Operand2.Node,
							Func: infix.Operand2,
							Arg:  nil,
						}),
					},
				},
			},
			Arg:  nil,
		},
	}
}

