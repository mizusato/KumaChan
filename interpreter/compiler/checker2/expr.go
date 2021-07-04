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
}
type LocalBindingMap (map[string] LocalBinding)
type LocalBinding struct {
	Type      typsys.Type
	Location  source.Location
}
func (ctx ExprContext) makeAssignContext(s *typsys.InferringState) typsys.AssignContext {
	return typsys.MakeAssignContext(ctx.ModName, s)
}
func (ctx ExprContext) applyFinalCheck (
	expected  typsys.Type,
	s         *typsys.InferringState,
	checker   ExprChecker,
) (*checked.Expr, *typsys.InferringState, *source.Error) {
	return checker(expected, s, ctx)
}
func (ctx ExprContext) applyIntermediateCheck (
	expected  typsys.Type,
	s_ptr     **typsys.InferringState,
	checker   ExprChecker,
) (*checked.Expr, *source.Error) {
	if s_ptr == nil {
		panic("invalid argument")
	} else {
		var expr, s, err = checker(expected, nil, ctx)
		if err != nil { return nil, err }
		*s_ptr = s
		return expr, nil
	}
}

type ExprChecker func
	(expected typsys.Type, s0 *typsys.InferringState, ctx ExprContext) (
	*checked.Expr, *typsys.InferringState, *source.Error)

func assign(expr *checked.Expr) ExprChecker {
	return ExprChecker(func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		if expected == nil {
			return expr, nil, nil
		} else {
			var assign_ctx = ctx.makeAssignContext(s)
			var ok, s = typsys.Assign(expected, expr.Type, assign_ctx)
			if ok {
				return expr, s, nil
			} else {
				return nil, nil, source.MakeError(expr.Info.Location,
					E_NotAssignable {
						From: typsys.DescribeType(expr.Type, s),
						To:   typsys.DescribeType(expected, s),
					})
			}
		}
	})
}

func check(expr ast.Expr) ExprChecker {
	return ExprChecker(func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var L = len(expr.Pipeline)
		if L == 0 {
			return checkTerm(expr.Term)(expected, s, ctx)
		} else {
			var current, _, err = checkTerm(expr.Term)(nil, s, ctx)
			if err != nil { return nil, nil, err }
			var last = (L - 1)
			for _, pipe := range expr.Pipeline[:last] {
				var new_current, _, err  = checkPipe(current, pipe)(nil, s, ctx)
				if err != nil { return nil, nil, err }
				current = new_current
			}
			return checkPipe(current, expr.Pipeline[last])(expected, s, ctx)
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


