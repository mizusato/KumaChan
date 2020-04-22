package checker

import (
	"kumachan/transformer/ast"
)


func (impl SemiTypedArray) SemiExprVal() {}
type SemiTypedArray struct {
	Items  [] SemiExpr
}

func (impl Array) ExprVal() {}
type Array struct {
	Items  [] Expr
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
	if expected == nil {
		var item_type Type = nil
		var items = make([]Expr, len(array.Items))
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
			array_t = NamedType {
				Name: __Array,
				Args: []Type { WildcardRhsType {} },
			}
		} else {
			if item_type == nil { panic("something went wrong") }
			array_t = NamedType {
				Name: __Array,
				Args: []Type { item_type },
			}
		}
		return Expr {
			Type:  array_t,
			Info:  info,
			Value: Array { items },
		}, nil
	}
	switch E := expected.(type) {
	case NamedType:
		if E.Name == __Array {
			if len(E.Args) != 1 { panic("something went wrong") }
			var item_expected = E.Args[0]
			var items = make([]Expr, len(array.Items))
			for i, item_semi := range array.Items {
				var item, err = AssignTo(item_expected, item_semi, ctx)
				if err != nil { return Expr{}, err }
				items[i] = item
			}
			var point = info.ErrorPoint
			var item_type, err = GetCertainType(item_expected, point, ctx)
			if err != nil { return Expr{}, err }
			return Expr {
				Type:  NamedType {
					Name: __Array,
					Args: []Type { item_type },
				},
				Info:  info,
				Value: Array { items },
			}, nil
		}
	}
	return Expr{}, &ExprError {
		Point:    info.ErrorPoint,
		Concrete: E_ArrayAssignedToNonArrayType {},
	}
}
