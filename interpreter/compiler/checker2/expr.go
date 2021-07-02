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
type LocalBindingMap (map[string] LocalBinding)
type LocalBinding struct {
	Type      typsys.Type
	Location  source.Location
}
func (ctx ExprContext) AssignContext() typsys.AssignContext {
	return typsys.MakeAssignContext(ctx.ModName, ctx.InferringState)
}
func (ctx ExprContext) DescribeType(t typsys.Type, s *typsys.InferringState) string {
	if s != nil {
		return typsys.DescribeType(t, s)
	} else {
		return typsys.DescribeType(t, ctx.InferringState)
	}
}

type ExprChecker func
	(expected typsys.Type, ctx ExprContext) (
	*checked.Expr, *typsys.InferringState, *source.Error)

func check(expr ast.Expr) ExprChecker {
	return ExprChecker(func(expected typsys.Type, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var L = len(expr.Pipeline)
		if L == 0 {
			return checkTerm(expr.Term)(expected, ctx)
		} else {
			var current, _, err = checkTerm(expr.Term)(nil, ctx)
			if err != nil { return nil, nil, err }
			var last = (L - 1)
			for _, pipe := range expr.Pipeline[:last] {
				var new_current, _, err  = checkPipe(current, pipe)(nil, ctx)
				if err != nil { return nil, nil, err }
				current = new_current
			}
			return checkPipe(current, expr.Pipeline[last])(expected, ctx)
		}
	})
}

func checkTerm(term ast.VariousTerm) ExprChecker {
	// TODO
}

func checkPipe(in *checked.Expr, pipe ast.VariousPipe) ExprChecker {
	// TODO
}

func CheckLambda(lambda ast.Lambda) ExprChecker {
	// TODO
}


