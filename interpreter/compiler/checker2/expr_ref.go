package checker2

import (
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/checker2/checked"
)


type Ref interface { implRef() }

func (FuncRefs) implRef() {}
type FuncRefs struct {
	Functions  [] *Function
}

func (LocalRef) implRef() {}
type LocalRef struct {
	Binding  *checked.LocalBinding
}

func (LocalRefWithFuncRefs) implRef() {}
type LocalRefWithFuncRefs struct {
	LocalRef  LocalRef
	FuncRefs  FuncRefs
}

func checkInlineRef(I ast.InlineRef) ExprChecker {
	return makeExprChecker(I.Location, func(cc *checkContext) checkResult {
		var ref, params, err = cc.resolveInlineRef(I, nil)
		if err != nil { return cc.propagate(err) }
		return cc.forward(checkRef(I.Location, ref, params))
	})
}

func checkName(name_ string, pivot typsys.Type, loc source.Location) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		var n = name.MakeName("", name_)
		var ref, found = cc.resolveName(n, pivot)
		if !(found) {
			return cc.error(
				E_NoSuchBindingOrFunction {
					Name: n.String(),
				})
		}
		return cc.forward(checkRef(loc, ref, nil))
	})
}

func checkRef(loc source.Location, ref Ref, params ([] typsys.Type)) ExprChecker {
	switch R := ref.(type) {
	case LocalRef:
		return checkLocalRef(loc, R, params)
	case FuncRefs:
		return checkFuncRefs(loc, R, params)
	case LocalRefWithFuncRefs:
		// shadowing: global functions are ignored
		return checkLocalRef(loc, R.LocalRef, params)
	default:
		panic("impossible branch")
	}
}

func checkLocalRef(loc source.Location, ref LocalRef, params ([] typsys.Type)) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		if len(params) > 0 {
			return cc.error(
				E_TypeParametersOnLocalBindingRef {})
		}
		return cc.assign(
			ref.Binding.Type,
			checked.LocalRef {
				Binding: ref.Binding,
			})
	})
}

func checkFuncRefs(loc source.Location, refs FuncRefs, params ([] typsys.Type)) ExprChecker {
	return makeExprChecker(loc, func(cc *checkContext) checkResult {
		if len(refs.Functions) == 0 { panic("something went wrong") }
		var ctx = cc.exprContext
		var options = make([] overloadOption, 0)
		var candidates = make([] overloadCandidate, 0)
		for _, f := range refs.Functions {
			var params_def = f.Signature.TypeParameters
			var io = f.Signature.InputOutput
			var result = (func() checkResult {
				if f.AstNode.Kind == ast.FK_Constant {
					if !(typsys.TypeOpEqual(io.Input, typsys.UnitType {})) {
						panic("something went wrong")
					}
					var out_t = io.Output
					return cc.infer(params_def, params, out_t, func(s0 *typsys.InferringState) (checked.ExprContent, *typsys.InferringState, *source.Error) {
						var f_expr, s1, err = makeFuncRef(f, s0, nil, loc, ctx)
						if err != nil { return nil, nil, err }
						// TODO: consider special instruction to call a thunk
						//       (to retrieve saved value)
						return checked.Call {
							Callee:   f_expr,
							Argument: &checked.Expr {
								Type:    typsys.UnitType {},
								Info:    f_expr.Info,
								Content: checked.UnitValue {},
							},
						}, s1, nil
					})
				} else {
					var io_t = &typsys.NestedType { Content: io }
					return cc.infer(params_def, params, io_t, func(s0 *typsys.InferringState) (checked.ExprContent, *typsys.InferringState, *source.Error) {
						var f_expr, s1, err = makeFuncRef(f, s0, nil, loc, ctx)
						if err != nil { return nil, nil, err }
						return f_expr.Content, s1, nil
					})
				}
			})()
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
	})
}

func makeFuncRef(f *Function, s *typsys.InferringState, pivot typsys.Type, loc source.Location, ctx ExprContext) (*checked.Expr, *typsys.InferringState, *source.Error) {
	var record = f.Signature.ImplicitContext
	var values = make([] *checked.Expr, len(record.Fields))
	for i, field := range record.Fields {
		var v, s_, err = checkName(field.Name, pivot, loc)(field.Type, s, ctx)
		if err != nil { return nil, nil, source.MakeError(loc,
			E_ImplicitContextNotFound {
				InnerError: err.Content.DescribeError(),
			}) }
		s = s_
		values[i] = v
	}
	var t = &typsys.NestedType { Content: f.Signature.InputOutput }
	var t_certain, ok = typsys.TypeOpGetCertainType(t, s)
	if !(ok) {
		return nil, nil, source.MakeError(loc,
			E_ExplicitTypeRequired {})
	}
	var content = checked.FuncRef {
		Name:     f.Name,
		Implicit: values,
	}
	var expr = &checked.Expr {
		Type:    t_certain,
		Info:    checked.ExprInfoFrom(loc),
		Content: content,
	}
	return expr, s, nil
}


