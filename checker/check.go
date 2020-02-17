package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
)

type ExprContext struct {
	TypeCtx  TypeExprContext
	ValMap   map[loader.Symbol]Type
	FunMap   FunctionCollection
}

func ExprFrom (e node.Expr, ctx ExprContext, expected Type) (Expr, *ExprError) {
	// TODO
	return Expr{}, nil
}

