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
	var expr, err = (func() (Expr, *ExprError) {
		switch semi_value := semi.Value.(type) {
		case TypedExpr:
			return AssignTypedTo(expected, Expr(semi_value), ctx)
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
		case UntypedMacroInflation:
			return AssignMacroInflationTo(expected, semi_value, semi.Info, ctx)
		default:
			panic("impossible branch")
		}
	})()
	if err != nil { return Expr{}, err }
	if ctx.InferTypeArgs {
		return Expr {
			Type:  FillMarkedParams(expr.Type, ctx),
			Value: expr.Value,
			Info:  expr.Info,
		}, nil
	} else {
		return expr, nil
	}
}

func AssignTypedTo(expected Type, expr Expr, ctx ExprContext) (Expr, *ExprError) {
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
		var compare_named func(NamedType, NamedType) *ExprError
		compare_named = func(E NamedType, T NamedType) *ExprError {
			if !(ctx.InferTypeArgs) { panic("something went wrong") }
			if E.Name != T.Name {
				return throw("")
			}
			if len(E.Args) != len(T.Args) {
				return throw("")
			}
			var L = len(T.Args)
			for i := 0; i < L; i += 1 {
				var _, t_arg_is_wildcard = T.Args[i].(WildcardRhsType)
				if t_arg_is_wildcard {
					continue
				}
				switch E_arg := E.Args[i].(type) {
				case ParameterType:
					if E_arg.BeingInferred {
						var inferred, exists = ctx.Inferred[E_arg.Index]
						if exists {
							if !(CheckStrictAssignable(inferred, T.Args[i])) {
								return throw(fmt.Sprintf (
									"cannot infer type parameter %s",
									ctx.InferredNames[E_arg.Index],
								))
							}
						} else {
							ctx.Inferred[E_arg.Index] = T.Args[i]
						}
					} else {
						if !(CheckStrictAssignable(E.Args[i], T.Args[i])) {
							return throw("")
						}
					}
				case NamedType:
					switch T_arg := T.Args[i].(type) {
					case NamedType:
						var err = compare_named(E_arg, T_arg)
						if err != nil { return err }
					default:
						if !(CheckStrictAssignable(E.Args[i], T.Args[i])) {
							return throw("")
						}
					}
				default:
					if !(CheckStrictAssignable(E.Args[i], T.Args[i])) {
						return throw("")
					}
				}
			}
			return nil
		}
		var apply_params_mapping = func(args []Type, mapping []uint) []Type {
			var mapped = make([]Type, len(mapping))
			for i, j := range mapping {
				mapped[i] = args[j]
			}
			return mapped
		}
		// -- behavior of assigning a named type to an union type --
		var assign_union func(NamedType, NamedType, Union) (Expr, *ExprError)
		assign_union = func(exp NamedType, given NamedType, union Union) (Expr, *ExprError) {
			// 1. Find the given type in the list of case types of the union
			for index, case_type := range union.CaseTypes {
				if case_type.Name == given.Name {
					// 1.1. If found, check if type parameters matching
					var case_exp_type = NamedType {
						Name: case_type.Name,
						Args: apply_params_mapping(exp.Args, case_type.Params),
					}
					// 1.1.1. If not matching, throw an error
					if ctx.InferTypeArgs {
						var err = compare_named(case_exp_type, given)
						if err != nil { return Expr{}, err }
					} else {
						if !(CheckStrictAssignable(case_exp_type, given)) {
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
			for index, case_type := range union.CaseTypes {
				var g = ctx.ModuleInfo.Types[case_type.Name]
				var item_union, ok = g.Value.(Union)
				if ok {
					var item_expr, err = assign_union(NamedType {
						Name: case_type.Name,
						Args: apply_params_mapping(exp.Args, case_type.Params),
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
			return Expr{}, throw("given type is not a case type of the expected union type")
		}
		switch E := expected.(type) {
		case ParameterType:
			if ctx.InferTypeArgs && E.BeingInferred {
				var inferred, exists = ctx.Inferred[E.Index]
				if exists {
					if CheckStrictAssignable(inferred, expr.Type) {
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
				if E.Name == T.Name && ctx.InferTypeArgs {
					var err = compare_named(E, T)
					if err != nil {
						return Expr{}, err
					} else {
						return expr, nil
					}
				} else if E.Name != T.Name {
					var g = ctx.ModuleInfo.Types[E.Name]
					var union, is_union = g.Value.(Union)
					if is_union {
						return assign_union(E, T, union)
					} else {
						// "fallthrough" to strict check or unbox
					}
				}
			}
		}
		if CheckStrictAssignable(expected, expr.Type) {
			return expr, nil
		} else {
			var ctx_mod = ctx.ModuleInfo.Module.Name
			var reg = ctx.ModuleInfo.Types
			var unboxed, ok = Unbox(expr.Type, ctx_mod, reg).(Unboxed)
			if !ok {
				return Expr{}, throw("")
			} else {
				var expr_unboxed = Expr {
					Type:  unboxed.Type,
					Value: expr.Value,
					Info:  expr.Info,
				}
				var expr_expected, err = AssignTypedTo (
					expected, expr_unboxed, ctx,
				)
				// if err != nil { return Expr{}, err }
				if err != nil { return Expr{}, throw("") }
				return expr_expected, nil
			}
		}
	}
}

func CheckStrictAssignable(expected Type, given Type) bool {
	switch T := given.(type) {
	case WildcardRhsType:
		return true
	case ParameterType:
		switch E := expected.(type) {
		case ParameterType:
			return T.Index == E.Index
		default:
			return false
		}
	case NamedType:
		switch E := expected.(type) {
		case NamedType:
			if E.Name == T.Name {
				var L1 = len(T.Args)
				var L2 = len(E.Args)
				if L1 != L2 { panic("type registration went wrong") }
				var L = L1
				for i := 0; i < L; i += 1 {
					if !(CheckStrictAssignable(E.Args[i], T.Args[i])) {
						return false
					}
				}
				return true
			} else {
				return false
			}
		default:
			return false
		}
	case AnonymousType:
		switch E := expected.(type) {
		case AnonymousType:
			switch rt := T.Repr.(type) {
			case Unit:
				switch E.Repr.(type) {
				case Unit:
					return true
				default:
					return false
				}
			case Tuple:
				switch re := E.Repr.(type) {
				case Tuple:
					var L1 = len(rt.Elements)
					var L2 = len(re.Elements)
					if L1 == L2 {
						var L = L1
						for i := 0; i < L; i += 1 {
							if !(CheckStrictAssignable(re.Elements[i], rt.Elements[i])) {
								return false
							}
						}
						return true
					} else {
						return false
					}
				default:
					return false
				}
			case Bundle:
				switch re := E.Repr.(type) {
				case Bundle:
					if len(rt.Fields) == len(re.Fields) {
						for name, ft := range rt.Fields {
							var fe, exists = re.Fields[name]
							if !exists || !(CheckStrictAssignable(fe.Type, ft.Type)) {
								return false
							}
						}
						return true
					} else {
						return false
					}
				default:
					return false
				}
			case Func:
				switch re := E.Repr.(type) {
				case Func:
					if !(CheckStrictAssignable(re.Input, rt.Input)) {
						return false
					}
					if !(CheckStrictAssignable(re.Output, rt.Output)) {
						return false
					}
					return true
				default:
					return true
				}
			default:
				panic("impossible branch")
			}
		default:
			return false
		}
	default:
		panic("impossible branch")
	}
}
