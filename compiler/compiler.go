package compiler

import (
	. "kumachan/error"
	ch "kumachan/checker"
	c "kumachan/runtime/common"
	"kumachan/runtime/lib"
	"kumachan/kmd"
	"fmt"
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
) [] E {
	var _, exists = idx[mod.Name]
	if exists {
		return nil
	}
	var errs = make([] E, 0)
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
				item.Body, item.Implicit, mod.Name, name, item.Point,
			)
			if err != nil { errs = append(errs, err...) }
			var kmd_info = item.FunctionKmdInfo
			var f = FuncNodeFrom(f_raw, refs, data, closures, kmd_info)
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
		constants[name] = FuncNodeFrom(f, refs, data, closures, __NoKmdInfo)
	}
	for _, item := range mod.Effects {
		var value = ch.ExprExpr(item.Value)
		var f, refs, err = CompileConstant(value, mod.Name, "(do)", item.Point)
		if err != nil {
			errs = append(errs, err...)
		}
		effects = append(effects, FuncNodeFrom(f, refs, data, closures, __NoKmdInfo))
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
	imp    [] string,
	mod    string,
	name   string,
	point  ErrorPoint,
) (*c.Function, [] GlobalRef, [] E) {
	switch b := body.(type) {
	case ch.ExprKmdApi:
		var f c.NativeFunctionValue
		switch id := b.Id.(type) {
		case kmd.SerializerId:
			f = func(arg c.Value, h c.InteropContext) c.Value {
				var t = h.KmdGetTypeFromId(id.TypeId)
				var binary, err = h.KmdSerialize(arg, t)
				if err != nil {
					var wrapped = fmt.Errorf("serialiation error: %w", err)
					panic(wrapped)
				}
				return binary
			}
		case kmd.DeserializerId:
			f = func(arg c.Value, h c.InteropContext) c.Value {
				var t = h.KmdGetTypeFromId(id.TypeId)
				var obj, err = h.KmdDeserialize(arg.([] byte), t)
				if err != nil { return c.Ng(err) }
				return c.Ok(obj)
			}
		default:
			panic("impossible branch")
		}
		return &c.Function {
			Kind:        c.F_PREDEFINED,
			NativeIndex: ^uint(0),
			Predefined:  f,
			Code:        nil,
			BaseSize:    c.FrameBaseSize {},
			Info: c.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: nil,
			},
		}, make([] GlobalRef, 0), nil
	case ch.ExprPredefinedValue:
		if len(imp) > 0 { panic("something went wrong") }
		return &c.Function {
			Kind:        c.F_PREDEFINED,
			NativeIndex: ^uint(0),
			Predefined:  b.Value,
			Code:        nil,
			BaseSize:    c.FrameBaseSize {},
			Info: c.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: nil,
			},
		}, make([] GlobalRef, 0), nil
	case ch.ExprNative:
		if len(imp) > 0 { panic("something went wrong") }
		var native_name = b.Name
		var index, exists = lib.NativeFunctionIndex[native_name]
		var errs [] E = nil
		if !exists {
			errs = [] E { &Error {
				Point:    b.Point,
				Concrete: E_NativeFunctionNotFound { native_name },
			} }
		}
		return &c.Function {
			Kind:        c.F_NATIVE,
			NativeIndex: index,
			Predefined:  nil,
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
		var context_size = uint(len(imp))
		var ctx = MakeContextWithImplicit(imp)
		var scope = ctx.LocalScope
		var buf = MakeCodeBuffer()
		var info = body_expr.Info
		switch p := pattern.Concrete.(type) {
		case ch.TrivialPattern:
			var offset = scope.AddBinding(p.ValueName, p.Point)
			var bind_inst = InstStore(offset)
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
		if (context_size + binding_peek) > c.LocalSlotMaxSize {
			panic("maximum quantity of local bindings exceeded")
		}
		return &c.Function {
			Kind:        c.F_USER,
			NativeIndex: ^uint(0),
			Predefined:  nil,
			Code:        code.InstSeq,
			BaseSize:    c.FrameBaseSize {
				Context:  c.Short(context_size),
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
) (*c.Function, [] GlobalRef, [] E) {
	switch b := body.(type) {
	case ch.ExprPredefinedValue:
		return &c.Function {
			Kind:        c.F_PREDEFINED,
			NativeIndex: ^uint(0),
			Predefined:  b.Value,
			Code:        nil,
			BaseSize:    c.FrameBaseSize {},
			Info: c.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: nil,
			},
		}, make([] GlobalRef, 0), nil
	case ch.ExprNative:
		var native_name = b.Name
		var index, exists = lib.NativeConstantIndex[native_name]
		var errs [] E = nil
		if !exists {
			errs = [] E { &Error {
				Point:    b.Point,
				Concrete: E_NativeConstantNotFound { native_name },
			} }
		}
		return &c.Function {
			Kind:        c.F_NATIVE,
			NativeIndex: index,
			Predefined:  nil,
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
		if binding_peek > c.LocalSlotMaxSize {
			panic("maximum quantity of local bindings exceeded")
		}
		return &c.Function {
			Kind:        c.F_USER,
			NativeIndex: ^uint(0),
			Predefined:  nil,
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
