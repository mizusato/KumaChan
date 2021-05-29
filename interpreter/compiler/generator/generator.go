package generator

import (
	. "kumachan/standalone/util/error"
	"kumachan/interpreter/base"
	ch "kumachan/interpreter/compiler/checker"
)


type CompiledModule struct {
	Functions   map[string] ([] FuncNode)
	Effects     [] FuncNode
}

type Index  map[string] *CompiledModule


func CompileModule (
	mod       *ch.CheckedModule,
	idx       Index,
	data      *([] base.DataValue),
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
	var effects = make([] FuncNode, 0)
	for name, instances := range mod.Functions {
		for _, item := range instances {
			var f_raw, refs, err = CompileFunction (
				item.Body, item.Implicit, mod.Name, name, item.Point,
			)
			if err != nil { errs = append(errs, err...) }
			var flags = item.FunctionGeneratorFlags
			var kmd_info = item.FunctionKmdInfo
			var f = FuncNodeFrom (
				f_raw, refs, data, closures, flags, kmd_info,
			)
			var existing, exists = functions[name]
			if exists {
				functions[name] = append(existing, f)
			} else {
				functions[name] = [] FuncNode { f }
			}
		}
	}
	for _, item := range mod.Effects {
		var body = ch.BodyThunk {
			Value: item.Value,
		}
		var f, refs, err = CompileFunction (
			body, ([] string {}), mod.Name, "(do)", item.Point,
		)
		if err != nil {
			errs = append(errs, err...)
		}
		effects = append(effects, FuncNodeFrom (
			f, refs, data, closures, __DefaultFlags, __DefaultKmdInfo,
		))
	}
	idx[mod.Name] = &CompiledModule {
		Functions: functions,
		Effects:   effects,
	}
	if len(errs) != 0 {
		return errs
	} else {
		return nil
	}
}


func CompileFunction (
	body   ch.Body,
	imp    [] string,
	mod    string,
	name   string,
	point  ErrorPoint,
) (*base.Function, [] GlobalRef, [] E) {
	var imp_size = uint(len(imp))
	if imp_size > base.ClosureMaxSize {
		panic("something went wrong")
	}
	switch b := body.(type) {
	case ch.BodyGenerated:
		return &base.Function {
			Kind:      base.F_GENERATED,
			Generated: b.Value,
			Code:      nil,
			BaseSize:  base.FrameBaseSize {},
			Info: base.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: nil,
			},
		}, make([] GlobalRef, 0), nil
	case ch.BodyRuntimeGenerated:
		return &base.Function {
			Kind:      base.F_RUNTIME_GENERATED,
			Generated: b.Value,
			Code:      nil,
			BaseSize:  base.FrameBaseSize {},
			Info: base.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: nil,
			},
		}, make([] GlobalRef, 0), nil
	case ch.BodyNative:
		var native_id = b.Name
		return &base.Function {
			Kind:      base.F_NATIVE,
			NativeId:  native_id,
			Generated: nil,
			Code:      nil,
			BaseSize:  base.FrameBaseSize {},
			Info:       base.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: nil,
			},
		}, make([] GlobalRef, 0), nil
	case ch.BodyThunk:
		var ctx = MakeContextWithImplicit(imp)
		var scope = ctx.LocalScope
		var code = CompileExpr(b.Value, ctx)
		var errs = scope.CollectUnusedAsErrors()
		var binding_peek = *(scope.BindingPeek)
		if (imp_size + binding_peek) > base.LocalSlotMaxSize {
			panic("maximum quantity of local bindings exceeded")
		}
		return &base.Function {
			Kind:      base.F_USER,
			Generated: nil,
			Code:      code.InstSeq,
			BaseSize:   base.FrameBaseSize {
				Context:  base.Short(imp_size),
				Reserved: base.Long(binding_peek),
			},
			Info:       base.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: code.SourceMap,
			},
		}, *(ctx.GlobalRefs), errs
	case ch.BodyLambda:
		var ctx = MakeContextWithImplicit(imp)
		var scope = ctx.LocalScope
		var info = b.Info
		var lambda = b.Lambda
		var pattern = lambda.Input
		var buf = MakeCodeBuffer()
		switch p := pattern.Concrete.(type) {
		case ch.TrivialPattern:
			var offset = scope.AddBinding(p.ValueName, p.Point)
			var bind_inst = InstStore(offset)
			buf.Write(CodeFrom(bind_inst, info))
		case ch.TuplePattern:
			BindPatternItems(pattern, p.Items, scope, buf)
		case ch.RecordPattern:
			BindPatternItems(pattern, p.Items, scope, buf)
		default:
			panic("impossible branch")
		}
		var out_code = CompileExpr(lambda.Output, ctx)
		var errs = scope.CollectUnusedAsErrors()
		buf.Write(out_code)
		var code = buf.Collect()
		var binding_peek = *(scope.BindingPeek)
		if (imp_size + binding_peek) > base.LocalSlotMaxSize {
			panic("maximum quantity of local bindings exceeded")
		}
		return &base.Function {
			Kind:      base.F_USER,
			Generated: nil,
			Code:      code.InstSeq,
			BaseSize:   base.FrameBaseSize {
				Context:  base.Short(imp_size),
				Reserved: base.Long(binding_peek),
			},
			Info:       base.FuncInfo {
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

