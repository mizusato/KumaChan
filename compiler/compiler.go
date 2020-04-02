package compiler

import (
	ch "kumachan/checker"
	. "kumachan/error"
	c "kumachan/runtime/common"
	"kumachan/runtime/lib"
)


type CompiledModule struct {
	Functions   map[string] ([] FuncNode)
	Constants   map[string] FuncNode
	Effects     [] FuncNode
}

type Index  map[string] *CompiledModule


func CompileModule (
	mod       *ch.CheckedModule,
	idx       Index,
	data      *([] c.DataValue),
	closures  *([] FuncNode),
) []*Error {
	var _, exists = idx[mod.Name]
	if exists {
		return nil
	}
	var errs = make([] *Error, 0)
	for _, imported := range mod.Imported {
		var err = CompileModule(imported, idx, data, closures)
		if err != nil {
			errs = append(errs, err...)
		}
	}
	var functions = make(map[string] ([] FuncNode))
	var constants = make(map[string] FuncNode)
	var effects = make([] FuncNode, 0)
	for name, instances := range mod.Functions {
		for _, item := range instances {
			var f_raw, refs, err = CompileFunction (
				item.Body, mod.Name, name, item.Point,
			)
			if err != nil {
				errs = append(errs, err...)
			}
			var f = FuncNodeFrom(f_raw, refs, idx, data, closures)
			var existing, exists = functions[name]
			if exists {
				functions[name] = append(existing, f)
			} else {
				functions[name] = [] FuncNode { f }
			}
		}
	}
	for name, item := range mod.Constants {
		var f, refs, err = CompileConstant (
			item.Value, mod.Name, name, item.Point,
		)
		if err != nil {
			errs = append(errs, err...)
		}
		constants[name] = FuncNodeFrom(f, refs, idx, data, closures)
	}
	for _, item := range mod.Effects {
		var value = ch.ExprExpr(item.Value)
		var f, refs, err = CompileConstant(value, mod.Name, "(do)", item.Point)
		if err != nil {
			errs = append(errs, err...)
		}
		effects = append(effects, FuncNodeFrom(f, refs, idx, data, closures))
	}
	idx[mod.Name] = &CompiledModule {
		Functions: functions,
		Constants: constants,
		Effects:   effects,
	}
	if len(errs) != 0 {
		return errs
	} else {
		return nil
	}
}


func CompileFunction (
	body   ch.ExprLike,
	mod    string,
	name   string,
	point  ErrorPoint,
) (*c.Function, []GlobalRef, []*Error) {
	switch b := body.(type) {
	case ch.ExprNative:
		var native_name = b.Name
		var index, exists = lib.NativeFunctionIndex[native_name]
		var errs []*Error = nil
		if !exists {
			errs = []*Error { &Error {
				Point:    b.Point,
				Concrete: E_NativeFunctionNotFound { native_name },
			} }
		}
		return &c.Function {
			IsNative:    true,
			NativeIndex: index,
			Code:        nil,
			BaseSize:    c.FrameBaseSize {},
			Info:        c.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: nil,
			},
		}, make([]GlobalRef, 0), errs
	case ch.ExprExpr:
		var body_expr = ch.Expr(b)
		var lambda = body_expr.Value.(ch.Lambda)
		var pattern = lambda.Input
		var ctx = MakeContext()
		var scope = ctx.LocalScope
		var buf = MakeCodeBuffer()
		var info = body_expr.Info
		switch p := pattern.Concrete.(type) {
		case ch.TrivialPattern:
			var offset = scope.AddBinding(p.ValueName, p.Point)
			var bind_inst = InstAddBinding(offset)
			buf.Write(CodeFrom(bind_inst, info))
		case ch.TuplePattern:
			BindPatternItems(pattern, p.Items, scope, buf)
		case ch.BundlePattern:
			BindPatternItems(pattern, p.Items, scope, buf)
		default:
			panic("impossible branch")
		}
		var out_code = CompileExpr(lambda.Output, ctx)
		var errs = ctx.LocalScope.CollectUnusedAsErrors()
		buf.Write(out_code)
		var code = buf.Collect()
		var binding_peek = *(scope.BindingPeek)
		if binding_peek >= c.LocalSlotMaxSize {
			panic("maximum quantity of local bindings exceeded")
		}
		return &c.Function {
			IsNative:    false,
			NativeIndex: -1,
			Code:        code.InstSeq,
			BaseSize:    c.FrameBaseSize {
				Context:  0,
				Reserved: c.Long(binding_peek),
			},
			Info:        c.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: code.SourceMap,
			},
		}, *(ctx.GlobalRefs), errs
	default:
		panic("impossible branch")
	}
}


func CompileConstant (
	body   ch.ExprLike,
	mod    string,
	name   string,
	point  ErrorPoint,
) (*c.Function, []GlobalRef, []*Error) {
	switch b := body.(type) {
	case ch.ExprNative:
		var native_name = b.Name
		var index, exists = lib.NativeConstantIndex[native_name]
		var errs []*Error = nil
		if !exists {
			errs = []*Error { &Error {
				Point:    b.Point,
				Concrete: E_NativeConstantNotFound { native_name },
			} }
		}
		return &c.Function {
			IsNative:    true,
			NativeIndex: index,
			Code:        nil,
			BaseSize:    c.FrameBaseSize {},
			Info: c.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: nil,
			},
		}, make([] GlobalRef, 0), errs
	case ch.ExprExpr:
		var body_expr = ch.Expr(b)
		var ctx = MakeContext()
		var code = CompileExpr(body_expr, ctx)
		var errs = ctx.LocalScope.CollectUnusedAsErrors()
		var binding_peek = *(ctx.LocalScope.BindingPeek)
		if binding_peek >= c.LocalSlotMaxSize {
			panic("maximum quantity of local bindings exceeded")
		}
		return &c.Function {
			IsNative:    false,
			NativeIndex: -1,
			Code:        code.InstSeq,
			BaseSize:    c.FrameBaseSize {
				Context:  0,
				Reserved: c.Long(binding_peek),
			},
			Info:        c.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: code.SourceMap,
			},
		}, *(ctx.GlobalRefs), errs
	default:
		panic("impossible branch")
	}
}
