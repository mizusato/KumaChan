package checker

import . "kumachan/util/error"


func GenericFunctionCall (
	f          *GenericFunction,
	name       string,
	index      uint,
	type_args  [] Type,
	arg        SemiExpr,
	f_info     ExprInfo,
	call_info  ExprInfo,
	ctx        ExprContext,
) (Expr, *ExprError) {
	var type_arity = len(f.TypeParams)
	var f_node = f_info.ErrorPoint.Node
	if len(type_args) == type_arity {
		var f_raw_type = &AnonymousType { f.DeclaredType }
		var f_type = FillTypeArgs(f_raw_type, type_args)
		var f_type_repr = f_type.(*AnonymousType).Repr.(Func)
		var input_type = f_type_repr.Input
		var output_type = f_type_repr.Output
		arg_typed, err := AssignTo(input_type, arg, ctx)
		if err != nil { return Expr{}, err }
		f_ref, err := MakeRefFunction(name, index, type_args, f_node, ctx)
		if err != nil { return Expr{}, err }
		var call = Expr {
			Type:  output_type,
			Value: Call {
				Function: Expr {
					Type:  f_type,
					Value: f_ref,
					Info:  f_info,
				},
				Argument: arg_typed,
			},
			Info:  call_info,
		}
		return call, nil
	} else if len(type_args) == 0 {
		var inf_ctx = ctx.WithInferringEnabled(f.TypeParams)
		var raw_input_type = f.DeclaredType.Input
		var raw_output_type = f.DeclaredType.Output
		var marked_input_type = MarkParamsAsBeingInferred(raw_input_type)
		// var marked_output_type = MarkParamsAsBeingInferred(raw_output_type)
		var arg_typed, err = AssignTo(marked_input_type, arg, inf_ctx)
		if err != nil { return Expr{}, err }
		var output_v = GetVariance(raw_output_type, TypeVarianceContext {
			Parameters: f.TypeParams,
			Registry:   ctx.ModuleInfo.Types,
		})
		var inferred_args = make([] Type, type_arity)
		for i := 0; i < type_arity; i += 1 {
			var t, exists = inf_ctx.Inferring.Arguments[uint(i)]
			if exists {
				inferred_args[i] = t
			} else {
				if output_v[i] == Covariant {
					inferred_args[i] = &NeverType {}
				} else if output_v[i] == Contravariant {
					inferred_args[i] = &AnyType {}
				} else {
					return Expr{}, &ExprError {
						Point:    f_info.ErrorPoint,
						Concrete: E_ExplicitTypeParamsRequired {},
					}
				}
			}
		}
		var input_type = FillTypeArgs(raw_input_type, inferred_args)
		var _, ok = AssignType(input_type, arg_typed.Type, Matching, ctx)
		if !(ok) {
			// var inf_ctx = ctx.WithInferringEnabled(f.TypeParams)
			// var _, _ = AssignTo(marked_input_type, arg, inf_ctx)
			panic("type system internal error (likely a bug)")
		}
		var output_type = FillTypeArgs(raw_output_type, inferred_args)
		var f_type = &AnonymousType { Func {
			Input:  input_type,
			Output: output_type,
		} }
		f_ref, err := MakeRefFunction(name, index, inferred_args, f_node, ctx)
		if err != nil { return Expr{}, err }
		var call = Expr {
			Type:  output_type,
			Value: Call {
				Function: Expr {
					Type:  f_type,
					Value: f_ref,
					Info:  f_info,
				},
				Argument: arg_typed,
			},
			Info:  call_info,
		}
		return call, nil
	} else {
		return Expr{}, &ExprError {
			Point:    f_info.ErrorPoint,
			Concrete: E_FunctionWrongTypeParamsQuantity {
				FuncName: name,
				Given:    uint(len(type_args)),
				Required: uint(type_arity),
			},
		}
	}
}

func GenericFunctionAssignTo (
	expected   Type,
	name       string,
	index      uint,
	f          *GenericFunction,
	type_args  [] Type,
	info       ExprInfo,
	ctx        ExprContext,
) (Expr, *ExprError) {
	var unit_t = &AnonymousType { Unit {} }
	if AreTypesEqualInSameCtx(f.DeclaredType.Input, unit_t) {
		// globally defined constant-like thunk
		var arg = LiftTyped(Expr {
			Type:  unit_t,
			Value: UnitValue {},
			Info:  info,
		})
		var expr, err = GenericFunctionCall (
			f, name, index, type_args, arg, info, info, ctx,
		)
		if err == nil {
			var assigned, err = TypedAssignTo(expected, expr, ctx)
			if err == nil {
				return assigned, nil
			}
		}
	}
	var type_arity = len(f.TypeParams)
	var f_node = info.ErrorPoint.Node
	if len(type_args) == type_arity {
		var f_raw_type = &AnonymousType { f.DeclaredType }
		var f_type = FillTypeArgs(f_raw_type, type_args)
		f_ref, err := MakeRefFunction(name, index, type_args, f_node, ctx)
		if err != nil { return Expr{}, err }
		var f_expr = Expr {
			Type:  f_type,
			Value: f_ref,
			Info:  info,
		}
		return TypedAssignTo(expected, f_expr, ctx)
	} else if len(type_args) == 0 {
		if expected == nil {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_ExplicitTypeRequired {},
			}
		}
		var exp_certain, err = GetCertainType(expected, info.ErrorPoint, ctx)
		if err != nil { return Expr{}, err }
		var inf_ctx = ctx.WithInferringEnabled(f.TypeParams)
		var f_raw_type = &AnonymousType { f.DeclaredType }
		var f_marked_type = MarkParamsAsBeingInferred(f_raw_type)
		var _, ok = AssignType(f_marked_type, exp_certain, FromInferred, inf_ctx)
		if !(ok) { return Expr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_NotAssignable {
				From:   inf_ctx.DescribeInferredType(f_marked_type),
				To:     ctx.DescribeCertainType(exp_certain),
				Reason: "",
			},
		} }
		if len(inf_ctx.Inferring.Arguments) != type_arity {
			panic("something went wrong")
		}
		var inferred_args = make([] Type, type_arity)
		for i := 0; i < type_arity; i += 1 {
			inferred_args[i] = inf_ctx.Inferring.Arguments[uint(i)]
		}
		var f_type = FillTypeArgs(f_raw_type, inferred_args)
		_, ok = AssignType(exp_certain, f_type, Matching, ctx)
		if !(ok) {
			panic("something went wrong")
		}
		f_ref, err := MakeRefFunction(name, index, inferred_args, f_node, ctx)
		if err != nil { return Expr{}, err }
		return Expr {
			Type:  f_type,
			Value: f_ref,
			Info:  info,
		}, nil
	} else {
		return Expr{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_FunctionWrongTypeParamsQuantity {
				FuncName: name,
				Given:    uint(len(type_args)),
				Required: uint(type_arity),
			},
		}
	}
}


// TODO: simplify common patterns of usage of this function
func FillTypeArgsWithDefaults(t Type, given_args ([] Type), defaults (map[uint] Type)) Type {
	switch T := t.(type) {
	case *NeverType:
		return &NeverType {}
	case *AnyType:
		return &AnyType {}
	case *ParameterType:
		if T.Index < uint(len(given_args)) {
			return given_args[T.Index]
		} else {
			if defaults != nil {
				var t, exists = defaults[T.Index]
				if exists { return t }
			}
			panic("something went wrong")
		}
	case *NamedType:
		var filled = make([]Type, len(T.Args))
		for i, arg := range T.Args {
			filled[i] = FillTypeArgsWithDefaults(arg, given_args, defaults)
		}
		return &NamedType {
			Name: T.Name,
			Args: filled,
		}
	case *AnonymousType:
		switch r := T.Repr.(type) {
		case Unit:
			return &AnonymousType { Unit {} }
		case Tuple:
			var filled = make([]Type, len(r.Elements))
			for i, element := range r.Elements {
				filled[i] = FillTypeArgsWithDefaults(element, given_args, defaults)
			}
			return &AnonymousType {
				Repr: Tuple {
					Elements: filled,
				},
			}
		case Bundle:
			var filled = make(map[string]Field, len(r.Fields))
			for name, field := range r.Fields {
				filled[name] = Field {
					Type:  FillTypeArgsWithDefaults(field.Type, given_args, defaults),
					Index: field.Index,
				}
			}
			return &AnonymousType {
				Repr: Bundle {
					Fields: filled,
				},
			}
		case Func:
			return &AnonymousType {
				Repr:Func {
					Input:  FillTypeArgsWithDefaults(r.Input, given_args, defaults),
					Output: FillTypeArgsWithDefaults(r.Output, given_args, defaults),
				},
			}
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

func FillTypeArgs(t Type, given_args ([] Type)) Type {
	return FillTypeArgsWithDefaults(t, given_args, nil)
}

func MarkParamsAsBeingInferred(type_ Type) Type {
	switch t := type_.(type) {
	case *NeverType:
		return &NeverType {}
	case *AnyType:
		return &AnyType {}
	case *ParameterType:
		return &ParameterType {
			Index:         t.Index,
			BeingInferred: true,
		}
	case *NamedType:
		var marked_args = make([]Type, len(t.Args))
		for i, arg := range t.Args {
			marked_args[i] = MarkParamsAsBeingInferred(arg)
		}
		return &NamedType {
			Name: t.Name,
			Args: marked_args,
		}
	case *AnonymousType:
		switch r := t.Repr.(type) {
		case Unit:
			return &AnonymousType { Unit{} }
		case Tuple:
			var marked_elements = make([]Type, len(r.Elements))
			for i, el := range r.Elements {
				marked_elements[i] = MarkParamsAsBeingInferred(el)
			}
			return &AnonymousType { Tuple { marked_elements } }
		case Bundle:
			var marked_fields = make(map[string]Field)
			for name, f := range r.Fields {
				marked_fields[name] = Field {
					Type:  MarkParamsAsBeingInferred(f.Type),
					Index: f.Index,
				}
			}
			return &AnonymousType { Bundle { marked_fields } }
		case Func:
			var marked_input = MarkParamsAsBeingInferred(r.Input)
			var marked_output = MarkParamsAsBeingInferred(r.Output)
			return &AnonymousType { Func {
				Input:  marked_input,
				Output: marked_output,
			} }
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

func GetCertainType(type_ Type, point ErrorPoint, ctx ExprContext) (Type, *ExprError) {
	if !(ctx.Inferring.Enabled) {
		return type_, nil
	}
	switch T := type_.(type) {
	case *NeverType:
		return &NeverType {}, nil
	case *AnyType:
		return &AnyType {}, nil
	case *ParameterType:
		if T.BeingInferred {
			var inferred, exists = ctx.Inferring.Arguments[T.Index]
			if exists {
				return inferred, nil
			} else {
				return nil, &ExprError {
					Point:    point,
					Concrete: E_ExplicitTypeRequired {},
				}
			}
		} else {
			return T, nil
		}
	case *NamedType:
		var result_args = make([]Type, len(T.Args))
		for i, arg := range T.Args {
			var t, err = GetCertainType(arg, point, ctx)
			if err != nil { return nil, err }
			result_args[i] = t
		}
		return &NamedType {
			Name: T.Name,
			Args: result_args,
		}, nil
	case *AnonymousType:
		switch r := T.Repr.(type) {
		case Unit:
			return &AnonymousType { Unit {} }, nil
		case Tuple:
			var result_elements = make([]Type, len(r.Elements))
			for i, element := range r.Elements {
				var t, err = GetCertainType(element, point, ctx)
				if err != nil { return nil, err }
				result_elements[i] = t
			}
			return &AnonymousType {
				Repr: Tuple {
					Elements: result_elements,
				},
			}, nil
		case Bundle:
			var result_fields = make(map[string]Field, len(r.Fields))
			for name, field := range r.Fields {
				var t, err = GetCertainType(field.Type, point, ctx)
				if err != nil { return nil, err }
				result_fields[name] = Field {
					Type:  t,
					Index: field.Index,
				}
			}
			return &AnonymousType {
				Repr: Bundle {
					Fields: result_fields,
				},
			}, nil
		case Func:
			var input, err1 = GetCertainType(r.Input, point, ctx)
			if err1 != nil { return nil, err1 }
			var output, err2 = GetCertainType(r.Output, point, ctx)
			if err2 != nil { return nil, err2 }
			return &AnonymousType {
				Repr:Func {
					Input:  input,
					Output: output,
				},
			}, nil
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}

