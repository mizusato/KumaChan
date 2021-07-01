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
	ModInfo     *ModuleInfo
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
	Expr  *checked.Expr
}
func (NativeBody) functionBody() {}
type NativeBody struct {
	Id  string
}

type FunctionRegistry (map[name.Name] ([] *Function))
type functionList ([] *Function)
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
	mic    ModuleInfoCollection,
	sc     SectionCollection,
	al     AliasRegistry,
	types  TypeRegistry,
) (FunctionRegistry, functionList, source.Errors) {
	var reg = make(FunctionRegistry)
	var mvs = make(ModuleVisitedSet)
	var err = registerFunctions(entry, mic, sc, mvs, reg, al, types)
	if err != nil { return nil, nil, err }
	var functions = make(functionList, 0, len(reg))
	for _, group := range reg {
		for _, f := range group {
			functions = append(functions, f)
		}
	}
	var all_reg = &Registry {
		Aliases:   al,
		Types:     types,
		Functions: reg,
	}
	var step = func(f func(*Function) *source.Error) source.Errors {
		var errs source.Errors
		for _, function := range functions {
			source.ErrorsJoin(&errs, f(function))
		}
		return errs
	}
	var step1_check_alias = step
	var step2_generate_body = step
	{ var err = step1_check_alias(func(f *Function) *source.Error {
		var _, conflict = al[f.Name.Name]
		if conflict {
			return source.MakeError(f.Location,
				E_FunctionConflictWithAlias { Which: f.Name.String() })
		} else {
			return nil
		}
	})
	if err != nil { return nil, nil, err } }
	{ var err = step2_generate_body(func(f *Function) *source.Error {
		var body, err = (func() (FunctionBody, *source.Error) {
			switch body := f.AstNode.Body.Body.(type) {
			case ast.Lambda:
				var ctx = ExprContext {
					Registry:   all_reg,
					ModInfo:    f.ModInfo,
					Inferring:  nil,
				}
				var expected = &typsys.NestedType {
					Content: f.Signature.InputOutput,
				}
				var expr, _, err = CheckLambda(body)(expected, ctx)
				if err != nil { return nil, err }
				return OrdinaryBody { Expr: expr }, nil
			case ast.NativeRef:
				return NativeBody { Id: string(body.Id.Value) }, nil
			default:
				panic("unimplemented")  // TODO: more cases
			}
		})()
		if err != nil { return err }
		f.Body = bodyWrite(body)
		return nil
	})
	if err != nil { return nil, nil, err } }
	return reg, functions, nil
}

func registerFunctions (
	mod    *loader.Module,
	mic    ModuleInfoCollection,
	sc     SectionCollection,
	mvs    ModuleVisitedSet,
	reg    FunctionRegistry,
	al     AliasRegistry,
	types  TypeRegistry,
) source.Errors {
	return TraverseStatements(mod, mic, sc, mvs, func(stmt ast.VariousStatement, sec *source.Section, mi *ModuleInfo) *source.Error {
		var decl, is_func_decl = stmt.Statement.(ast.DeclFunction)
		if !(is_func_decl) { return nil }
		var _, err = registerFunction(&decl, sec, mi, reg, al, types)
		return err
	})
}

func registerFunction (
	decl   *ast.DeclFunction,
	sec    *source.Section,
	mi     *ModuleInfo,
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
	var n = name.MakeName(mi.ModName, func_item_name)
	var f = new(Function)
	var existing, _ = reg[n]
	reg[n] = append(existing, f)
	var index = uint(len(existing))
	var func_name = name.FunctionName { Name: n, Index: index }
	var doc = ast.GetDocContent(decl.Docs)
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
			Section:  sec,
			Doc:      doc,
		},
		Metadata: meta,
	}
	var params, params_err = (func() ([] typsys.Parameter, *source.Error) {
		var bound_ctx = TypeConsContext {
			ModInfo:  mi,
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
		ModInfo:  mi,
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



