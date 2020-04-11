package checker

import (
	"kumachan/parser/scanner"
	"kumachan/transformer/ast"
)


func (impl UndecidedCall) SemiExprVal() {}
type UndecidedCall struct {
	Options   [] AvailableCall
	FuncName  string
}
type AvailableCall struct {
	Expr      Expr
	IsExact   bool
	Function  *GenericFunction
}
type AssignableCall struct {
	Expr      Expr
	IsExact   bool
	Function  *GenericFunction
}

func (impl Call) ExprVal() {}
type Call struct {
	Function  Expr
	Argument  Expr
}


func CheckCall(call ast.Call, ctx ExprContext) (SemiExpr, *ExprError) {
	var get_macro_ref = func(semi SemiExpr) (UntypedRefToMacro, bool) {
		switch ref := semi.Value.(type) {
		case UntypedRef:
			switch ref_body := ref.RefBody.(type) {
			case UntypedRefToMacro:
				return ref_body, true
			}
		}
		return UntypedRefToMacro{}, false
	}
	var arg_node, has_arg = call.Arg.(ast.Call)
	if has_arg {
		var info = ctx.GetExprInfo(call.Node)
		var f, err = CheckTerm(call.Func, ctx)
		if err != nil { return SemiExpr{}, err }
		var macro_ref, is_macro_ref = get_macro_ref(f)
		if is_macro_ref {
			return SemiExpr {
				Value: UntypedMacroInflation {
					Macro:     macro_ref.Macro,
					MacroName: macro_ref.MacroName,
					Arguments: AdaptMacroArgs(arg_node),
					Point:     info.ErrorPoint,
					Context:   ctx,
				},
				Info:  info,
			}, nil
		} else {
			var arg, err = CheckCall(arg_node, ctx)
			if err != nil { return SemiExpr{}, err }
			return CheckSingleCall(f, arg, info, ctx)
		}
	} else {
		return CheckTerm(call.Func, ctx)
	}
}

func CheckSingleCall(f SemiExpr, arg SemiExpr, info ExprInfo, ctx ExprContext) (SemiExpr, *ExprError) {
	switch f_concrete := f.Value.(type) {
	case TypedExpr:
		var t = f_concrete.Type
		switch T := t.(type) {
		case AnonymousType:
			switch r := T.Repr.(type) {
			case Func:
				var arg_typed, err = AssignTo(r.Input, arg, ctx)
				if err != nil { return SemiExpr{}, err }
				return LiftTyped(Expr {
					Type:  r.Output,
					Value: Call {
						Function: Expr(f_concrete),
						Argument: arg_typed,
					},
					Info:  f.Info,
				}), nil
			}
		}
		return SemiExpr{}, &ExprError {
			Point:    f.Info.ErrorPoint,
			Concrete: E_ExprTypeNotCallable {
				Type: ctx.DescribeType(t),
			},
		}
	case UntypedLambda:
		return CallUntypedLambda(arg, f_concrete, f.Info, info, ctx)
	case UntypedRef:
		return CallUntypedRef(arg, f_concrete, f.Info, info, ctx)
	case SemiTypedMatch,
		 SemiTypedBlock,
		 UndecidedCall:
		return SemiExpr{}, &ExprError {
			Point:    f.Info.ErrorPoint,
			Concrete: E_ExplicitTypeRequired {},
		}
	default:
		return SemiExpr{}, &ExprError {
			Point:    f.Info.ErrorPoint,
			Concrete: E_ExprNotCallable {},
		}
	}
}

func CheckInfix(infix ast.Infix, ctx ExprContext) (SemiExpr, *ExprError) {
	return CheckCall(DesugarInfix(infix), ctx)
}


func AssignCallTo(expected Type, call UndecidedCall, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var err = RequireExplicitType(expected, info)
	if err != nil { return Expr{}, err }
	var types_desc = make([]string, 0)
	var assignable = make([]AssignableCall, 0)
	for _, option := range call.Options {
		var expr, err = AssignTypedTo(expected, option.Expr, ctx)
		if err != nil {
			types_desc = append (
				types_desc,
				ctx.DescribeType(option.Expr.Type),
			)
		} else {
			var is_exact = AreTypesEqualInSameCtx(expr.Type, option.Expr.Type)
			assignable = append(assignable, AssignableCall{
				Expr:     expr,
				IsExact:  is_exact,
				Function: option.Function,
			})
		}
	}
	if len(assignable) == 0 {
		return Expr{}, &ExprError{
			Point: info.ErrorPoint,
			Concrete: E_NoneOfTypesAssignable{
				From: types_desc,
				To:   ctx.DescribeExpectedType(expected),
			},
		}
	} else {
		var exact_quantity = 0
		var exact = -1
		for i, a := range assignable {
			if a.IsExact {
				exact_quantity += 1
				exact = i
			}
		}
		if exact_quantity == 1 {
			return assignable[exact].Expr, nil
		} else {
			var candidates = make([]string, len(assignable))
			for i, a := range assignable {
				candidates[i] = DescribeCandidate(call.FuncName, a.Function)
			}
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_AmbiguousCall {
					Candidates: candidates,
				},
			}
		}
	}
}


func DesugarExpr(expr ast.Expr) ast.Call {
	return DesugarPipeline(expr.Call, expr.Pipeline)
}

func DesugarPipeline(left ast.Call, p ast.MaybePipeline) ast.Call {
	var pipeline, ok = p.(ast.Pipeline)
	if !ok {
		return left
	}
	var f = pipeline.Func
	var maybe_right = pipeline.Arg
	var right, exists = maybe_right.(ast.Call)
	var arg ast.Tuple
	var current_node ast.Node
	if exists {
		arg = ast.Tuple {
			Node:     pipeline.Operator.Node,
			Elements: []ast.Expr {
				ast.Expr {
					Node:     left.Node,
					Call:     left,
					Pipeline: nil,
				},
				ast.Expr {
					Node:     right.Node,
					Call:     right,
					Pipeline: nil,
				},
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
		arg = ast.Tuple {
			Node:     left.Node,
			Elements: []ast.Expr { {
				Node:     left.Node,
				Call:     left,
				Pipeline: nil,
			} },
		}
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
			Func: ast.VariousTerm {
				Node: arg.Node,
				Term: arg,
			},
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
						ast.Expr {
							Node:     infix.Operand1.Node,
							Call:     ast.Call {
								Node: infix.Operand1.Node,
								Func: infix.Operand1,
								Arg:  nil,
							},
							Pipeline: nil,
						},
						ast.Expr {
							Node:     infix.Operand2.Node,
							Call:     ast.Call {
								Node: infix.Operand2.Node,
								Func: infix.Operand2,
								Arg:  nil,
							},
							Pipeline: nil,
						},
					},
				},
			},
			Arg:  nil,
		},
	}
}

// TODO: DesugarProcedure
