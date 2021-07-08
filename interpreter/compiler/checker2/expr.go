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
type localBindingMap (map[string] *checked.LocalBinding)
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
func (lm localBindingMap) add(binding *checked.LocalBinding) {
	lm[binding.Name] = binding
}
func (lm localBindingMap) lookup(name string) (*checked.LocalBinding, bool) {
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

type ExprChecker func
	(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (
	*checked.Expr, *typsys.InferringState, *source.Error)

func assign(expr *checked.Expr) ExprChecker {
	return ExprChecker(func(expected typsys.Type, s0 *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		if expected == nil {
			return expr, nil, nil
		} else {
			var assign_ctx = ctx.makeAssignContext(s0)
			var ok, s1 = typsys.Assign(expected, expr.Type, assign_ctx)
			if ok {
				return expr, s1, nil
			} else {
				return nil, nil, source.MakeError(expr.Info.Location,
					E_NotAssignable {
						From: typsys.DescribeType(expr.Type, s0),
						To:   typsys.DescribeType(expected, s0),
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
// TODO: consider describeType()
func (cc checkContext) fork() checkContext {
	var s = *(cc.inferring)
	return checkContext {
		location:    cc.location,
		inferring:   &s,
		exprContext: cc.exprContext,
	}
}
func (cc checkContext) productPatternMatch(pattern ast.VariousPattern, in typsys.Type) (checked.ProductPatternInfo, *source.Error) {
	var mod = cc.exprContext.ModName
	var lm = cc.exprContext.localBindingMap
	return productPatternMatch(pattern, in, mod, lm)
}
func (cc checkContext) checkChildExpr(expected typsys.Type, node ast.Expr) (*checked.Expr, *source.Error) {
	var expr, s, err = check(node)(expected, *(cc.inferring), cc.exprContext)
	if err != nil { return nil, err }
	*(cc.inferring) = s
	return expr, nil
}
func (cc checkContext) checkChildTerm(expected typsys.Type, node ast.VariousTerm) (*checked.Expr, *source.Error) {
	return cc.checkChildExpr(expected, ast.WrapTermAsExpr(node))
}
func (cc checkContext) assignType(to typsys.Type, from typsys.Type) bool {
	var a = typsys.MakeAssignContext(cc.exprContext.ModName, *(cc.inferring))
	var ok, s = typsys.Assign(to, from, a)
	if ok {
		*(cc.inferring) = s
		return true
	} else {
		return false
	}
}
func (cc checkContext) assign(expected typsys.Type, t typsys.Type, content checked.ExprContent) (*checked.Expr, *typsys.InferringState, *source.Error) {
	return assign(&checked.Expr {
		Type:    t,
		Info:    checked.ExprInfoFrom(cc.location),
		Content: content,
	})(expected, *(cc.inferring), cc.exprContext)
}
func (cc checkContext) ok(t typsys.Type, content checked.ExprContent) (*checked.Expr, *typsys.InferringState, *source.Error) {
	return &checked.Expr {
		Type:    t,
		Info:    checked.ExprInfoFrom(cc.location),
		Content: content,
	}, *(cc.inferring), nil
}
func (cc checkContext) error(content source.ErrorContent) (*checked.Expr, *typsys.InferringState, *source.Error) {
	return nil, nil, source.MakeError(cc.location, content)
}
func (cc checkContext) propagate(err *source.Error) (*checked.Expr, *typsys.InferringState, *source.Error) {
	if err == nil {
		panic("something went wrong")
	}
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
	case ast.Lambda:
		return checkLambda(T)
	default:
		// TODO
	}
}

func checkPipe(in *checked.Expr, pipe ast.VariousPipe) ExprChecker {
	// TODO
}


