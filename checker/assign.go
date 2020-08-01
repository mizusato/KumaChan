package checker

import . "kumachan/error"


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
			return TypedAssignTo(expected, Expr(semi_value), ctx)
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
		default:
			panic("impossible branch")
		}
	})()
	if err != nil { return Expr{}, err }
	return expr, nil
}

func TypedAssignTo(expected Type, expr Expr, ctx ExprContext) (Expr, *ExprError) {
	if expected == nil {
		return Expr {
			Type:  UnboxAsIs(expr.Type, ctx.ModuleInfo.Types),
			Value: expr.Value,
			Info:  expr.Info,
		}, nil
	}
	var result_type, ok = AssignTypeTo(expected, expr.Type, Covariant, ctx)
	if ok {
		return Expr {
			Type:  result_type,
			Value: expr.Value,
			Info:  expr.Info,
		}, nil
	} else {
		return Expr{}, &ExprError {
			Point:    expr.Info.ErrorPoint,
			Concrete: E_NotAssignable {
				From:   ctx.DescribeType(expr.Type),
				To:     ctx.DescribeExpectedType(expected),
				Reason: "",  // reserved for possible use in the future
			},
		}
	}
}

func AssignTypeTo(expected Type, given Type, v TypeVariance, ctx ExprContext) (Type, bool) {
	var direct_type, ok = DirectAssignTypeTo(expected, given, v, ctx)
	if ok {
		return direct_type, true
	} else {
		var ctx_mod = ctx.ModuleInfo.Module.Name
		var reg = ctx.ModuleInfo.Types
		switch v {
		case Covariant:
			var _, is_given_wildcard = given.(*WildcardRhsType)
			if is_given_wildcard {
				return AssignWildcardRhsTypeTo(expected, ctx)
			}
			var unboxed, can_unbox = Unbox(given, ctx_mod, reg).(Unboxed)
			if can_unbox {
				return AssignTypeTo(expected, unboxed.Type, v, ctx)
			} else {
				return nil, false
			}
		case Contravariant:
			var _, is_expected_wildcard = expected.(*WildcardRhsType)
			if is_expected_wildcard {
				return &WildcardRhsType {}, true
			}
			var unboxed, can_unbox = Unbox(expected, ctx_mod, reg).(Unboxed)
			if can_unbox {
				return AssignTypeTo(unboxed.Type, given, v, ctx)
			} else {
				return nil, false
			}
		default:
			return nil, false
		}
	}
}

func AssignWildcardRhsTypeTo(expected Type, ctx ExprContext) (Type, bool) {
	var t, err = GetCertainType(expected, ErrorPoint{}, ctx)
	if err != nil { return nil, false }
	return t, true
}

func DirectAssignTypeTo(expected Type, given Type, v TypeVariance, ctx ExprContext) (Type, bool) {
	var given_param, given_is_param = given.(*ParameterType)
	if given_is_param {
		var super, has_super = ctx.TypeBounds.Super[given_param.Index]
		if has_super {
			return AssignTypeTo(expected, super, Covariant, ctx)
		}
	}
	switch E := expected.(type) {
	case *ParameterType:
		if E.BeingInferred {
			if !(ctx.Inferring.Enabled) { panic("something went wrong") }
			var inferred, exists = ctx.Inferring.Arguments[E.Index]
			if exists {
				if AreTypesEqualInSameCtx(inferred, given) {
					return given, true
				} else {
					return nil, false
				}
			} else {
				ctx.Inferring.Arguments[E.Index] = given
				return given, true
			}
		} else {
			switch T := given.(type) {
			case *ParameterType:
				if E.Index == T.Index {
					return given, true
				}
			default:
				var sub, has_sub = ctx.TypeBounds.Sub[E.Index]
				if has_sub {
					return AssignTypeTo(sub, given, Covariant, ctx)
				} else {
					return nil, false
				}
			}
		}
	case *NamedType:
		switch T := given.(type) {
		case *NamedType:
			if E.Name == T.Name {
				if len(T.Args) != len(E.Args) {
					panic("something went wrong")
				}
				var g = ctx.ModuleInfo.Types[E.Name]
				var L = len(T.Args)
				var name = T.Name
				var args = make([] Type, L)
				for i := 0; i < L; i += 1 {
					var v = g.Params[i].Variance
					var t, ok = AssignTypeTo(E.Args[i], T.Args[i], v, ctx)
					if ok {
						args[i] = t
					} else {
						return nil, false
					}
				}
				return &NamedType {
					Name: name,
					Args: args,
				}, true
			} else {
				return nil, false
			}
		}
	case *AnonymousType:
		switch T := given.(type) {
		case *AnonymousType:
			switch E_ := E.Repr.(type) {
			case Unit:
				switch T.Repr.(type) {
				case Unit:
					return given, true
				}
			case Tuple:
				switch T_ := T.Repr.(type) {
				case Tuple:
					if len(T_.Elements) != len(E_.Elements) {
						return nil, false
					}
					var L = len(T_.Elements)
					var elements = make([] Type, L)
					for i := 0; i < L; i += 1 {
						var e = E_.Elements[i]
						var t = T_.Elements[i]
						var el_t, ok = AssignTypeTo(e, t, v, ctx)
						if ok {
							elements[i] = el_t
						} else {
							return nil, false
						}
					}
					return &AnonymousType { Tuple { elements } }, true
				}
			case Bundle:
				switch T_ := T.Repr.(type) {
				case Bundle:
					if len(E_.Fields) != len(T_.Fields) {
						return nil, false
					}
					var fields = make(map[string] Field)
					for name, e := range E_.Fields {
						var t, exists = T_.Fields[name]
						if !exists {
							return nil, false
						}
						if t.Index != e.Index {
							return nil, false
						}
						var field_index = t.Index
						var field_t, ok = AssignTypeTo (
							e.Type, t.Type, v, ctx,
						)
						if ok {
							fields[name] = Field {
								Type:  field_t,
								Index: field_index,
							}
						} else {
							return nil, false
						}
					}
					return &AnonymousType { Bundle { fields } }, true
				}
			case Func:
				switch T_ := T.Repr.(type) {
				case Func:
					var input_t, ok1 = AssignTypeTo (
						E_.Input, T_.Input, InverseVariance(v), ctx,
					)
					if !(ok1) {
						return nil, false
					}
					var output_t, ok2 = AssignTypeTo (
						E_.Output, T_.Output, v, ctx,
					)
					if !(ok2) {
						return nil, false
					}
					return &AnonymousType { Func {
						Input:  input_t,
						Output: output_t,
					} }, true
				}
			}
		}
	}
	return nil, false
}
