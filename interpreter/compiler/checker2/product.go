package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
	"kumachan/interpreter/lang/common/source"
)


func unboxTuple(t typsys.Type, mod string) (typsys.Tuple, bool) {
	var nested, is_nested = t.(*typsys.NestedType)
	if is_nested {
		var tuple, is_tuple = nested.Content.(typsys.Tuple)
		if is_tuple {
			return tuple, true
		} else {
			goto unbox
		}
	} else {
		goto unbox
	}
	unbox:
	var inner, exists = typsys.Unbox(t, mod)
	if exists {
		return unboxTuple(inner, mod)
	} else {
		return typsys.Tuple {}, false
	}
}

func checkTuple(T ast.Tuple) ExprChecker {
	return ExprChecker(func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var cc = makeCheckContext(T.Location, &s, ctx, nil)
		if expected == nil {
			var L = len(T.Elements)
			var elements = make([] *checked.Expr, L)
			var types = make([] typsys.Type, L)
			for i := 0; i < L; i += 1 {
				var el, err = cc.checkExpr(nil, T.Elements[i])
				if err != nil { return nil, nil, err }
				elements[i] = el
				types[i] = el.Type
			}
			var tuple_t = &typsys.NestedType {
				Content: typsys.Tuple { Elements: types },
			}
			return cc.ok(tuple_t, checked.Tuple { Elements: elements })
		} else {
			var tuple, accept_tuple = unboxTuple(expected, ctx.ModName)
			if !(accept_tuple) {
				return cc.error(
					E_TupleAssignedToIncompatible {
						TypeName: typsys.DescribeType(expected, s),
					})
			}
			var L = len(T.Elements)
			var L_required = len(tuple.Elements)
			if L != L_required {
				return cc.error(
					E_TupleSizeNotMatching {
						Required: uint(L_required),
						Given:    uint(L),
					})
			}
			var elements = make([] *checked.Expr, L)
			for i := 0; i < L; i += 1 {
				var el, err = cc.checkExpr(tuple.Elements[i], T.Elements[i])
				if err != nil { return nil, nil, err }
				elements[i] = el
			}
			var tuple_t = &typsys.NestedType { Content: tuple }
			return cc.ok(tuple_t, checked.Tuple { Elements: elements })
		}
	})
}


