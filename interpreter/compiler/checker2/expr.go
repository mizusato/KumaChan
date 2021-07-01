package checker2

import (
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
)


type Registry struct {
	Aliases    AliasRegistry
	Types      TypeRegistry
	Functions  FunctionRegistry
}

type ExprContext struct {
	*Registry
	*ModuleInfo
	LocalBindingMap
	*typsys.InferringState
}
type LocalBindingMap (map[string] typsys.Type)

type ExprChecker func
	(expected typsys.Type, ctx ExprContext) (
	*checked.Expr, *typsys.InferringState, *source.Error)

func Check(expr ast.Expr) ExprChecker {
	return ExprChecker(func(expected typsys.Type, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var L = len(expr.Pipeline)
		if L == 0 {
			return CheckTerm(expr.Term)(expected, ctx)
		} else {
			var current, _, err = CheckTerm(expr.Term)(nil, ctx)
			if err != nil { return nil, nil, err }
			var last = (L - 1)
			for _, pipe := range expr.Pipeline[:last] {
				var new_current, _, err  = CheckPipe(current, pipe)(nil, ctx)
				if err != nil { return nil, nil, err }
				current = new_current
			}
			return CheckPipe(current, expr.Pipeline[last])(expected, ctx)
		}
	})
}

func CheckTerm(term ast.VariousTerm) ExprChecker {
	// TODO
}

func CheckPipe(in *checked.Expr, pipe ast.VariousPipe) ExprChecker {
	// TODO
}

func CheckLambda(lambda ast.Lambda) ExprChecker {
	// TODO
}


