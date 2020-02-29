package checker

func AssignSemiTo(expected Type, semi SemiExpr, ctx ExprContext) (Expr, *ExprError) {
	var throw = func(e ConcreteExprError) (Expr, *ExprError) {
		return Expr{}, &ExprError {
			Point:    semi.Info.ErrorPoint,
			Concrete: e,
		}
	}
	switch given_semi := semi.Value.(type) {
	case TypedExpr:
		return AssignTo(expected, Expr(given_semi), ctx)
	case UntypedLambda:
		if expected == nil {
			throw(E_ExplicitTypeRequired {})
		}
		switch E := expected.(type) {
		case AnonymousType:
			switch func_repr := E.Repr.(type) {
			case Func:
				var lambda = given_semi
				var input = func_repr.Input
				var output = func_repr.Output
				var inner_ctx, err1 = ctx.WithPatternMatching (
					input, lambda.Input, lambda.InputNode, false,
				)
				if err1 != nil { return Expr{}, err1 }
				var output_semi, err2 = SemiExprFrom(lambda.Output, inner_ctx)
				if err2 != nil { return Expr{}, err2 }
				var output_typed, err3 = AssignSemiTo(output, output_semi, ctx)
				if err3 != nil { return Expr{}, err3 }
				return Expr {
					Type:  expected,
					Info:  semi.Info,
					Value: Lambda {
						Input:  lambda.Input,
						Output: output_typed,
					},
				}, nil
			}
		}
		return throw(E_LambdaAssignedToNonFuncType {
			NonFuncType: ctx.DescribeType(expected),
		})
	case UntypedInteger:
		var integer = given_semi
		if expected == nil {
			return Expr {
				Type:  NamedType {
					Name: __Int,
					Args: make([]Type, 0),
				},
				Info:  semi.Info,
				Value: IntLiteral { integer.Value },
			}, nil
		}
		switch E := expected.(type) {
		case NamedType:
			var sym = E.Name
			var kind, exists = __IntegerTypeMap[sym]
			if exists {
				if len(E.Args) > 0 { panic("something went wrong") }
				var val, ok = AdaptInteger(kind, integer.Value)
				if ok {
					return Expr {
						Type:  expected,
						Info:  semi.Info,
						Value: val,
					}, nil
				} else {
					return throw(E_IntegerOverflow { kind })
				}
			}
		}
		return throw(E_IntegerAssignedToNonIntegerType {})
	case SemiTypedTuple:
		var tuple_semi = given_semi
		switch E := expected.(type) {
		case AnonymousType:
			switch tuple := E.Repr.(type) {
			case Tuple:
				var required = len(tuple.Elements)
				var given = len(tuple_semi.Values)
				if given != required {
					return throw(E_TupleSizeNotMatching {
						Required:  required,
						Given:     given,
						GivenType: ctx.DescribeType(AnonymousType { tuple }),
					})
				}
				var typed_exprs = make([]Expr, given)
				for i, el := range tuple_semi.Values {
					var el_expected = tuple.Elements[i]
					var typed, err = AssignSemiTo(el_expected, el, ctx)
					if err != nil { return Expr{}, err }
					typed_exprs[i] = typed
				}
				return Expr {
					Type:  expected,
					Info:  semi.Info,
					Value: Product { typed_exprs },
				}, nil
			}
		}
		return throw(E_TupleAssignedToNonTupleType {})
	case SemiTypedBundle:
		var bundle_semi = given_semi
		switch E := expected.(type) {
		case AnonymousType:
			switch bundle := E.Repr.(type) {
			case Bundle:
				var values = make([]Expr, len(bundle.Index))
				for field_name, index := range bundle.Index {
					var field_type = bundle.Fields[field_name]
					var given_index, exists = bundle_semi.Index[field_name]
					if !exists {
						return throw(E_MissingField {
							Field: field_name,
							Type:  ctx.DescribeType(field_type),
						})
					}
					var given_value = bundle_semi.Values[given_index]
					var value, err = AssignSemiTo(field_type, given_value, ctx)
					if err != nil { return Expr{}, err }
					values[index] = value
				}
				for given_field_name, index := range bundle_semi.Index {
					var _, exists = bundle.Fields[given_field_name]
					if !exists {
						var key_node = bundle_semi.KeyNodes[index]
						return Expr{}, &ExprError {
							Point:    ctx.GetErrorPoint(key_node),
							Concrete: E_SurplusField { given_field_name },
						}
					}
				}
				return Expr {
					Type:  expected,
					Info:  semi.Info,
					Value: Product { values },
				}, nil
			}
		}
	// TODO
	}
	// TODO
	return Expr{}, nil
}

func AssignTo(expected Type, expr Expr, ctx ExprContext) (Expr, *ExprError) {
	if expected == nil {
		// 1. If the expected type is not specified,
		//    no further process is required.
		return expr, nil
	} else if AreTypesEqualInSameCtx(expected, expr.Type) {
		// 2. If the expected type is identical to the given type,
		//    no further process is required.
		// TODO: derive the types of type arguments in ctx
		//       (remove this branch, process in the following else branch)
		return expr, nil
	} else {
		// 3. Otherwise, try some implicit type conversions
		// 3.1. Define some inner functions
		// -- shortcut to produce a "not assignable" error --
		var throw = func(reason string) *ExprError {
			return &ExprError {
				Point: expr.Info.ErrorPoint,
				Concrete: E_NotAssignable {
					From:   ctx.DescribeType(expr.Type),
					To:     ctx.DescribeType(expected),
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
			var types = ctx.ModuleInfo.Types
			for index, subtype := range union.SubTypes {
				var sub_g = types[subtype]
				switch sub_union := sub_g.Value.(type) {
				case Union:
					var sub_exp = NamedType {
						Name: subtype,
						Args: exp.Args,
					}
					var sub, err = assign_union(sub_exp, given, sub_union)
					if err != nil {
						continue
					} else {
						return Expr {
							Type:  exp,
							Value: Sum { Value: sub, Index: uint(index) },
							Info:  sub.Info,
						}, nil
					}
				default:
					continue
				}
			}
			// 1.2. Otherwise, throw an error.
			return Expr{}, throw("given type is not a subtype of the expected union type")
		}
		// -- behavior of unpacking wrapped inner types --
		var assign_inner = func(inner Type, opaque bool, mod string) (Expr, *ExprError) {
			// 1. Create an alternative expression with the inner type
			var expr_with_inner = Expr {
				Type:  inner,
				Value: expr.Value,
				Info:  expr.Info,
			}
			// 2. Try to assign the created expression to the expected type
			var result, err = AssignTo(expected, expr_with_inner, ctx)
			if err != nil {
				return Expr{}, throw("")
			}
			// 3. Check if the module encapsulation is violated
			var ctx_mod = ctx.GetModuleName()
			if opaque && ctx_mod != mod {
				return Expr{}, throw("cannot cast out of opaque type")
			} else {
				return result, nil
			}
		}
		/*
		// -- behavior of adapting tuple literals --
		var assign_tuple = func(exp Tuple, given Tuple) (Expr, *ExprError) {
			// 1. Check if quantities of elements are identical
			if len(given.Elements) != len(exp.Elements) {
				return Expr{}, throw("tuple arity not matching")
			}
			// 2. Try to adapt the given expression
			switch v := expr.Value.(type) {
			case Product:
				// 2.1. If the given expression is a tuple literal
				var L = len(given.Elements)
				var items = make([]Expr, L)
				for i := 0; i < L; i += 1 {
					var item_exp_type = exp.Elements[i]
					var raw = v.Values[i]
					// 2.1.1. Try to adapt each element of the literal
					//        to the expected item type
					var item, err = AssignTo(item_exp_type, raw, ctx)
					if err != nil {
						return Expr{}, err
					}
					items[i] = item
				}
				// 2.1.2. Collect all adapted elements as a new tuple.
				return Expr{
					Type:  expected,
					Value: Product { items },
					Info:  expr.Info,
				}, nil
			default:
				// 2.2. Otherwise, a non-literal value is not adaptable.
				return Expr{}, throw("non-literal tuple cannot be assigned to different tuple type")
			}
		}
		// -- behavior of adapting bundle literals --
		var assign_bundle = func(exp Bundle, given Bundle) (Expr, *ExprError) {
			switch v := expr.Value.(type) {
			case Product:
				// 1. If the given expression is a bundle literal
				for name, _ := range given.Index {
					var _, exists = exp.Index[name]
					if !exists {
						// 1.1. Check for surplus fields
						return Expr{}, throw("surplus field " + name)
					}
				}
				var values = make([]Expr, len(exp.Index))
				for name, index := range exp.Index {
					// 1.2. Adapt each expected field
					var field_exp_type = exp.Fields[name]
					var given_index, exists_in_given = given.Index[name]
					if exists_in_given {
						// 1.2.1. If an expected field exists in the given
						//        bundle literal, try to adapt the field
						//        to its expected type
						var raw = v.Values[given_index]
						var field, err = AssignTo(field_exp_type, raw, ctx)
						if err != nil {
							return Expr{}, err
						}
						values[index] = field
					} else {
						// 1.2.2. Otherwise, if an expected field is missing,
						//        throw an error.
						return Expr{}, throw("missing field " + name)
					}
				}
				// 1.3. Collect all adapted fields as a new bundle.
				return Expr {
					Type:  expected,
					Value: Product { values },
					Info:  expr.Info,
				}, nil
			default:
				// 2. Otherwise, a non-literal bundle is not adaptable.
				return Expr{}, throw("non-literal bundle cannot be assigned to different bundle type")
			}
		}*/
		// 3.2. Determine the conversion behavior according to type details
		switch G := expr.Type.(type) {
		case NamedType:
			// 3.2.1. If the given type is a named type
			var types = ctx.ModuleInfo.Types
			var given_g = types[G.Name]
			switch E := expected.(type) {
			case NamedType:
				var expected_g = types[E.Name]
				switch union := expected_g.Value.(type) {
				case Union:
					// 3.2.1.1. If the expected type is a union type,
					//          try to lift the given expression
					//          to the union type.
					return assign_union(E, G, union)
				}
			}
			switch tv := given_g.Value.(type) {
			case Wrapped:
				var given_inner = FillArgs(tv.InnerType, G.Args)
				var given_mod = G.Name.ModuleName
				var given_opaque = tv.Opaque
				// 3.2.1.2. Otherwise, if the given type has an inner type,
				//          try to unpack the inner type.
				return assign_inner(given_inner, given_opaque, given_mod)
			}
		case AnonymousType:
			// 3.2.2. If the given type is an anonymous type
			switch G.Repr.(type) {
			case Tuple:
				switch E := expected.(type) {
				case AnonymousType:
					switch E.Repr.(type) {
					case Tuple:
						return Expr{}, throw("non-literal tuple cannot be adapt to different tuple type")
					default:
					}
				}
			case Bundle:
				switch E := expected.(type) {
				case AnonymousType:
					switch E.Repr.(type) {
					case Bundle:
						return Expr{}, throw("non-literal bundle cannot be adapt to different bundle type")
					}
				}
			}
		}
		// 3.3. If no conversion is available, types are not compatible.
		return Expr{}, throw("")
	}
}

func LiftToMaxType(exprs []Expr, ctx ExprContext) ([]Expr, Type, bool) {
	var L = len(exprs)
	var result = make([]Expr, L)
	for i := 0; i < L; i += 1 {
		var expected = exprs[i].Type
		var ok = true
		for j := 0; j < L; j += 1 {
			var item, err = AssignTo(expected, exprs[j], ctx)
			if err != nil {
				ok = false
				break
			}
			result[j] = item
		}
		if ok {
			return result, expected, true
		}
	}
	return nil, nil, false
}