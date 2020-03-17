package checker


func GenericFunctionCall (
	f            *GenericFunction,
	name         string,
	index        uint,
	type_args    [] Type,
	arg          SemiExpr,
	f_info       ExprInfo,
	call_info    ExprInfo,
	ctx          ExprContext,
	unbox_count  *uint,  // mutate it instead of additional return value
) (Expr, *ExprError) {
	var type_arity = len(f.TypeParams)
	ctx = ctx.WithUnboxCounted(unbox_count)
	if len(type_args) == type_arity {
		var f_raw_type = AnonymousType { f.DeclaredType }
		var f_type = FillTypeArgs(f_raw_type, type_args)
		var f_type_repr = f_type.(AnonymousType).Repr.(Func)
		var input_type = f_type_repr.Input
		var output_type = f_type_repr.Output
		var arg_typed, err = AssignTo(input_type, arg, ctx)
		if err != nil { return Expr{}, err }
		return Expr {
			Type:  output_type,
			Value: Call {
				Function: Expr {
					Type:  f_type,
					Value: RefFunction {
						Name:  name,
						Index: index,
					},
					Info:  f_info,
				},
				Argument: arg_typed,
			},
			Info:  call_info,
		}, nil
	} else if len(type_args) == 0 {
		var inf_ctx = ctx.WithTypeArgsInferringEnabled(f.TypeParams)
		var raw_input_type = f.DeclaredType.Input
		var raw_output_type = f.DeclaredType.Output
		var arg_typed, err = AssignTo(raw_input_type, arg, inf_ctx)
		if err != nil { return Expr{}, err }
		if len(inf_ctx.Inferred) != type_arity {
			return Expr{}, &ExprError {
				Point:    f_info.ErrorPoint,
				Concrete: E_ExplicitTypeParamsRequired {},
			}
		}
		var inferred_args = make([]Type, type_arity)
		for i, t := range inf_ctx.Inferred {
			inferred_args[i] = t
		}
		var input_type = FillTypeArgs(raw_input_type, inferred_args)
		/*
		if !(AreTypesEqualInSameCtx(input_type, arg_typed.Type)) {
			panic("something went wrong ")
		}
		*/
		var input = Expr {
			Type:  input_type,
			Value: arg_typed.Value,
			Info:  arg_typed.Info,
		}
		var output_type = FillTypeArgs(raw_output_type, inferred_args)
		var f_type = AnonymousType { Func {
			Input:  input_type,
			Output: output_type,
		} }
		return Expr {
			Type:  output_type,
			Value: Call {
				Function: Expr {
					Type:  f_type,
					Value: RefFunction {
						Name:  name,
						Index: index,
					},
					Info:  f_info,
				},
				Argument: input,
			},
			Info:  call_info,
		}, nil
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
	type_args  []Type,
	info       ExprInfo,
	ctx        ExprContext,
) (Expr, *ExprError) {
	var type_arity = len(f.TypeParams)
	if len(type_args) == type_arity {
		var f_raw_type = AnonymousType { f.DeclaredType }
		var f_type = FillTypeArgs(f_raw_type, type_args)
		var f_expr = Expr {
			Type:  f_type,
			Value: RefFunction {
				Name:  name,
				Index: index,
			},
			Info:  info,
		}
		return AssignTypedTo(expected, f_expr, ctx, true)
	} else if len(type_args) == 0 {
		if expected == nil {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_ExplicitTypeRequired {},
			}
		}
		var f_raw_type = AnonymousType { f.DeclaredType }
		// Note: Unbox/Union related inferring is not required
		//       since function types are anonymous types and invariant.
		//       Just apply NaivelyInferTypeArgs() here.
		var inferred = make(map[uint]Type)
		NaivelyInferTypeArgs(f_raw_type, expected, inferred)
		if len(inferred) == type_arity {
			var args = make([]Type, type_arity)
			for i, t := range inferred {
				args[i] = t
			}
			var f_type = FillTypeArgs(f_raw_type, args)
			if !(AreTypesEqualInSameCtx(f_type, expected)) {
				panic("something went wrong")
			}
			return Expr {
				Type:  f_type,
				Value: RefFunction {
					Name:  name,
					Index: index,
				},
				Info:  info,
			}, nil
		} else {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_ExplicitTypeParamsRequired {},
			}
		}
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


func FillTypeArgs(t Type, given []Type) Type {
	switch T := t.(type) {
	case ParameterType:
		return given[T.Index]
	case NamedType:
		var filled = make([]Type, len(T.Args))
		for i, arg := range T.Args {
			filled[i] = FillTypeArgs(arg, given)
		}
		return NamedType {
			Name: T.Name,
			Args: filled,
		}
	case AnonymousType:
		switch r := T.Repr.(type) {
		case Unit:
			return t
		case Tuple:
			var filled = make([]Type, len(r.Elements))
			for i, element := range r.Elements {
				filled[i] = FillTypeArgs(element, given)
			}
			return AnonymousType {
				Repr: Tuple {
					Elements: filled,
				},
			}
		case Bundle:
			var filled = make(map[string]Field, len(r.Fields))
			for name, field := range r.Fields {
				filled[name] = Field {
					Type:  FillTypeArgs(field.Type, given),
					Index: field.Index,
				}
			}
			return AnonymousType {
				Repr: Bundle {
					Fields: filled,
				},
			}
		case Func:
			return AnonymousType {
				Repr:Func {
					Input:  FillTypeArgs(r.Input, given),
					Output: FillTypeArgs(r.Output, given),
				},
			}
		default:
			panic("impossible branch")
		}
	default:
		panic("impossible branch")
	}
}


func NaivelyInferTypeArgs(template Type, given Type, inferred map[uint]Type) {
	switch T := template.(type) {
	case ParameterType:
		var existing, exists = inferred[T.Index]
		if !exists || AreTypesEqualInSameCtx(existing, given) {
			inferred[T.Index] = given
		}
	case NamedType:
		switch G := given.(type) {
		case NamedType:
			var L1 = len(T.Args)
			var L2 = len(G.Args)
			if L1 != L2 { panic("type registration went wrong") }
			var L = L1
			for i := 0; i < L; i += 1 {
				NaivelyInferTypeArgs(T.Args[i], G.Args[i], inferred)
			}
		}
	case AnonymousType:
		switch G := given.(type) {
		case AnonymousType:
			switch Tr := T.Repr.(type) {
			case Tuple:
				switch Gr := G.Repr.(type) {
				case Tuple:
					var L1 = len(Tr.Elements)
					var L2 = len(Gr.Elements)
					if L1 == L2 {
						var L = L1
						for i := 0; i < L; i += 1 {
							NaivelyInferTypeArgs(Tr.Elements[i], Gr.Elements[i], inferred)
						}
					}
				}
			case Bundle:
				switch Gr := G.Repr.(type) {
				case Bundle:
					for name, Tf := range Tr.Fields {
						var Gf, exists = Gr.Fields[name]
						if exists {
							NaivelyInferTypeArgs(Tf.Type, Gf.Type, inferred)
						}
					}
				}
			case Func:
				switch Gr := G.Repr.(type) {
				case Func:
					NaivelyInferTypeArgs(Tr.Input, Gr.Input, inferred)
					NaivelyInferTypeArgs(Tr.Output, Gr.Output, inferred)
				}
			default:
				panic("impossible branch")
			}
		}
	default:
		panic("impossible branch")
	}
}


