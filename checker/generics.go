package checker

import . "kumachan/error"


func GenericFunctionCall (
	f            *GenericFunction,
	name         string,
	index        uint,
	type_args    [] Type,
	arg          SemiExpr,
	f_info       ExprInfo,
	call_info    ExprInfo,
	ctx          ExprContext,
	is_exact     *bool,  // mutate it instead of additional return value
) (Expr, *ExprError) {
	var type_arity = len(f.TypeParams)
	if len(type_args) == type_arity {
		var f_raw_type = AnonymousType { f.DeclaredType }
		var f_type = FillTypeArgs(f_raw_type, type_args)
		var f_type_repr = f_type.(AnonymousType).Repr.(Func)
		var input_type = f_type_repr.Input
		var output_type = f_type_repr.Output
		var arg_typed, err = AssignTo(input_type, arg, ctx)
		if err != nil { return Expr{}, err }
		if IsExactAssignTo(arg_typed.Type, arg) {
			*is_exact = true
		}
		return Expr {
			Type:  output_type,
			Value: Call {
				Function: Expr {
					Type:  f_type,
					Value: MakeRefFunction(name, index, ctx),
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
		var marked_input_type = MarkParamsAsBeingInferred(raw_input_type)
		var arg_typed, err = AssignTo(marked_input_type, arg, inf_ctx)
		if err != nil { return Expr{}, err }
		var inferred_args = make([]Type, type_arity)
		for i := 0; i < type_arity; i += 1 {
			var t, exists = inf_ctx.Inferred[uint(i)]
			if exists {
				inferred_args[i] = t
			} else {
				inferred_args[i] = WildcardRhsType {}
			}
		}
		var input_type = FillTypeArgs(raw_input_type, inferred_args)
		var _, ok = DirectAssignTo(input_type, arg_typed.Type, ctx)
		if !(ok) {
			// var inf_ctx = ctx.WithTypeArgsInferringEnabled(f.TypeParams)
			// var _, _ = AssignTo(marked_input_type, arg, inf_ctx)
			panic("something went wrong")
		}
		var output_type = FillTypeArgs(raw_output_type, inferred_args)
		var f_type = AnonymousType { Func {
			Input:  input_type,
			Output: output_type,
		} }
		if IsExactAssignTo(arg_typed.Type, arg) {
			*is_exact = true
		}
		return Expr {
			Type:  output_type,
			Value: Call {
				Function: Expr {
					Type:  f_type,
					Value: MakeRefFunction(name, index, ctx),
					Info:  f_info,
				},
				Argument: arg_typed,
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
			Value: MakeRefFunction(name, index, ctx),
			Info:  info,
		}
		return AssignTypedTo(expected, f_expr, ctx)
	} else if len(type_args) == 0 {
		if expected == nil {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_ExplicitTypeRequired {},
			}
		}
		if ctx.InferTypeArgs {
			return Expr{}, &ExprError {
				Point:    info.ErrorPoint,
				Concrete: E_ExplicitTypeParamsRequired {},
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
				Value: MakeRefFunction(name, index, ctx),
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


func FillTypeArgs(t Type, given_args []Type) Type {
	switch T := t.(type) {
	case WildcardRhsType:
		return WildcardRhsType {}
	case ParameterType:
		return given_args[T.Index]
	case NamedType:
		var filled = make([]Type, len(T.Args))
		for i, arg := range T.Args {
			filled[i] = FillTypeArgs(arg, given_args)
		}
		return NamedType {
			Name: T.Name,
			Args: filled,
		}
	case AnonymousType:
		switch r := T.Repr.(type) {
		case Unit:
			return AnonymousType { Unit {} }
		case Tuple:
			var filled = make([]Type, len(r.Elements))
			for i, element := range r.Elements {
				filled[i] = FillTypeArgs(element, given_args)
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
					Type:  FillTypeArgs(field.Type, given_args),
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
					Input:  FillTypeArgs(r.Input, given_args),
					Output: FillTypeArgs(r.Output, given_args),
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
	case WildcardRhsType:
		return
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

func MarkParamsAsBeingInferred(type_ Type) Type {
	switch t := type_.(type) {
	case WildcardRhsType:
		return WildcardRhsType {}
	case ParameterType:
		return ParameterType {
			Index:         t.Index,
			BeingInferred: true,
		}
	case NamedType:
		var marked_args = make([]Type, len(t.Args))
		for i, arg := range t.Args {
			marked_args[i] = MarkParamsAsBeingInferred(arg)
		}
		return NamedType {
			Name: t.Name,
			Args: marked_args,
		}
	case AnonymousType:
		switch r := t.Repr.(type) {
		case Unit:
			return AnonymousType { Unit{} }
		case Tuple:
			var marked_elements = make([]Type, len(r.Elements))
			for i, el := range r.Elements {
				marked_elements[i] = MarkParamsAsBeingInferred(el)
			}
			return AnonymousType { Tuple { marked_elements } }
		case Bundle:
			var marked_fields = make(map[string]Field)
			for name, f := range r.Fields {
				marked_fields[name] = Field {
					Type:  MarkParamsAsBeingInferred(f.Type),
					Index: f.Index,
				}
			}
			return AnonymousType { Bundle { marked_fields } }
		case Func:
			var marked_input = MarkParamsAsBeingInferred(r.Input)
			var marked_output = MarkParamsAsBeingInferred(r.Output)
			return AnonymousType { Func {
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
	if !(ctx.InferTypeArgs) {
		return type_, nil
	}
	switch T := type_.(type) {
	case WildcardRhsType:
		return WildcardRhsType {}, nil
	case ParameterType:
		if T.BeingInferred {
			var inferred, exists = ctx.Inferred[T.Index]
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
	case NamedType:
		var result_args = make([]Type, len(T.Args))
		for i, arg := range T.Args {
			var t, err = GetCertainType(arg, point, ctx)
			if err != nil { return nil, err }
			result_args[i] = t
		}
		return NamedType {
			Name: T.Name,
			Args: result_args,
		}, nil
	case AnonymousType:
		switch r := T.Repr.(type) {
		case Unit:
			return AnonymousType { Unit {} }, nil
		case Tuple:
			var result_elements = make([]Type, len(r.Elements))
			for i, element := range r.Elements {
				var t, err = GetCertainType(element, point, ctx)
				if err != nil { return nil, err }
				result_elements[i] = t
			}
			return AnonymousType {
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
			return AnonymousType {
				Repr: Bundle {
					Fields: result_fields,
				},
			}, nil
		case Func:
			var input, err1 = GetCertainType(r.Input, point, ctx)
			if err1 != nil { return nil, err1 }
			var output, err2 = GetCertainType(r.Output, point, ctx)
			if err2 != nil { return nil, err2 }
			return AnonymousType {
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

