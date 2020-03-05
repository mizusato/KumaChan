package checker


func RequireExplicitType(t Type, info ExprInfo) (struct{}, *ExprError) {
	if t == nil {
		return struct{}{}, &ExprError {
			Point:    info.ErrorPoint,
			Concrete: E_ExplicitTypeRequired {},
		}
	} else {
		return struct{}{}, nil
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
	default:
		panic("impossible branch")
	}
}

// TODO: Add types: TypedAssignContext, TypeInferringContext

func AssignTypedTo(expected Type, expr Expr, ctx ExprContext, unbox bool) (Expr, *ExprError) {
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
		var assign_union = func(exp NamedType, given NamedType, union Union) (Expr, *ExprError) {
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
			var result, err = AssignTypedTo(expected, expr_with_inner, ctx, false)
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
					var item, err = AssignTypedTo(item_exp_type, raw, ctx)
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
						var field, err = AssignTypedTo(field_exp_type, raw, ctx)
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
			if unbox {
				switch tv := given_g.Value.(type) {
				case Wrapped:
					var given_inner = FillArgs(tv.InnerType, G.Args)
					var given_mod = G.Name.ModuleName
					var given_opaque = tv.Opaque
					// 3.2.1.2. Otherwise, if the given type has an
					//          inner type, try to unbox the inner type.
					return assign_inner(given_inner, given_opaque, given_mod)
				}
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
