package checker

import (
	"kumachan/loader/parser/ast"
	"kumachan/runtime/common"
	"reflect"
	"kumachan/stdlib"
)


func (impl SemiTypedArray) SemiExprVal() {}
type SemiTypedArray struct {
	Items  [] SemiExpr
}

func (impl Array) ExprVal() {}
type Array struct {
	Items     [] Expr
	ItemType  Type
}


func CheckArray(array ast.Array, ctx ExprContext) (SemiExpr, *ExprError) {
	var info = ctx.GetExprInfo(array.Node)
	var L = len(array.Items)
	if L == 0 {
		return SemiExpr {
			Value: SemiTypedArray { make([]SemiExpr, 0) },
			Info: info,
		}, nil
	} else {
		var item_exprs = make([]SemiExpr, L)
		for i, item_node := range array.Items {
			var item, err = Check(item_node, ctx)
			if err != nil { return SemiExpr{}, err }
			item_exprs[i] = item
		}
		return SemiExpr {
			Value: SemiTypedArray { item_exprs },
			Info:  info,
		}, nil
	}
}


func AssignArrayTo(expected Type, array SemiTypedArray, info ExprInfo, ctx ExprContext) (Expr, *ExprError) {
	switch E := expected.(type) {
	default:
		var param, is_param = expected.(*ParameterType)
		if is_param {
			if param.BeingInferred {
				if !(ctx.Inferring.Enabled) { panic("something went wrong") }
				var inferred, exists = ctx.Inferring.Arguments[param.Index]
				if exists {
					return AssignArrayTo(inferred, array, info, ctx)
				}
			} else {
				var sub, has_sub = ctx.TypeBounds.Sub[param.Index]
				if has_sub {
					return AssignArrayTo(sub, array, info, ctx)
				}
			}
		}
		var item_type Type = nil
		var items = make([] Expr, len(array.Items))
		for i, item_semi := range array.Items {
			var item, err = AssignTo(item_type, item_semi, ctx)
			if err != nil { return Expr{}, err }
			items[i] = item
			if item_type == nil {
				item_type = item.Type
			}
		}
		var array_t Type
		if len(array.Items) == 0 {
			array_t = &NamedType {
				Name: __Array,
				Args: []Type { &NeverType {} },
			}
		} else {
			if item_type == nil { panic("something went wrong") }
			array_t = &NamedType {
				Name: __Array,
				Args: []Type { item_type },
			}
		}
		var typed_array = Expr {
			Type:  array_t,
			Info:  info,
			Value: Array { Items: items, ItemType: item_type },
		}
		return TypedAssignTo(expected, typed_array, ctx)
	case *NamedType:
		if E.Name == __Array {
			if len(E.Args) != 1 { panic("something went wrong") }
			var item_expected = E.Args[0]
			if len(array.Items) == 0 {
				var empty_array = Expr {
					Type:  &NamedType {
						Name: __Array,
						Args: [] Type { &NeverType {} },
					},
					Value: Array { Items: [] Expr {}, ItemType: nil },
					Info:  info,
				}
				return TypedAssignTo(expected, empty_array, ctx)
			}
			var items = make([] Expr, len(array.Items))
			for i, item_semi := range array.Items {
				var item, err = AssignTo(item_expected, item_semi, ctx)
				if err != nil { return Expr{}, err }
				items[i] = item
			}
			var point = info.ErrorPoint
			var item_type, err = GetCertainType(item_expected, point, ctx)
			if err != nil { return Expr{}, err }
			return Expr {
				Type:  &NamedType {
					Name: __Array,
					Args: []Type { item_type },
				},
				Info:  info,
				Value: Array { Items: items, ItemType: item_type },
			}, nil
		}
	}
	return Expr{}, &ExprError {
		Point:    info.ErrorPoint,
		Concrete: E_ArrayAssignedToNonArrayType {
			NonArrayType: ctx.DescribeInferredType(expected),
		},
	}
}

func GetArrayInfo(length uint, item_type Type) common.ArrayInfo {
	var item_reflect_type = (func() reflect.Type {
		switch t := item_type.(type) {
		case *NamedType:
			var mod_name = t.Name.ModuleName
			var sym_name = t.Name.SymbolName
			if mod_name == stdlib.Core && len(t.Args) == 0 {
				var rt, ok = stdlib.GetPrimitiveReflectType(sym_name)
				if ok {
					return rt
				}
			}
			return common.ValueReflectType()
		default:
			return common.ValueReflectType()
		}
	})()
	return common.ArrayInfo {
		Length:   length,
		ItemType: item_reflect_type,
	}
}

