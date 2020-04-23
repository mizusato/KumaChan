package checker


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
		case SemiTypedSwitch:
			return AssignSwitchTo(expected, semi_value, semi.Info, ctx)
		case SemiTypedMultiSwitch:
			return AssignMultiSwitchTo(expected, semi_value, semi.Info, ctx)
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
	return expr, nil
}

func AssignTypedTo(expected Type, expr Expr, ctx ExprContext) (Expr, *ExprError) {
	// TODO: update comments
	var failed = func() (Expr, *ExprError) {
		return Expr{}, &ExprError {
			Point:    expr.Info.ErrorPoint,
			Concrete: E_NotAssignable {
				From:   ctx.DescribeType(expr.Type),
				To:     ctx.DescribeExpectedType(expected),
				Reason: "",  // reserved for possible use in the future
			},
		}
	}
	var get_case_args = func(union_args []Type, mapping []uint) []Type {
		var mapped = make([]Type, len(mapping))
		for i, j := range mapping {
			mapped[i] = union_args[j]
		}
		return mapped
	}
	var get_union_args = func(case_args []Type, mapping []uint) []Type {
		var mapped = make([]Type, len(mapping))
		for i := 0; i < len(mapped); i += 1 {
			mapped[i] = WildcardRhsType {}
		}
		for i, j := range mapping {
			mapped[j] = case_args[i]
		}
		return mapped
	}
	// -- behavior of assigning a named type to an union type --
	var assign_union func(NamedType, NamedType, Union) (Expr, *ExprError)
	assign_union = func(union_t NamedType, given NamedType, union Union) (Expr, *ExprError) {
		// 1. Find the given type in the list of case types of the union
		for index, case_type := range union.CaseTypes {
			if case_type.Name == given.Name {
				// 1.1. If found, check if type parameters matching
				var mapping = case_type.Params
				var case_expected_type = NamedType {
					Name: case_type.Name,
					Args: get_case_args(union_t.Args, mapping),
				}
				if DirectAssignableTo(case_expected_type, given, ctx) {
					// 1.1.1. If matching, return a lifted value.
					return Expr {
						Type:  NamedType {
							Name: union_t.Name,
							Args: get_union_args(given.Args, mapping),
						},
						Value: Sum { Value: expr, Index: uint(index) },
						Info:  expr.Info,
					}, nil
				} else {
					// 1.1.2. Otherwise, throw an error.
					return failed()
				}
			}
		}
		for index, case_type := range union.CaseTypes {
			var g = ctx.ModuleInfo.Types[case_type.Name]
			var item_union, item_is_union = g.Value.(Union)
			if item_is_union {
				var mapping = case_type.Params
				var item_expr, err = assign_union(NamedType {
					Name: case_type.Name,
					Args: get_case_args(union_t.Args, mapping),
				}, given, item_union)
				if err != nil {
					continue  // if `err` is thrown at 1.1.2, return is OK,
					          // but `err` could be thrown at 1.2, so...
				}
				return Expr {
					Type:  NamedType {
						Name: union_t.Name,
						Args: get_union_args(given.Args, mapping),
					},
					Value: Sum { Value: item_expr, Index: uint(index) },
					Info:  expr.Info,
				}, nil
			}
		}
		// 1.2. Otherwise, throw an error.
		return failed()
	}
	// 1. If the expected type is not specified,
	//    no further process is required.
	if expected == nil {
		return expr, nil
	}
	// Try Direct
	if DirectAssignableTo(expected, expr.Type, ctx) {
		return expr, nil
	}
	// Try Unbox
	var ctx_mod = ctx.ModuleInfo.Module.Name
	var reg = ctx.ModuleInfo.Types
	var unboxed, can_unbox = Unbox(expr.Type, ctx_mod, reg).(Unboxed)
	if can_unbox {
		var expr_with_unboxed = Expr {
			Type:  unboxed.Type,
			Value: expr.Value,
			Info:  expr.Info,
		}
		var expr_with_expected, err = AssignTypedTo (
			expected, expr_with_unboxed, ctx,
		)
		// if err != nil { return Expr{}, err }
		// if err != nil { return failed() }
		if err == nil {
			return expr_with_expected, nil
		}  // else fallthrough to union
	}
	// Try Union
	switch E := expected.(type) {
	case NamedType:
		switch T := expr.Type.(type) {
		case NamedType:
			if E.Name != T.Name {
				var g = ctx.ModuleInfo.Types[E.Name]
				var union, is_union = g.Value.(Union)
				if is_union {
					return assign_union(E, T, union)
				}
			}
		}
	}
	// Failed
	return failed()
}

func DirectAssignableTo(expected Type, given Type, ctx ExprContext) bool {
	switch E := expected.(type) {
	case ParameterType:
		if E.BeingInferred {
			if !(ctx.InferTypeArgs) { panic("something went wrong") }
			switch given.(type) {
			case WildcardRhsType:
				return true
			default:
				var inferred, exists = ctx.Inferred[E.Index]
				if exists {
					return AreTypesEqualInSameCtx(inferred, given)
				} else {
					ctx.Inferred[E.Index] = given
					return true
				}
			}
		} else {
			switch T := given.(type) {
			case WildcardRhsType:
				return true
			case ParameterType:
				return E.Index == T.Index
			default:
				return false
			}
		}
	}
	switch given.(type) {
	case WildcardRhsType:
		return true
	}
	switch E := expected.(type) {
	case NamedType:
		switch T := given.(type) {
		case NamedType:
			if E.Name == T.Name {
				if len(T.Args) != len(E.Args) {
					panic("something went wrong")
				}
				var L = len(T.Args)
				for i := 0; i < L; i += 1 {
					if !(DirectAssignableTo(E.Args[i], T.Args[i], ctx)) {
						return false
					}
				}
				return true
			} else {
				return false
			}
		}
	case AnonymousType:
		switch T := given.(type) {
		case AnonymousType:
			switch E_ := E.Repr.(type) {
			case Unit:
				switch T.Repr.(type) {
				case Unit:
					return true
				}
			case Tuple:
				switch T_ := T.Repr.(type) {
				case Tuple:
					if len(T_.Elements) != len(E_.Elements) {
						return false
					}
					var L = len(T_.Elements)
					for i := 0; i < L; i += 1 {
						var e = E_.Elements[i]
						var t = T_.Elements[i]
						if !(DirectAssignableTo(e, t, ctx)) {
							return false
						}
					}
					return true
				}
			case Bundle:
				switch T_ := T.Repr.(type) {
				case Bundle:
					if len(E_.Fields) != len(T_.Fields) {
						return false
					}
					for name, e := range E_.Fields {
						var t, exists = T_.Fields[name]
						if !exists {
							return false
						}
						if !(DirectAssignableTo(e.Type, t.Type, ctx)) {
							return false
						}
					}
					return true
				}
			case Func:
				switch T_ := T.Repr.(type) {
				case Func:
					if !(DirectAssignableTo(E_.Input, T_.Input, ctx)) {
						return false
					}
					if !(DirectAssignableTo(E_.Output, T_.Output, ctx)) {
						return false
					}
					return true
				}
			}
		}
	}
	return false
}
