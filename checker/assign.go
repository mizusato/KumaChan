package checker

import (
	"fmt"
)


func RequireExplicitType(t Type, info ExprInfo) *ExprError {
	if t == nil {
		return &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_ExplicitTypeRequired {},
		}
	} else {
		return nil
	}
}

func AssignTo(expected Type, semi SemiExpr, ctx ExprContext) (Expr, *ExprError) {
	switch semi_value := semi.Value.(type) {
	case TypedExpr:
		return AssignTypedTo(expected, Expr(semi_value), ctx, true)
	case UntypedLambda:
		return AssignLambdaTo(expected, semi_value, semi.Info, ctx)
	case UntypedInteger:
		return AssignIntegerTo(expected, semi_value, semi.Info, ctx)
	case SemiTypedTuple:
		return AssignTupleTo(expected, semi_value, semi.Info, ctx)
	case SemiTypedBundle:
		return AssignBundleTo(expected, semi_value, semi.Info, ctx)
	case SemiTypedArray:
		return AssignArrayTo(expected, semi_value, semi.Info, ctx)
	case SemiTypedBlock:
		return AssignBlockTo(expected, semi_value, semi.Info, ctx)
	case SemiTypedMatch:
		return AssignMatchTo(expected, semi_value, semi.Info, ctx)
	case UntypedRef:
		return AssignRefTo(expected, semi_value, semi.Info, ctx)
	case UndecidedCall:
		return AssignCallTo(expected, semi_value, semi.Info, ctx)
	default:
		panic("impossible branch")
	}
}

func AssignTypedTo(expected Type, expr Expr, ctx ExprContext, unbox bool) (Expr, *ExprError) {
	// TODO: update comments
	if expected == nil {
		// 1. If the expected type is not specified,
		//    no further process is required.
		return expr, nil
	} else {
		// 3. Otherwise, try some implicit type conversions
		// 3.1. Define some inner functions
		// -- shortcut to produce a "not assignable" error --
		var throw = func(reason string) *ExprError {
			return &ExprError {
				Point:    expr.Info.ErrorPoint,
				Concrete: E_NotAssignable {
					From:   ctx.DescribeType(expr.Type),
					To:     ctx.DescribeExpectedType(expected),
					Reason: reason,
				},
			}
		}
		// -- behavior of assigning a named type to an union type --
		var assign_union func(NamedType, NamedType, Union) (Expr, *ExprError)
		assign_union = func(exp NamedType, given NamedType, union Union) (Expr, *ExprError) {
			// 1. Find the given type in the list of subtypes of the union
			for index, subtype := range union.SubTypes {
				if subtype == given.Name {
					// 1.1. If found, check if type parameters are identical
					if len(exp.Args) != len(given.Args) {
						panic("something went wrong")
					}
					var L = len(exp.Args)
					for i := 0; i < L; i += 1 {
						var arg_exp = exp.Args[i]
						var arg_given = given.Args[i]
						if !(AreTypesEqualInSameCtx(arg_exp, arg_given)) {
							// 1.1.1. If not identical, throw an error.
							return Expr{}, throw("type parameters not matching")
						}
					}
					// 1.1.2. Otherwise, return a lifted value.
					return Expr {
						Type:  exp,
						Value: Sum { Value: expr, Index: uint(index) },
						Info:  expr.Info,
					}, nil
				}
			}
			for index, subtype := range union.SubTypes {
				var t = ctx.ModuleInfo.Types[subtype]
				var item_union, ok = t.Value.(Union)
				if ok {
					var item_expr, err = assign_union(NamedType {
						Name: subtype,
						Args: exp.Args,
					}, given, item_union)
					if err != nil { continue }
					return Expr {
						Type:  exp,
						Value: Sum { Value: item_expr, Index: uint(index) },
						Info:  expr.Info,
					}, nil
				}
			}
			// 1.2. Otherwise, throw an error.
			return Expr{}, throw("given type is not a subtype of the expected union type")
		}
		switch E := expected.(type) {
		case ParameterType:
			if ctx.InferTypeArgs {
				var inferred, exists = ctx.Inferred[E.Index]
				if exists {
					if AreTypesEqualInSameCtx(inferred, expected) {
						return expr, nil
					} else {
						return Expr{}, throw(fmt.Sprintf (
							"cannot infer type parameter %s",
							ctx.InferredNames[E.Index],
						))
					}
				} else {
					ctx.Inferred[E.Index] = expr.Type
					return expr, nil
				}
			}
		case NamedType:
			switch T := expr.Type.(type) {
			case NamedType:
				if T.Name == E.Name {
					if len(T.Args) != len(E.Args) {
						return Expr{}, throw("quantity of type parameters not matching")
					}
					var L = len(T.Args)
					if ctx.InferTypeArgs {
						for i := 0; i < L; i += 1 {
							switch E_arg := E.Args[i].(type) {
							case ParameterType:
								var inferred, exists = ctx.Inferred[E_arg.Index]
								if exists {
									if !AreTypesEqualInSameCtx(inferred, expected) {
										return Expr{}, throw(fmt.Sprintf(
											"cannot infer type parameter %s",
											ctx.InferredNames[E_arg.Index],
										))
									}
								} else {
									ctx.Inferred[E_arg.Index] = T.Args[i]
								}
							default:
								// TODO: a recursion required if E.Args[i] and T.Args[i] are both named types
								if !(AreTypesEqualInSameCtx(T.Args[i], E.Args[i])) {
									return Expr{}, throw("")
								}
							}
						}
					} else {
						for i := 0; i < L; i += 1 {
							if !(AreTypesEqualInSameCtx(T.Args[i], E.Args[i])) {
								return Expr{}, throw("")
							}
						}
					}
					return expr, nil
				} else {
					var g = ctx.ModuleInfo.Types[E.Name]
					var union, is_union = g.Value.(Union)
					if is_union {
						return assign_union(E, T, union)
					}
				}
			}
		}
		if AreTypesEqualInSameCtx(expected, expr.Type) {
			return expr, nil
		} else {
			if unbox {
				var unboxed, ok = Unbox(expr.Type, ctx).(Unboxed)
				if ok {
					var expr_unboxed = Expr {
						Type:  unboxed.Type,
						Value: expr.Value,
						Info:  expr.Info,
					}
					var expr_expected, err = AssignTypedTo (
						expected, expr_unboxed, ctx, false,
					)
					if err != nil { return Expr{}, err }
					if ctx.UnboxCounted {
						*(ctx.UnboxCount) += 1
					}
					return expr_expected, nil
				}
			}
			return Expr{}, throw("")
		}
	}
}
