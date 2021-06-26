package checker2

import (
	"encoding/json"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/attr"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/loader"
)


type Function struct {
	attr.FunctionAttrs
	AstNode     *ast.DeclFunction
	Exported    bool
	Name        name.FunctionName
	Signature   FunctionSignature
	Body        FunctionBody
}
type FunctionSignature struct {
	TypeParameters   [] typsys.Parameter
	ImplicitContext  typsys.Record
	InputOutput      typsys.Lambda
}
type FunctionBody interface { functionBody() }

type FunctionRegistry (map[name.Name] ([] *Function))

var bodyInit, bodyWrite = (func() (FunctionBody, func(FunctionBody)(FunctionBody)) {
	return nil, func(body FunctionBody) FunctionBody { return body }
})()

// TODO: note: check conflicts with alias

func registerFunctions (
	mod    *loader.Module,
	reg    FunctionRegistry,
	al     AliasRegistry,
	types  TypeRegistry,
) source.Errors {
	var sb SectionBuffer
	var errs source.Errors
	for _, stmt := range mod.AST.Statements {
		var title, is_title = stmt.Statement.(ast.Title)
		if is_title { sb.SetFrom(title) }
		var decl, is_func_decl = stmt.Statement.(ast.DeclFunction)
		if !(is_func_decl) { continue }
		var _, err = registerFunction(&decl, &sb, mod, reg, al, types)
		source.ErrorsJoin(&errs, err)
	}
	return errs
}

func registerFunction (
	decl *ast.DeclFunction,
	sb *SectionBuffer,
	mod *loader.Module,
	reg FunctionRegistry,
	al     AliasRegistry,
	types TypeRegistry,
) (*Function, *source.Error) {
	var loc = decl.Name.Location
	var func_item_name = ast.Id2String(decl.Name)
	if !(isValidFunctionItemName(func_item_name)) {
		return nil, source.MakeError(loc,
			E_InvalidFunctionName { Name: func_item_name })
	}
	var n = name.MakeName(mod.Name, func_item_name)
	var f = new(Function)
	var existing, _ = reg[n]
	reg[n] = append(existing, f)
	var index = uint(len(existing))
	var func_name = name.FunctionName { Name: n, Index: index }
	var doc = ast.GetDocContent(decl.Docs)
	var section = sb.GetFrom(decl.Location.File)
	var meta attr.FunctionMetadata
	var meta_text = ast.GetMetadataContent(decl.Meta)
	var meta_err = json.Unmarshal(([] byte)(meta_text), &meta)
	if meta_err != nil {
		return nil, source.MakeError(decl.Meta.Location,
			E_InvalidMetadata { Reason: meta_err.Error() })
	}
	var attrs = attr.FunctionAttrs {
		Attrs:    attr.Attrs {
			Location: loc,
			Section:  section,
			Doc:      doc,
		},
		Metadata: meta,
	}
	var params, params_err = (func() ([] typsys.Parameter, *source.Error) {
		var bound_ctx = TypeConsContext {
			Module:   mod,
			TypeReg:  types,
			AliasReg: al,
		}
		var arity = len(decl.Params)
		if arity > MaxTypeParameters {
			return nil, source.MakeError(loc,
				E_TooManyTypeParameters {})
		}
		var params = make([] typsys.Parameter, arity)
		for i, p := range decl.Params {
			var p_name = ast.Id2String(p.Name)
			if !(isValidTypeItemName(p_name)) {
				return nil, source.MakeError(p.Name.Location,
					E_InvalidTypeName { Name: p_name })
			}
			var bound, err = (func() (typsys.Bound, *source.Error) {
				switch bound := p.Bound.Bound.(type) {
				case ast.TypeLowerBound:
					var v, err = newType(bound.Value, bound_ctx)
					if err != nil { return typsys.Bound{}, nil }
					return typsys.Bound { Kind: typsys.InfBound, Value: v }, nil
				case ast.TypeHigherBound:
					var v, err = newType(bound.Value, bound_ctx)
					if err != nil { return typsys.Bound{}, nil }
					return typsys.Bound { Kind: typsys.SupBound, Value: v }, nil
				default:
					return typsys.Bound { Kind:  typsys.NullBound }, nil
				}
			})()
			if err != nil { return nil, err }
			params[i] = typsys.Parameter {
				Name:  p_name,
				Bound: bound,
			}
		}
		return params, nil
	})()
	if params_err != nil { return nil, params_err }
	var ctx = TypeConsContext {
		Module:   mod,
		TypeReg:  types,
		AliasReg: al,
		ParamVec: params,
	}
	var implicit, implicit_err = (func() (typsys.Record, *source.Error) {
		var t, err = newTypeFromRepr(decl.Implicit, ctx)
		if err != nil { return typsys.Record {}, err }
		return t.(*typsys.NestedType).Content.(typsys.Record), nil
	})()
	if implicit_err != nil { return nil, implicit_err }
	var io, io_err = (func() (typsys.Lambda, *source.Error) {
		var t, err = newTypeFromRepr(decl.InOut, ctx)
		if err != nil { return typsys.Lambda {}, err }
		return t.(*typsys.NestedType).Content.(typsys.Lambda), nil
	})()
	if io_err != nil { return nil, io_err }
	*f = Function {
		FunctionAttrs: attrs,
		AstNode:       decl,
		Exported:      decl.Public,
		Name:          func_name,
		Signature:     FunctionSignature {
			TypeParameters:  params,
			ImplicitContext: implicit,
			InputOutput:     io,
		},
		Body: bodyInit,
	}
	return f, nil
}



