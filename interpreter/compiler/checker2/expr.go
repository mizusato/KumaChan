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
	localBindingMap
}
type localBindingMap (map[string] *LocalBinding)
type LocalBinding struct {
	Name      string
	Type      typsys.Type
	Location  source.Location
}
func (lm localBindingMap) clone() localBindingMap {
	var clone = make(localBindingMap)
	for k, v := range lm {
		clone[k] = v
	}
	return clone
}
func (lm localBindingMap) mergedTo(another localBindingMap) localBindingMap {
	if another != nil {
		for k, v := range lm {
			another[k] = v
		}
		return another
	} else {
		return lm.clone()
	}
}
func (lm localBindingMap) add(loc source.Location, name string, t typsys.Type) {
	lm[name] = &LocalBinding {
		Name:     name,
		Type:     t,
		Location: loc,
	}
}
func (lm localBindingMap) lookup(name string) (*LocalBinding, bool) {
	var binding, exists = lm[name]
	return binding, exists
}
func (ctx ExprContext) withLocalBindings(lm localBindingMap) ExprContext {
	return ExprContext {
		Registry:        ctx.Registry,
		ModuleInfo:      ctx.ModuleInfo,
		localBindingMap: ctx.localBindingMap.mergedTo(lm),
	}

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
	(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (
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

type checkContext struct {
	location     source.Location
	inferring    **typsys.InferringState
	exprContext  ExprContext
}
func makeCheckContext(loc source.Location, s_ptr **typsys.InferringState, ctx ExprContext, lm localBindingMap) checkContext {
	if s_ptr == nil {
		panic("invalid argument")
	}
	return checkContext {
		location:    loc,
		inferring:   s_ptr,
		exprContext: ctx.withLocalBindings(lm),
	}
}
func (cc checkContext) fork() checkContext {
	var s = *(cc.inferring)
	return checkContext {
		location:    cc.location,
		inferring:   &s,
		exprContext: cc.exprContext,
	}
}
func (cc checkContext) checkExpr(expected typsys.Type, node ast.Expr) (*checked.Expr, *source.Error) {
	var expr, s, err = check(node)(expected, *(cc.inferring), cc.exprContext)
	if err != nil { return nil, err }
	*(cc.inferring) = s
	return expr, nil
}
func (cc checkContext) assign(expected typsys.Type, t typsys.Type, content checked.ExprContent) (*checked.Expr, *typsys.InferringState, *source.Error) {
	return assign(&checked.Expr {
		Type: t,
		Info: checked.ExprInfoFrom(cc.location),
		Expr: content,
	})(expected, *(cc.inferring), cc.exprContext)
}
func (cc checkContext) ok(t typsys.Type, content checked.ExprContent) (*checked.Expr, *typsys.InferringState, *source.Error) {
	return &checked.Expr {
		Type: t,
		Info: checked.ExprInfoFrom(cc.location),
		Expr: content,
	}, *(cc.inferring), nil
}
func (cc checkContext) error(content source.ErrorContent) (*checked.Expr, *typsys.InferringState, *source.Error) {
	return nil, nil, source.MakeError(cc.location, content)
}
func (cc checkContext) propagate(err *source.Error) (*checked.Expr, *typsys.InferringState, *source.Error) {
	return nil, nil, err
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
	switch T := term.Term.(type) {
	case ast.CharLiteral:
		return checkChar(T)
	case ast.FloatLiteral:
		return checkFloat(T)
	case ast.IntegerLiteral:
		return checkInteger(T)
	case ast.Tuple:
		return checkTuple(T)
	}
}

func checkPipe(in *checked.Expr, pipe ast.VariousPipe) ExprChecker {
	// TODO
}

func checkLambda(lambda ast.Lambda) ExprChecker {
	// TODO
}


