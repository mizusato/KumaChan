package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/checked"
	"kumachan/interpreter/compiler/checker2/typsys"
)


type overloadOption struct {
	function  *Function
	value     *checked.Expr
}
type overloadCandidate struct {
	function  *Function
	error     *source.Error
}

type ArgGetter  func(in typsys.Type) (*checked.Expr, *source.Error)

func callLocalRefWithArgGetter (
	cc       *checkContext,
	ref      LocalRef,
	params   [] typsys.Type,
	ref_loc  source.Location,
	get_arg  ArgGetter,
) checkResult {
	if len(params) > 0 {
		return cc.propagate(source.MakeError(ref_loc,
			E_TypeParametersOnLocalBindingRef {}))
	}
	var ref_type = ref.Binding.Type
	var io, callable = cc.unboxLambdaFromType(ref_type)
	if !(callable) {
		return cc.propagate(source.MakeError(ref_loc,
			E_TypeNotCallable {
				TypeName: typsys.DescribeType(ref_type, nil),
			}))
	}
	var in = io.Input
	var out = io.Output
	var ref_expr = &checked.Expr {
		Type:    ref_type,
		Info:    checked.ExprInfoFrom(ref_loc),
		Content: checked.LocalRef { Binding: ref.Binding },
	}
	var arg_expr, err2 = get_arg(in)
	if err2 != nil { return cc.propagate(err2) }
	return cc.assign(out, checked.Call {
		Callee:   ref_expr,
		Argument: arg_expr,
	})
}

func callExprWithArgGetter (
	cc       *checkContext,
	callee   ast.Expr,
	get_arg  ArgGetter,
) checkResult {
	var callee_expr, err1 = cc.checkChildExpr(nil, callee)
	if err1 != nil { return cc.propagate(err1) }
	var io, callable = cc.unboxLambda(callee_expr)
	if !(callable) {
		return cc.propagate(source.MakeError(callee.Location,
			E_TypeNotCallable {
				TypeName: typsys.DescribeType(callee_expr.Type, nil),
			}))
	}
	var in = io.Input
	var out = io.Output
	var arg_expr, err2 = get_arg(in)
	if err2 != nil { return cc.propagate(err2) }
	return cc.assign(out, checked.Call {
		Callee:   callee_expr,
		Argument: arg_expr,
	})
}

func callExpr(cc *checkContext, callee ast.Expr, arg ast.Expr) checkResult {
	return callExprWithArgGetter(cc, callee, func(in typsys.Type) (*checked.Expr, *source.Error) {
		return cc.checkChildExpr(in, arg)
	})
}

func callExprWithCheckedArg(cc *checkContext, callee ast.Expr, arg *checked.Expr) checkResult {
	return callExprWithArgGetter(cc, callee, func(in typsys.Type) (*checked.Expr, *source.Error) {
		var err = cc.assignType(in, arg.Type)
		if err != nil { return nil, err }
		return arg, nil
	})
}

func callLocalRef(cc *checkContext, ref LocalRef, params ([] typsys.Type), ref_loc source.Location, arg ast.Expr) checkResult {
	return callLocalRefWithArgGetter(cc, ref, params, ref_loc, func(in typsys.Type) (*checked.Expr, *source.Error) {
		return cc.checkChildExpr(in, arg)
	})
}

func callLocalRefWithCheckedArg(cc *checkContext, ref LocalRef, params ([] typsys.Type), ref_loc source.Location, arg *checked.Expr) checkResult {
	return callLocalRefWithArgGetter(cc, ref, params, ref_loc, func(in typsys.Type) (*checked.Expr, *source.Error) {
		var err = cc.assignType(in, arg.Type)
		if err != nil { return nil, err }
		return arg, nil
	})
}

func callFuncRefs (
	cc        *checkContext,
	callee    FuncRefs,
	params    [] typsys.Type,
	arg_expr  *checked.Expr,
	pivot     typsys.Type,
) checkResult {
	if len(callee.Functions) == 0 { panic("something went wrong") }
	var ctx = cc.exprContext
	var loc = cc.location
	var options = make([] overloadOption, 0)
	var candidates = make([] overloadCandidate, 0)
	for _, f := range callee.Functions {
		var params_def = f.Signature.TypeParameters
		var io = f.Signature.InputOutput
		var in_t = io.Input
		var out_t = io.Output
		var result = cc.infer(params_def, params, out_t, func(s0 *typsys.InferringState) (checked.ExprContent, *typsys.InferringState, *source.Error) {
			var a = typsys.MakeAssignContext(ctx.ModName, s0)
			var ok, s1 = typsys.Assign(in_t, arg_expr.Type, a)
			if !(ok) { return nil, nil, source.MakeError(arg_expr.Info.Location,
				E_NotAssignable {
					From: typsys.DescribeType(arg_expr.Type, nil),
					To:   typsys.DescribeType(in_t, s0),
				}) }
			var f_expr, s2, err = makeFuncRef(f, s1, pivot, loc, ctx)
			if err != nil { return nil, nil, err }
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

func checkCall1(callee ast.Expr, arg ast.Expr, loc source.Location) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		var ref_node, is_ref = getInlineRef(callee)
		if is_ref {
			var ref, params, err = cc.resolveInlineRef(ref_node, nil)
			if err != nil { return cc.propagate(err) }
			var ref_loc = ref_node.Location
			switch R := ref.(type) {
			case FuncRefs:
				var arg_expr, arg_err = cc.checkChildExpr(nil, arg)
				if arg_err != nil { return cc.propagate(arg_err) }
				return callFuncRefs(cc, R, params, arg_expr, nil)
			case LocalRef:
				return callLocalRef(cc, R, params, ref_loc, arg)
			case LocalRefWithFuncRefs:
				// a local binding shadows global functions in a prefix call
				return callLocalRef(cc, R.LocalRef, params, ref_loc, arg)
			default:
				panic("impossible branch")
			}
		} else {
			return callExpr(cc, callee, arg)
		}
	})
}

func checkCall2(callee ast.Expr, arg ast.Expr, pivot *checked.Expr, loc source.Location) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		if pivot == nil { panic("something went wrong") }
		var arg_expr, err = cc.checkChildExpr(nil, arg)
		if err != nil { return cc.propagate(err) }
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
		var ref_node, is_ref = getInlineRef(callee)
		if is_ref {
			var ref, params, err = cc.resolveInlineRef(ref_node, pivot.Type)
			if err != nil { return cc.propagate(err) }
			var ref_loc = ref_node.Location
			switch R := ref.(type) {
			case FuncRefs:
				return callFuncRefs(cc, R, params, pair, pivot.Type)
			case LocalRef:
				return callLocalRefWithCheckedArg(cc, R, params, ref_loc, pair)
			case LocalRefWithFuncRefs:
				// global functions are preferred in a pipeline or infix call
				return callFuncRefs(cc, R.FuncRefs, params, pair, pivot.Type)
			default:
				panic("something went wrong")
			}
		} else {
			return callExprWithCheckedArg(cc, callee, pair)
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
			E_InvalidFunctionUsage {
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
				E_AmbiguousFunctionUsage {
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
func describeOverloadCandidates(candidates ([] overloadCandidate)) ([] OverloadCandidateDescription) {
	var desc = make([] OverloadCandidateDescription, 0)
	for i, candidate := range candidates {
		desc[i] = OverloadCandidateDescription {
			Signature: describeFunctionSignature(candidate.function.Signature),
			Error:     candidate.error.Content.DescribeError(),
		}
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


