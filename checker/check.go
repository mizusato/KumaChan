package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/transformer/node"
)

type ExprContext struct {
	TypeCtx  TypeContext
	ValMap   map[loader.Symbol] Type
	FunMap   FunctionCollection
}

func ExprFrom(e node.Expr, ctx ExprContext, expected Type) (Expr, *ExprError) {
	// TODO
	return Expr{}, nil
}

func ExprFromPipe(p node.Pipe, ctx ExprContext, input Type) (Expr, *ExprError) {
	// TODO
	// if input == nil { ...
	return Expr{}, nil
}

func ExprFromTerm(t node.VariousTerm, ctx ExprContext, expected Type) (Expr, *ExprError) {
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
