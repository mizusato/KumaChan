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
		return cc.forward(checkRef(ref, params))
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
		return cc.forward(checkRef(ref, nil))
	})
}

func checkRef(ref Ref, params ([] typsys.Type)) ExprChecker {
	// TODO
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


