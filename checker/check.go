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
	if input == nil {

	} else {

	}
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
			v = Product{Values: el_exprs}
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
	// TODO
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
			switch tv := gt.Value.(type) {
			case UnionTypeVal:
				return Expr{}, throw("")
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
			case NativeTypeVal:
				return Expr{}, throw("")
			default:
				panic("impossible branch")
			}
		case AnonymousType:
			// TODO
		}
		return Expr{}, nil
	}
}