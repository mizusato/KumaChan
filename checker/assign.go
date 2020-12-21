package checker

import (
	. "kumachan/error"
	"kumachan/parser/ast"
)


type AssignDirection int
const (
	ToInferred AssignDirection = iota
	FromInferred
	Matching
)

func DirectionFromVariance(v TypeVariance) AssignDirection {
	switch v {
	case Covariant:
		return ToInferred
	case Contravariant:
		return FromInferred
	case Invariant:
		return Matching
	case Bivariant:
		panic("bivariant is not declarable")
	default:
		panic("impossible branch")
	}
}

func InverseDirection(d AssignDirection) AssignDirection {
	switch d {
	case ToInferred:
		return FromInferred
	case FromInferred:
		return ToInferred
	default:
		return d
	}
}


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

func AssignAstExprTo(expected Type, raw ast.Expr, ctx ExprContext) (Expr, *ExprError) {
	semi, err := Check(raw, ctx)
	if err != nil { return Expr{}, err }
	expr, err := AssignTo(expected, semi, ctx)
	if err != nil { return Expr{}, err }
	return expr, nil
}

func AssignTo(expected Type, semi SemiExpr, ctx ExprContext) (Expr, *ExprError) {
	var expr, err = (func() (Expr, *ExprError) {
		switch semi_value := semi.Value.(type) {
		case TypedExpr:
			return TypedAssignTo(expected, Expr(semi_value), ctx)
		case UndecidedCall:
			return AssignUndecidedTo(expected, semi_value, semi.Info, ctx)
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
			Type:  UnboxWeak(expr.Type, ctx.ModuleInfo.Types),
			Value: expr.Value,
			Info:  expr.Info,
		}, nil
	}
	var result_type, ok = AssignType(expected, expr.Type, ToInferred, ctx)
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

func AssignType(inferred Type, given Type, d AssignDirection, ctx ExprContext) (Type, bool) {
	var direct_type, ok = DirectAssignType(inferred, given, d, ctx)
	if ok {
		return direct_type, true
	} else {
		var ctx_mod = ctx.GetModuleName()
		var reg = ctx.GetTypeRegistry()
		switch d {
		case ToInferred:
			var _, is_given_wildcard = given.(*NeverType)
			if is_given_wildcard {
				return AssignWildcardRhsTypeTo(inferred, ctx)
			}
			var unboxed, can_unbox = Unbox(given, ctx_mod, reg).(Unboxed)
			if can_unbox {
				return AssignType(inferred, unboxed.Type, d, ctx)
			} else {
				return nil, false
			}
		case FromInferred:
			var _, is_expected_wildcard = inferred.(*NeverType)
			if is_expected_wildcard {
				return &NeverType {}, true
			}
			var unboxed, can_unbox = Unbox(inferred, ctx_mod, reg).(Unboxed)
			if can_unbox {
				return AssignType(unboxed.Type, given, d, ctx)
			} else {
				return nil, false
			}
		case Matching:
			return nil, false
		default:
			panic("impossible branch")
		}
	}
}
func AssignWildcardRhsTypeTo(expected Type, ctx ExprContext) (Type, bool) {
	var t, err = GetCertainType(expected, ErrorPoint{}, ctx)
	if err != nil { return nil, false }
	return t, true
}

func DirectAssignType(inferred Type, given Type, d AssignDirection, ctx ExprContext) (Type, bool) {
	var given_param, given_is_param = given.(*ParameterType)
	if given_is_param {
		if d == ToInferred {
			var super, has_super = ctx.TypeBounds.Super[given_param.Index]
			if has_super {
				return AssignType(inferred, super, d, ctx)
			}
		} else if d == FromInferred {
			var sub, has_sub = ctx.TypeBounds.Sub[given_param.Index]
			if has_sub {
				return AssignType(inferred, sub, d, ctx)
			}
		}
	}
	switch I := inferred.(type) {
	case *NeverType:
		var _, given_is_also_this = given.(*NeverType)
		if given_is_also_this {
			return given, true
		}
	case *ParameterType:
		if I.BeingInferred {
			if !(ctx.Inferring.Enabled) { panic("something went wrong") }
			var inferred_t, exists = ctx.Inferring.Arguments[I.Index]
			if exists {
				return DirectAssignType(inferred_t, given, d, ctx)
			} else {
				ctx.Inferring.Arguments[I.Index] = given
				return given, true
			}
		} else {
			switch T := given.(type) {
			case *ParameterType:
				if I.Index == T.Index {
					return given, true
				}
			default:
				if d == ToInferred {
					var sub, has_sub = ctx.TypeBounds.Sub[I.Index]
					if has_sub {
						return AssignType(sub, given, d, ctx)
					}
				} else if d == FromInferred {
					var super, has_super = ctx.TypeBounds.Super[I.Index]
					if has_super {
						return AssignType(super, given, d, ctx)
					}
				}
				return nil, false
			}
		}
	case *NamedType:
		switch T := given.(type) {
		case *NamedType:
			if I.Name == T.Name {
				if len(T.Args) != len(I.Args) {
					panic("something went wrong")
				}
				var name = T.Name
				var L = len(T.Args)
				var g = ctx.ModuleInfo.Types[name]
				var args = make([] Type, L)
				for i := 0; i < L; i += 1 {
					var param_v = g.Params[i].Variance
					var param_d = DirectionFromVariance(param_v)
					if d == FromInferred {
						param_d = InverseDirection(param_d)
					}
					var t, ok = AssignType(I.Args[i], T.Args[i], param_d, ctx)
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
			switch I_ := I.Repr.(type) {
			case Unit:
				switch T.Repr.(type) {
				case Unit:
					return given, true
				}
			case Tuple:
				switch T_ := T.Repr.(type) {
				case Tuple:
					if len(T_.Elements) != len(I_.Elements) {
						return nil, false
					}
					var L = len(T_.Elements)
					var elements = make([] Type, L)
					for i := 0; i < L; i += 1 {
						var e = I_.Elements[i]
						var t = T_.Elements[i]
						var el_t, ok = AssignType(e, t, d, ctx)
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
					if len(I_.Fields) != len(T_.Fields) {
						return nil, false
					}
					var fields = make(map[string] Field)
					for name, e := range I_.Fields {
						var t, exists = T_.Fields[name]
						if !exists {
							return nil, false
						}
						if t.Index != e.Index {
							return nil, false
						}
						var field_index = t.Index
						var field_t, ok = AssignType(e.Type, t.Type, d, ctx)
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
					var input_t, ok1 = AssignType(
						I_.Input, T_.Input, InverseDirection(d), ctx,
					)
					if !(ok1) {
						return nil, false
					}
					var output_t, ok2 = AssignType(
						I_.Output, T_.Output, d, ctx,
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

