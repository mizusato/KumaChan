package checker2

import (
	"encoding/json"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/attr"
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/compiler/loader"
	"kumachan/interpreter/compiler/checker2/checked"
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
func (OrdinaryBody) functionBody() {}
type OrdinaryBody struct {
	Expr  checked.Expr
}
func (NativeBody) functionBody() {}
type NativeBody struct {
	Id  string
}

type FunctionRegistry (map[name.Name] ([] *Function))
type functionList ([] functionWithModule)
type functionWithModule struct {
	*Function
	Module  *loader.Module
}
func (l functionList) Less(i int, j int) bool {
	var u = l[i].Name
	var v = l[j].Name
	if u.ModuleName < v.ModuleName {
		return true
	} else if u.ModuleName == v.ModuleName {
		return (u.ItemName < v.ItemName)
	} else {
		return false
	}
}
func (l functionList) Len() int {
	return len(l)
}
func (l functionList) Swap(i int, j int) {
	var I = &(l[i])
	var J = &(l[j])
	var t = *I
	*I = *J
	*J = t
}

var bodyInit, bodyWrite = (func() (FunctionBody, func(FunctionBody)(FunctionBody)) {
	return nil, func(body FunctionBody) FunctionBody { return body }
})()

func collectFunctions (
	entry  *loader.Module,
	idx    loader.Index,
	al     AliasRegistry,
	types  TypeRegistry,
) (FunctionRegistry, source.Errors) {
	var reg = make(FunctionRegistry)
	var err = registerFunctions(entry, reg, al, types)
	if err != nil { return nil, err }
	var functions = make(functionList, 0, len(reg))
	for _, group := range reg {
		for _, f := range group {
			var mod, exists = idx[f.Name.ModuleName]
			if !(exists) { panic("something went wrong") }
			functions = append(functions, functionWithModule {
				Function: f,
				Module:   mod,
			})
		}
	}
	var all_reg = &Registry {
		Aliases:   al,
		Types:     types,
		Functions: reg,
	}
	var step = func(f func(functionWithModule) *source.Error) source.Errors {
		var errs source.Errors
		for _, function := range functions {
			source.ErrorsJoin(&errs, f(function))
		}
		return errs
	}
	var step1_check_alias = step
	var step2_generate_body = step
	{ var err = step1_check_alias(func(f functionWithModule) *source.Error {
		var _, conflict = al[f.Name.Name]
		if conflict {
			return source.MakeError(f.Location,
				E_FunctionConflictWithAlias { Which: f.Name.String() })
		} else {
			return nil
		}
	})
	if err != nil { return nil, err } }
	{ var err = step2_generate_body(func(f functionWithModule) *source.Error {
		var body, err = (func() (FunctionBody, *source.Error) {
			switch ast_body := f.AstNode.Body.Body.(type) {
			case ast.Lambda:
				var ctx = ExprContext {
					Registry:   all_reg,
					Module:     f.Module,
					Parameters: f.Signature.TypeParameters,
					Inferring:  nil,
				}
				var expected = &typsys.NestedType {
					Content: f.Signature.InputOutput,
				}
				var expr, _, err = CheckLambda(ast_body)(expected, ctx)
				if err != nil { return nil, err }
				return OrdinaryBody { Expr: expr }, nil
			case ast.NativeRef:
				return NativeBody { Id: string(ast_body.Id.Value) }, nil
			default:
				panic("unimplemented")  // TODO: more cases
			}
		})()
		if err != nil { return err }
		f.Body = body
		return nil
	})
	if err != nil { return nil, err } }
	return reg, nil
}

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
	decl   *ast.DeclFunction,
	sb     *SectionBuffer,
	mod    *loader.Module,
	reg    FunctionRegistry,
	al     AliasRegistry,
	types  TypeRegistry,
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



