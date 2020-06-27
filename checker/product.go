package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/parser/ast"
)


func (impl SemiTypedTuple) SemiExprVal() {}
type SemiTypedTuple struct {
	Values  [] SemiExpr
}

func (impl SemiTypedBundle) SemiExprVal() {}
type SemiTypedBundle struct {
	Index     map[string] uint
	Values    [] SemiExpr
	KeyNodes  [] ast.Node
}

func (impl Product) ExprVal() {}
type Product struct {
	Values  [] Expr
}

func (impl Get) ExprVal() {}
type Get struct {
	Product  Expr
	Index    uint
}

func (impl Set) ExprVal() {}
type Set struct {
	Product   Expr
	Index     uint
	NewValue  Expr
}


func CheckTuple(tuple ast.Tuple, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(tuple.Node)
	var L = len(tuple.Elements)
	if L == 0 {
		return LiftTyped(Expr {
			Type:  &AnonymousType { Unit {} },
			Value: UnitValue {},
			Info:  info,
		}), nil
	} else if L == 1 {
		var expr, err = Check(tuple.Elements[0], ctx)
		if err != nil { return SemiExpr{}, err }
		return expr, nil
	} else {
		var el_exprs = make([]SemiExpr, L)
		var el_types = make([]Type, L)
		for i, el := range tuple.Elements {
			var expr, err = Check(el, ctx)
			if err != nil { return SemiExpr{}, err }
			el_exprs[i] = expr
			switch typed := expr.Value.(type) {
			case TypedExpr:
				el_types[i] = typed.Type
			}
		}
		return SemiExpr {
			Value: SemiTypedTuple { el_exprs },
			Info: info,
		}, nil
	}
}

func CheckBundle(bundle ast.Bundle, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(bundle.Node)
	switch update := bundle.Update.(type) {
	case ast.Update:
		var base_semi, err = Check(update.Base, ctx)
		if err != nil { return SemiExpr{}, err }
		switch b := base_semi.Value.(type) {
		case TypedExpr:
			if IsBundleLiteral(Expr(b)) { return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(update.Base.Node),
				Concrete: E_SetToLiteralBundle {},
			} }
			var L = len(bundle.Values)
			if !(L >= 1) { panic("something went wrong") }
			var base = Expr(b)
			switch target := UnboxBundle(base.Type, ctx).(type) {
			case Bundle:
				var occurred_names = make(map[string] bool)
				var current_base = base
				for _, field := range bundle.Values {
					var name = loader.Id2String(field.Key)
					var target_field, exists = target.Fields[name]
					if !exists {
						return SemiExpr{}, &ExprError {
							Point: ErrorPointFrom(field.Key.Node),
							Concrete: E_FieldDoesNotExist {
								Field:  name,
								Target: ctx.DescribeType(base.Type),
							},
						}
					}
					var _, duplicate = occurred_names[name]
					if duplicate {
						return SemiExpr{}, &ExprError {
							Point:    ErrorPointFrom(field.Key.Node),
							Concrete: E_ExprDuplicateField { name },
						}
					}
					occurred_names[name] = true
					var value_node = DesugarOmittedFieldValue(field)
					var value_semi, err1 = Check(value_node, ctx)
					if err1 != nil { return SemiExpr{}, err1 }
					var value, err2 = AssignTo(target_field.Type, value_semi, ctx)
					if err2 != nil { return SemiExpr{}, err2 }
					current_base = Expr {
						Type:  current_base.Type,
						Value: Set {
							Product:  current_base,
							Index:    target_field.Index,
							NewValue: value,
						},
						Info:  current_base.Info,
					}
				}
				var final = current_base
				return SemiExpr {
					Value: TypedExpr(final),
					Info:  info,
				}, nil
			case BR_BundleButOpaque:
				return SemiExpr{}, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_SetToOpaqueBundle {},
				}
			case BR_NonBundle:
				return SemiExpr{}, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_SetToNonBundle {},
				}
			default:
				panic("impossible branch")
			}
		case SemiTypedBundle:
			return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(update.Base.Node),
				Concrete: E_SetToLiteralBundle {},
			}
		default:
			return SemiExpr{}, &ExprError {
				Point:    ErrorPointFrom(update.Base.Node),
				Concrete: E_SetToNonBundle {},
			}
		}
	default:
		var L = len(bundle.Values)
		if L == 0 {
			return LiftTyped(Expr {
				Type:  &AnonymousType { Unit {} },
				Value: UnitValue {},
				Info:  info,
			}), nil
		} else {
			var f_exprs = make([]SemiExpr, L)
			var f_index_map = make(map[string]uint, L)
			var f_key_nodes = make([]ast.Node, L)
			for i, field := range bundle.Values {
				var name = loader.Id2String(field.Key)
				var _, exists = f_index_map[name]
				if exists { return SemiExpr{}, &ExprError {
					Point:    ErrorPointFrom(field.Key.Node),
					Concrete: E_ExprDuplicateField { name },
				} }
				var value = DesugarOmittedFieldValue(field)
				var expr, err = Check(value, ctx)
				if err != nil { return SemiExpr{}, err }
				f_exprs[i] = expr
				f_index_map[name] = uint(i)
				f_key_nodes[i] = field.Key.Node
			}
			return SemiExpr {
				Value: SemiTypedBundle {
					Index:    f_index_map,
					Values:   f_exprs,
					KeyNodes: f_key_nodes,
				},
				Info: info,
			}, nil
		}
	}
}

func CheckGet(get ast.Get, ctx ExprContext) (SemiExpr, *ExprError) {
	var base_semi, err = Check(get.Base, ctx)
	if err != nil { return SemiExpr{}, err }
	switch b := base_semi.Value.(type) {
	case TypedExpr:
		if IsBundleLiteral(Expr(b)) { return SemiExpr{}, &ExprError {
			Point:    ErrorPointFrom(get.Base.Node),
			Concrete: E_GetFromLiteralBundle {},
		} }
		var L = len(get.Path)
		if !(L >= 1) { panic("something went wrong") }
		var base = Expr(b)
		for _, member := range get.Path {
			switch bundle := UnboxBundle(base.Type, ctx).(type) {
			case Bundle:
				var key = loader.Id2String(member.Name)
				var field, exists = bundle.Fields[key]
				if !exists { return SemiExpr{}, &ExprError {
					Point:    ErrorPointFrom(member.Node),
					Concrete: E_FieldDoesNotExist {
						Field:  key,
						Target: ctx.DescribeType(&AnonymousType { bundle }),
					},
				} }
				var expr = Expr {
					Type: field.Type,
					Value: Get {
						Product: Expr(base),
						Index:   field.Index,
					},
					Info:  ctx.GetExprInfo(member.Node),
				}
				base = expr
			case BR_BundleButOpaque:
				return SemiExpr{}, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_GetFromOpaqueBundle {},
				}
			case BR_NonBundle:
				return SemiExpr{}, &ExprError {
					Point:    base.Info.ErrorPoint,
					Concrete: E_GetFromNonBundle {},
				}
			default:
				panic("impossible branch")
			}
		}
		var final = base
		return LiftTyped(final), nil
	case SemiTypedBundle:
		return SemiExpr{}, &ExprError {
			Point:    ErrorPointFrom(get.Base.Node),
			Concrete: E_GetFromLiteralBundle {},
		}
	default:
		return SemiExpr{}, &ExprError {
			Point:    ErrorPointFrom(get.Base.Node),
			Concrete: E_GetFromNonBundle {},
		}
	}
}


func AssignTupleTo(expected Type, tuple SemiTypedTuple, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var non_nil_expected Type
	if expected == nil {
		non_nil_expected = &AnonymousType {
			Tuple {
				// Fill with nil
				Elements: make([]Type, len(tuple.Values)),
			},
		}
	} else {
		non_nil_expected = expected
	}
	switch E := non_nil_expected.(type) {
	default:
		var typed_exprs = make([]Expr, len(tuple.Values))
		for i, el := range tuple.Values {
			var typed, err = AssignTo(nil, el, ctx)
			if err != nil { return Expr{}, err }
			typed_exprs[i] = typed
		}
		var el_types = make([]Type, len(tuple.Values))
		for i, el := range typed_exprs {
			el_types[i] = el.Type
		}
		var final_t = &AnonymousType { Tuple { el_types } }
		var typed_tuple = Expr {
			Type:  final_t,
			Value: Product { typed_exprs },
			Info:  info,
		}
		return TypedAssignTo(expected, typed_tuple, ctx)
	case *AnonymousType:
		switch tuple_t := E.Repr.(type) {
		case Tuple:
			var required = len(tuple_t.Elements)
			var given = len(tuple.Values)
			if given != required {
				return Expr{}, &ExprError {
					Point:    info.ErrorPoint,
					Concrete: E_TupleSizeNotMatching {
						Required:  required,
						Given:     given,
						GivenType: ctx.DescribeExpectedType(&AnonymousType { tuple_t }),
					},
				}
			}
			var typed_exprs = make([]Expr, given)
			for i, el := range tuple.Values {
				var el_expected = tuple_t.Elements[i]
				var typed, err = AssignTo(el_expected, el, ctx)
				if err != nil { return Expr{}, err }
				typed_exprs[i] = typed
			}
			var el_types = make([]Type, len(tuple.Values))
			for i, el := range typed_exprs {
				el_types[i] = el.Type
			}
			var final_t = &AnonymousType { Tuple { el_types } }
			return Expr {
				Type:  final_t,
				Info:  info,
				Value: Product { typed_exprs },
			}, nil
		}
	}
	return Expr{}, &ExprError {
		Point:    info.ErrorPoint,
		Concrete: E_TupleAssignedToNonTupleType {
			NonTupleType: ctx.DescribeExpectedType(non_nil_expected),
		},
	}
}

func AssignBundleTo(expected Type, bundle SemiTypedBundle, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	var err = RequireExplicitType(expected, info)
	if err != nil { return Expr{}, err }
	switch E := expected.(type) {
	case *AnonymousType:
		switch bundle_t := E.Repr.(type) {
		case Bundle:
			var values = make([]Expr, len(bundle_t.Fields))
			for field_name, field := range bundle_t.Fields {
				var given_index, exists = bundle.Index[field_name]
				if !exists {
					return Expr{}, &ExprError {
						Point:    info.ErrorPoint,
						Concrete: E_MissingField {
							Field: field_name,
							Type:  ctx.DescribeExpectedType(field.Type),
						},
					}
				}
				var given_value = bundle.Values[given_index]
				var value, err = AssignTo(field.Type, given_value, ctx)
				if err != nil { return Expr{}, err }
				values[field.Index] = value
			}
			for given_field_name, index := range bundle.Index {
				var _, exists = bundle_t.Fields[given_field_name]
				if !exists {
					var key_node = bundle.KeyNodes[index]
					return Expr{}, &ExprError {
						Point:    ErrorPointFrom(key_node),
						Concrete: E_SuperfluousField { given_field_name },
					}
				}
			}
			var final_fields = make(map[string]Field)
			for field_name, field := range bundle_t.Fields {
				final_fields[field_name] = Field {
					Type:  values[field.Index].Type,
					Index: field.Index,
				}
			}
			var final_t = &AnonymousType { Bundle{ final_fields } }
			return Expr {
				Type:  final_t,
				Info:  info,
				Value: Product { values },
			}, nil
		}
	}
	return  Expr{}, &ExprError {
		Point:    info.ErrorPoint,
		Concrete: E_BundleAssignedToNonBundleType {
			NonBundleType: ctx.DescribeExpectedType(expected),
		},
	}
}


func IsBundleLiteral(expr Expr) bool {
	switch expr.Value.(type) {
	case Product:
		switch t := expr.Type.(type) {
		case *AnonymousType:
			switch t.Repr.(type) {
			case Bundle:
				return true
			}
		}
	}
	return false
}

func DesugarOmittedFieldValue(field ast.FieldValue) ast.Expr {
	switch val_expr := field.Value.(type) {
	case ast.Expr:
		return val_expr
	default:
		return ast.WrapCallAsExpr(ast.Call {
			Node: field.Node,
			Func: ast.VariousTerm {
				Node: field.Node,
				Term: ast.InlineRef {
					Node:     field.Node,
					Module:   ast.Identifier {
						Node: field.Node,
						Name: []rune(""),
					},
					Specific: false,
					Id:       field.Key,
					TypeArgs: make([]ast.VariousType, 0),
				},
			},
		})
	}
}
