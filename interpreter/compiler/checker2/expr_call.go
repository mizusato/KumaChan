package checker2

import (
	"fmt"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/checked"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/standalone/util/richtext"
)


type overloadOption struct {
	function  *Function
	value     *checked.Expr
}
type overloadCandidate struct {
	function  *Function
	error     *source.Error
}

func checkCall1(callee ast.Expr, arg ast.Expr, loc source.Location) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		var call_certain = func() checkResult {
			var callee_expr, err1 = cc.checkChildExpr(nil, callee)
			if err1 != nil { return cc.propagate(err1) }
			var io, callable = cc.unboxLambda(callee_expr)
			if !(callable) {
				return cc.error(
					E_TypeNotCallable {})
			}
			var in = io.Input
			var out = io.Output
			var arg_expr, err2 = cc.checkChildExpr(in, arg)
			if err2 != nil { return cc.propagate(err2) }
			return cc.assign(out, checked.Call {
				Callee:   callee_expr,
				Argument: arg_expr,
			})
		}
		var call_overload = func(R FuncRefs) checkResult {
			if len(R.Functions) == 0 { panic("something went wrong") }
			var ctx = cc.exprContext
			var options = make([] overloadOption, 0)
			var candidates = make([] overloadCandidate, 0)
			for _, f := range R.Functions {
				var params = f.Signature.TypeParameters
				var io = f.Signature.InputOutput
				var in_t = io.Input
				var out_t = io.Output
				var result = cc.infer(params, out_t, func(s0 *typsys.InferringState) (checked.ExprContent, *typsys.InferringState, *source.Error) {
					var arg_expr, s1, err1 = check(arg)(in_t, s0, ctx)
					if err1 != nil { return nil, nil, err1 }
					var f_expr, s2, err2 = makeFuncRef(f, s1, nil, loc, ctx)
					if err2 != nil { return nil, nil, err2 }
					return checked.Call {
						Callee:   f_expr,
						Argument: arg_expr,
					}, s2, nil
				})
				if result.err != nil {
					candidates = append(candidates, overloadCandidate {
						function: f,
						error:    result.err,
					})
				} else {
					options = append(options, overloadOption {
						function: f,
						value:    result.expr,
					})
				}
			}
			var expr, err = decideOverload(options, candidates, loc)
			if err != nil { return cc.propagate(err) }
			return cc.confidentlyTrust(expr)
		}
		var ref_node, is_ref = getInlineRef(callee)
		if is_ref {
			var ref, err = cc.resolveInlineRef(ref_node, nil)
			if err != nil { return cc.propagate(err) }
			switch R := ref.(type) {
			case FuncRefs:
				return call_overload(R)
			case LocalRef:
				return call_certain()
			case LocalRefWithFuncRefs:
				return call_certain()
			default:
				panic("impossible branch")
			}
		} else {
			return call_certain()
		}
	})
}

func checkCall2(callee ast.Expr, arg ast.Expr, pivot *checked.Expr, loc source.Location) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		if pivot == nil { panic("something went wrong") }
		var ref, is_ref = getInlineRef(callee)
		if is_ref {
			// TODO (use pivot.Type to lookup name)
		} else {
			var callee_expr, err1 = cc.checkChildExpr(nil, callee)
			if err1 != nil { return cc.propagate(err1) }
			var io, callable = cc.unboxLambda(callee_expr)
			if !(callable) {
				return cc.error(
					E_TypeNotCallable {})
			}
			var in = io.Input
			var out = io.Output
			var tuple_exp, ok = getTuple(cc.expected)
			if !(ok) {
				// TODO
			}
			if len(tuple_exp.Elements) != 2 {
				// TODO
			}
			var pivot_exp = tuple_exp.Elements[0]
			var arg_exp = tuple_exp.Elements[1]
			var err2 = cc.assignType(pivot_exp, pivot.Type)
			if err2 != nil { return cc.propagate(err2) }
			var arg_expr, err3 = cc.checkChildExpr(arg_exp, arg)
			if err3 != nil { return cc.propagate(err3) }
			var pair_t = &typsys.NestedType {
				Content: typsys.Tuple { Elements: [] typsys.Type {
					pivot.Type,
					arg_expr.Type,
				} },
			}
			var pair = &checked.Expr {
				Type:    pair_t,
				Info:    checked.ExprInfoFrom(loc),
				Content: checked.Tuple { Elements: [] *checked.Expr {
					pivot,
					arg_expr,
				} },
			}
			var err4 = cc.assignType(in, pair_t)
			if err4 != nil { return cc.propagate(err4) }
			return cc.assign(out, checked.Call {
				Callee:   callee_expr,
				Argument: pair,
			})
		}
	})
}

func decideOverload (
	options     ([] overloadOption),
	candidates  ([] overloadCandidate),
	location    source.Location,
) (*checked.Expr, *source.Error) {
	if len(options) == 0 {
		return nil, source.MakeError(location,
			E_InvalidFunctionCall {
				Candidates: describeOverloadCandidates(candidates),
			})
	} else if len(options) == 1 {
		return options[0].value, nil
	} else {
		var output_types = make([] typsys.Type, len(options))
		for i, option := range options {
			output_types[i] = option.function.Signature.InputOutput.Output
		}
		var min_index, found = findUniqueMinimumNamedType(output_types)
		if !(found) {
			return nil, source.MakeError(location,
				E_AmbiguousFunctionCall {
					Options: describeOverloadOptions(options),
				})
		}
		return options[min_index].value, nil
	}
}
func findUniqueMinimumNamedType(types ([] typsys.Type)) (int, bool) {
	if len(types) < 2 {
		panic("invalid argument")
	}
	var defs = make([] *typsys.TypeDef, len(types))
	for i, t := range types {
		var def, ok = typeGetRefDef(t)
		if !(ok) { return -1, false }
		defs[i] = def
	}
	var min = defs[0]
	var min_index = 0
	for i, d := range defs[1:] {
		if typeDefSubtypingOrderLessThanOperator(d, min) {
			min = d
			min_index = i
		} else if typeDefSubtypingOrderLessThanOperator(min, d) {
			// do nothing
		} else if min == d {
			// do nothing
		} else {
			// not comparable
			return -1, false
		}
	}
	for i, d := range defs {
		if i != min_index && d == min {
			// not unique
			return -1, false
		}
	}
	return min_index, true
}
func describeOverloadOptions(options ([] overloadOption)) ([] string) {
	var desc = make([] string, 0)
	for i, option := range options {
		desc[i] = describeFunctionSignature(option.function.Signature)
	}
	return desc
}
func describeOverloadCandidates(candidates ([] overloadCandidate)) ([] string) {
	// TODO: structure instead of string
	var desc = make([] string, 0)
	for i, candidate := range candidates {
		var sig =  describeFunctionSignature(candidate.function.Signature)
		var info_rich = candidate.error.Content.DescribeError().ToText()
		var info = info_rich.RenderLinear(richtext.RenderOptionsLinear {})
		desc[i] = fmt.Sprintf("%s %s", sig, info)
	}
	return desc
}

func typeGetRefDef(t typsys.Type) (*typsys.TypeDef, bool) {
	var nested, is_nested = t.(*typsys.NestedType)
	if !(is_nested) { return nil, false }
	var ref, is_ref = nested.Content.(typsys.Ref)
	if !(is_ref) { return nil, false }
	return ref.Def, true
}
func typeDefSubtypingOrderLessThanOperator(u *typsys.TypeDef, v *typsys.TypeDef) bool {
	var box, is_box = u.Content.(*typsys.Box)
	if is_box {
		var u_, ok = typeGetRefDef(box.InnerType)
		if !(ok) { return false }
		if u_ == v {
			return true
		} else {
			return typeDefSubtypingOrderLessThanOperator(u_, v)
		}
	} else {
		return false
	}
}

func getInlineRef(expr ast.Expr) (ast.InlineRef, bool) {
	if expr.Pipeline == nil {
		var inline_ref, ok = expr.Term.Term.(ast.InlineRef)
		return inline_ref, ok
	} else {
		return ast.InlineRef {}, false
	}
}


