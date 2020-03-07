package checker

import "kumachan/transformer/node"


func (impl SemiTypedArray) SemiExprVal() {}
type SemiTypedArray struct {
	Items  [] SemiExpr
}

func (impl Array) ExprVal() {}
type Array struct {
	Items  [] Expr
}


func CheckArray(array node.Array, ctx ExprContext) (SemiExpr, *ExprError) {
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
		var cur_item_type Type = nil
		var items = make([]Expr, len(array.Items))
		for i, item_semi := range array.Items {
			var item, err = AssignTo(cur_item_type, item_semi, ctx)
			if err != nil { return Expr{}, err }
			items[i] = item
		}
		return Expr {
			Type:  expected,
			Info:  info,
			Value: Array { items },
		}, nil
	}
	switch E := expected.(type) {
	case NamedType:
		if E.Name == __Array {
			if len(E.Args) != 1 { panic("something went wrong") }
			var item_type = E.Args[0]
			var items = make([]Expr, len(array.Items))
			for i, item_semi := range array.Items {
				var item, err = AssignTo(item_type, item_semi, ctx)
				if err != nil { return Expr{}, err }
				items[i] = item
			}
			return Expr {
				Type:  expected,
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
