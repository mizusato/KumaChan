package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
	. "kumachan/error"
)

type ExprContext struct {
	TypeCtx  TypeContext
	ValMap   map[loader.Symbol] Type
	FunMap   FunctionCollection
}

func ExprFrom (e node.Expr, ctx ExprContext, expected Type) (Expr, *ExprError) {
	// TODO
	return Expr{}, nil
}

func ExprFromPipe (p node.Pipe, ctx ExprContext, input Type) (Expr, *ExprError) {
	// TODO
	// if input == nil { ...
	return Expr{}, nil
}

func ExprFromTerm (t node.VariousTerm, ctx ExprContext, expected Type) (Expr, *ExprError) {
	var T Type
	var v ExprVal
	switch term := t.Term.(type) {
	case node.Tuple:
		var L = len(term.Elements)
		if L == 0 {
			T = AnonymousType { Unit {} }
			v = UnitValue {}
		} else if L == 1 {
			var expr, err = ExprFrom(term.Elements[0], ctx, expected)
			if err != nil { return Expr{}, err }
			T = expr.Type
			v = expr.Value
		} else {
			var el_exprs = make([]Expr, L)
			var el_types = make([]Type, L)
			for i, el := range term.Elements {
				var expr, err = ExprFrom(el, ctx, nil)
				if err != nil {
					return Expr{}, err
				}
				el_exprs[i] = expr
				el_types[i] = expr.Type
			}
			T = AnonymousType { Tuple { Elements: el_types } }
			v = Product { Values: el_exprs }
		}
	case node.Bundle:

	}
	var info = ExprInfo { ErrorPoint: ErrorPoint {
		AST: ctx.TypeCtx.Module.AST,
		Node: t.Node,
	} }
	var expr = Expr { Type: T, Value: v, Info: info, }
	return AssignTo(expected, expr, ctx.TypeCtx)
}

func AssignTo (expected Type, expr Expr, ctx TypeContext) (Expr, *ExprError) {
	if expected == nil {
		// 1. If the expected type is not specified,
		//    no further process is required.
		return expr, nil
	} else if AreTypesEqualInSameCtx(expected, expr.Type) {
		// 2. If the expected type is identical to the given type,
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
					From:   DescribeType(expr.Type, ctx),
					To:     DescribeType(expected, ctx),
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
						Type:  expected,
						Value: Sum { Value: expr, Index: uint(index) },
						Info:  expr.Info,
					}, nil
				}
			}
			// 1.2. Otherwise, throw an error.
			return Expr{}, throw("given type is not a subtype of the expected union type")
		}
		// -- behavior of unpacking wrapped inner types --
		var assign_inner = func(inner Type, g *GenericType, mod string) (Expr, *ExprError) {
			// 1. Create an alternative expression with the inner type
			var expr_with_inner = Expr {
				Type:  inner,
				Value: expr.Value,
				Info:  expr.Info,
			}
			// 2. Try to assign the created expression to the expected type
			var result, err = AssignTo(expected, expr_with_inner, ctx)
			if err != nil { return Expr{}, throw("") }
			// 3. Check if the module encapsulation is violated
			var ctx_mod = loader.Id2String(ctx.Module.Node.Name)
			if g.IsOpaque && ctx_mod != mod {
				return Expr{}, throw("cannot cast out of opaque type")
			} else {
				return result, nil
			}
		}
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
					if err != nil { return Expr{}, err }
					items[i] = item
				}
				// 2.1.2. Collect all adapted elements as a new tuple.
				return Expr {
					Type:  expected,
					Value: Product { Values: items },
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
				var fields = make([]Expr, len(exp.Index))
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
						if err != nil { return Expr{}, err }
						fields[index] = field
					} else {
						// 1.2.2. Otherwise, if an expected field is missing
						if IsMaybeType(field_exp_type) {
							// 1.2.2.1. If the expected field type is Maybe[T],
							//          adapt a Nothing value
							fields[index] = Expr {
								Type:  field_exp_type,
								Value: UnitWithIndex(__Nothing, expr.Info),
								Info:  expr.Info,
							}
						} else {
							// 1.2.2.2. Otherwise, throw an error.
							return Expr{}, throw("missing field " + name)
						}
					}
				}
				// 1.3. Collect all adapted fields as a new bundle.
				return Expr {
					Type:  expected,
					Value: Product { Values: fields },
					Info:  expr.Info,
				}, nil
			default:
				// 2. Otherwise, a non-literal bundle is not adaptable.
				return Expr{}, throw("non-literal bundle cannot be assigned to different bundle type")
			}
		}
		// 3.2. Determine the conversion behavior according to type details
		switch G := expr.Type.(type) {
		case NamedType:
			// 3.2.1. If the given type is a named type
			var reg = ctx.Ireg.(TypeRegistry)
			var given_g = reg[G.Name]
			switch E := expected.(type) {
			case NamedType:
				var expected_g = reg[E.Name]
				switch union := expected_g.Value.(type) {
				case Union:
					// 3.2.1.1. If the expected type is a union type,
					//          try to lift the given expression
					//          to the union type.
					return assign_union(E, G, union)
				}
			}
			switch tv := given_g.Value.(type) {
			case Single:
				var given_inner = tv.InnerType
				var given_mod = G.Name.ModuleName
				// 3.2.1.2. Otherwise, if the given type has an inner type,
				//          try to unpack the inner type.
				return assign_inner(given_inner, given_g, given_mod)
			}
		case AnonymousType:
			// 3.2.2. If the given type is an anonymous type
			switch repr_given := G.Repr.(type) {
			case Tuple:
				switch E := expected.(type) {
				case AnonymousType:
					switch repr_expected := E.Repr.(type) {
					case Tuple:
						// 3.2.2.1. If types are both tuple type,
						//          try to adapt the given tuple.
						return assign_tuple(repr_expected, repr_given)
					default:
					}
				}
			case Bundle:
				switch E := expected.(type) {
				case AnonymousType:
					switch re := E.Repr.(type) {
					case Bundle:
						// 3.2.2.2. If types are both bundle type,
						//          try to adapt the given bundle.
						return assign_bundle(re, repr_given)
					}
				}
			}
		}
		// 3.3. If no conversion is available, types are not compatible.
		return Expr{}, throw("")
	}
}