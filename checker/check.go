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
		return expr, nil
	} else if AreTypesEqualInSameCtx(expected, expr.Type) {
		return expr, nil
	} else {
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
		switch T := expr.Type.(type) {
		case ParameterType:
			return Expr{}, throw("")
		case NamedType:
			var reg = ctx.Ireg.(TypeRegistry)
			var gt = reg[T.Name]
			switch E := expected.(type) {
			case NamedType:
				var ge = reg[E.Name]
				switch tv := ge.Value.(type) {
				case UnionTypeVal:
					for index, subtype := range tv.SubTypes {
						if subtype == T.Name {
							return Expr {
								Type: expected,
								Value: Sum {
									Value: expr,
									Index: uint(index),
								},
								Info: expr.Info,
							}, nil
						}
					}
				}
			}
			switch tv := gt.Value.(type) {
			case SingleTypeVal:
				var inner = tv.InnerType
				var with_inner = Expr {
					Type: inner,
					Value: expr.Value,
					Info: expr.Info,
				}
				var result, err = AssignTo(expected, with_inner, ctx)
				if err != nil {
					return Expr{}, throw("")
				} else {
					var ctx_mod = string(ctx.Module.Node.Name.Name)
					if gt.IsOpaque && T.Name.ModuleName != ctx_mod {
						return Expr{}, throw("cannot cast out of opaque type")
					} else {
						return result, nil
					}
				}
			default:
				return Expr{}, throw("")
			}
		case AnonymousType:
			switch rt := T.Repr.(type) {
			case Tuple:
				switch E := expected.(type) {
				case AnonymousType:
					switch re := E.Repr.(type) {
					case Tuple:
						if len(rt.Elements) != len(re.Elements) {
							return Expr{}, throw("tuple arity not matching")
						}
						switch v := expr.Value.(type) {
						case Product:
							var L = len(rt.Elements)
							var items = make([]Expr, L)
							for i := 0; i < L; i += 1 {
								var raw_item = v.Values[i]
								var item_expected = re.Elements[i]
								var item, err = AssignTo(item_expected, raw_item, ctx)
								if err != nil { return Expr{}, err }
								items[i] = item
							}
							return Expr {
								Type: expected,
								Value: Product {
									Values: items,
								},
								Info: expr.Info,
							}, nil
						default:
							return Expr{}, throw("non-literal tuple cannot be assigned to different tuple type")
						}
					default:
						return Expr{}, throw("")
					}
				default:
					return Expr{}, throw("")
				}
			case Bundle:
				switch E := expected.(type) {
				case AnonymousType:
					switch re := E.Repr.(type) {
					case Bundle:
						var Lt = len(rt.Fields)
						var Le = len(re.Fields)
						if Lt == Le {
							switch v := expr.Value.(type) {
							case Product:
								var L = Lt
								var fields = make([]Expr, L)
								for name, index := range re.Index {
									var field_expected = re.Fields[name]
									var raw_index, exists = rt.Index[name]
									if !exists { return Expr{}, throw("missing field " + name) }
									var raw_field = v.Values[raw_index]
									var field, err = AssignTo(field_expected, raw_field, ctx)
									if err != nil { return Expr{}, err }
									fields[index] = field
								}
								return Expr {
									Type: expected,
									Value: Product {
										Values: fields,
									},
									Info: expr.Info,
								}, nil
							default:
								return Expr{}, throw("non-literal bundle cannot be assigned to different bundle type")
							}
						} else if Lt < Le {
							// TODO: fill Nothing for Maybe fields
							panic("not implemented")
						} else {
							return Expr{}, throw("")
						}
					}
				default:
					return Expr{}, throw("")
				}
			default:
				return Expr{}, throw("")
			}
		}
		return Expr{}, nil
	}
}