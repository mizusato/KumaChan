package checker2

import (
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
)


type ExprContext struct {
	*Registry
	*ModuleInfo
	localBindingMap
}
type Registry struct {
	Aliases    AliasRegistry
	Types      TypeRegistry
	Functions  FunctionRegistry
}
type localBindingMap (map[string] *checked.LocalBinding)

func (lm localBindingMap) clone() localBindingMap {
	var clone = make(localBindingMap)
	for k, v := range lm {
		clone[k] = v
	}
	return clone
}
func (lm localBindingMap) add(binding *checked.LocalBinding) {
	lm[binding.Name] = binding
}
func (lm localBindingMap) lookup(name string) (*checked.LocalBinding, bool) {
	var binding, exists = lm[name]
	return binding, exists
}

func (ctx ExprContext) withNewLocalScope() ExprContext {
	return ExprContext {
		Registry:        ctx.Registry,
		ModuleInfo:      ctx.ModuleInfo,
		localBindingMap: ctx.localBindingMap.clone(),
	}
}
func (ctx ExprContext) makeAssignContext(s *typsys.InferringState) typsys.AssignContext {
	return typsys.MakeAssignContext(ctx.ModName, s)
}


type checkContext struct {
	location     source.Location
	expected     typsys.Type
	inferring    *typsys.InferringState
	exprContext  ExprContext
}
type checkResult struct {
	expr  *checked.Expr
	err   *source.Error
}
func (checkResult) Error() string {
	// abuse IDE inspection to ensure result returned
	panic("dummy method")
}

func makeCheckContext (
	loc  source.Location,
	exp  typsys.Type, // nullable
	s    *typsys.InferringState,
	ctx  ExprContext,
) *checkContext {
	return &checkContext {
		location:    loc,
		expected:    exp,
		inferring:   s,
		exprContext: ctx,
	}
}

func (cc *checkContext) getType(t nominalType) typsys.Type {
	return t(cc.exprContext.Types)
}
func (cc *checkContext) describeType(t typsys.Type) string {
	return typsys.DescribeType(t, cc.inferring)
}
func (cc *checkContext) describeTypeOf(expr *checked.Expr) string {
	return cc.describeType(expr.Type)
}

func (cc *checkContext) forkInferring() *checkContext {
	return makeCheckContext (
		cc.location, cc.expected, cc.inferring, cc.exprContext,
	)
}
func (cc *checkContext) assignType(to typsys.Type, from typsys.Type) *source.Error {
	var a = typsys.MakeAssignContext(cc.exprContext.ModName, cc.inferring)
	var ok, s = typsys.Assign(to, from, a)
	if ok {
		cc.inferring = s
		return nil
	} else {
		return source.MakeError(cc.location, E_NotAssignable {
			From: cc.describeType(from),
			To:   cc.describeType(to),
		})
	}
}

func (cc *checkContext) unboxRecord(expr *checked.Expr) (typsys.Record, bool) {
	return unboxRecord(expr.Type, cc.exprContext.ModName)
}
func (cc *checkContext) unboxLambda(expr *checked.Expr) (typsys.Lambda, bool) {
	return unboxLambda(expr.Type, cc.exprContext.ModName)
}

func (cc *checkContext) checkChildExpr(expected typsys.Type, node ast.Expr) (*checked.Expr, *source.Error) {
	var expr, s, err = check(node)(expected, cc.inferring, cc.exprContext)
	if err != nil { return nil, err }
	cc.inferring = s
	return expr, nil
}
func (cc *checkContext) checkChildTerm(expected typsys.Type, node ast.VariousTerm) (*checked.Expr, *source.Error) {
	return cc.checkChildExpr(expected, ast.WrapTermAsExpr(node))
}

func (cc *checkContext) forwardToChildExpr(node ast.Expr) checkResult {
	var expr, err = cc.checkChildExpr(cc.expected, node)
	return checkResult {
		expr: expr,
		err:  err,
	}
}
func (cc *checkContext) forwardToChildTerm(node ast.VariousTerm) checkResult {
	return cc.forwardToChildExpr(ast.WrapTermAsExpr(node))
}

func (cc *checkContext) assign(t typsys.Type, content checked.ExprContent) checkResult {
	if cc.expected == nil {
		var info = checked.ExprInfoFrom(cc.location)
		var expr = &checked.Expr {
			Type:    t,
			Info:    info,
			Content: content,
		}
		return checkResult { expr: expr }
	} else {
		var info = checked.ExprInfoFrom(cc.location)
		var s0 = cc.inferring
		var assign_ctx = cc.exprContext.makeAssignContext(s0)
		var ok, s1 = typsys.Assign(cc.expected, t, assign_ctx)
		if ok {
			cc.inferring = s1
			var expr = &checked.Expr {
				Type:    t,
				Info:    info,
				Content: content,
			}
			return checkResult { expr: expr }
		} else {
			var err = source.MakeError(info.Location,
				E_NotAssignable {
					From: typsys.DescribeType(t, s0),
					To:   typsys.DescribeType(cc.expected, s0),
				})
			return checkResult { err: err }
		}
	}
}

func (cc *checkContext) error(content source.ErrorContent) checkResult {
	return checkResult { err: source.MakeError(cc.location, content) }
}
func (cc *checkContext) propagate(err *source.Error) checkResult {
	if err == nil {
		panic("something went wrong")
	}
	return checkResult { err: err }
}

type checkContextWithLocalScope struct {
	*checkContext
}
func makeCheckContextWithLocalScope (
	loc  source.Location,
	exp  typsys.Type, // nullable
	s    *typsys.InferringState,
	ctx  ExprContext,
) *checkContextWithLocalScope {
	var cc = makeCheckContext(loc, exp, s, ctx)
	cc.exprContext = cc.exprContext.withNewLocalScope()
	return &checkContextWithLocalScope { checkContext: cc }
}
func (cc *checkContextWithLocalScope) productPatternMatch(pattern ast.VariousPattern, in typsys.Type) (checked.ProductPatternInfo, *source.Error) {
	var mod = cc.exprContext.ModName
	var lm = cc.exprContext.localBindingMap
	return productPatternMatch(pattern)(in, mod, lm)
}


type ExprChecker func (
	expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (
	*checked.Expr, *typsys.InferringState, *source.Error,
)

func makeExprChecker (
	loc  source.Location,
	k    func(*checkContext) checkResult,
) ExprChecker {
	return func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var cc = makeCheckContext(loc, expected, s, ctx)
		var result = k(cc)
		return result.expr, cc.inferring, result.err
	}
}

func makeExprCheckerWithLocalScope (
	loc  source.Location,
	k    func(*checkContextWithLocalScope) checkResult,
) ExprChecker {
	return func(expected typsys.Type, s *typsys.InferringState, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
		var cc = makeCheckContextWithLocalScope(loc, expected, s, ctx)
		var result = k(cc)
		return result.expr, cc.inferring, result.err
	}
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
	case ast.Record:
		return checkRecord(T)
	case ast.Lambda:
		return checkLambda(T)
	default:
		// TODO
	}
}

func checkPipe(in *checked.Expr, pipe ast.VariousPipe) ExprChecker {
	// TODO
}


