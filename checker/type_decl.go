package checker

import (
	"strings"
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/runtime/common"
	"kumachan/parser/ast"
)


var __ReservedTypeNames = [...]string {
	IgnoreMark, UnitAlias, WildcardRhsTypeDesc,
}
func IsReservedTypeName(name string) bool {
	for _, reserved := range __ReservedTypeNames {
		if name == reserved {
			return true
		}
	}
	return false
}

// Final Registry of Types
type TypeRegistry  map[loader.Symbol] *GenericType

// Intermediate Registry of Types, Used When Defining Types
type RawTypeRegistry struct {
	// a map from symbol to type declaration (AST node)
	DeclMap        map[loader.Symbol] ast.DeclType
	// a map from symbol to type parameters
	ParamsMap      map[loader.Symbol] ([] TypeParam)
	// a map from case types to their corresponding union information
	CaseInfoMap    map[loader.Symbol] CaseInfo
	// a context value to track all visited modules
	// (same module may appear many times when walking through the hierarchy)
	VisitedMod     map[string] bool
}
type CaseInfo struct {
	IsCaseType  bool
	UnionName   loader.Symbol
	UnionArity  uint
	CaseIndex   uint
	CaseParams  [] uint
}
func MakeRawTypeRegistry() RawTypeRegistry {
	return RawTypeRegistry {
		DeclMap:       make(map[loader.Symbol] ast.DeclType),
		ParamsMap:     make(map[loader.Symbol] ([] TypeParam)),
		CaseInfoMap:   make(map[loader.Symbol] CaseInfo),
		VisitedMod:    make(map[string] bool),
	}
}
func RegisterRawTypes (mod *loader.Module, raw RawTypeRegistry) *TypeDeclError {
	/**
	 *  Input: a loaded module and a raw registry
	 *  Output: error or nil
	 *  Effect: add all types declared in the module
	 *          and all modules imported by the module to the registry
	 */
	// 1. Check if the module was visited, if visited, do nothing
	var mod_name = mod.Name
	var _, visited = raw.VisitedMod[mod_name]
	if visited { return nil }
	raw.VisitedMod[mod_name] = true
	// 2. Extract all type declarations in the module,
	//    and record the root union types of all case types
	var decls = make([]ast.DeclType, 0)
	var params_map = make(map[uint] ([] TypeParam))
	var case_parent_map = make(map[uint]uint)
	var case_index_map = make(map[uint]uint)
	for _, stmt := range mod.Node.Statements {
		switch s := stmt.Statement.(type) {
		case ast.DeclType:
			var i = uint(len(decls))
			decls = append(decls, s)
			var params, err, err_node = TypeParams(s.Params)
			if err != nil { return &TypeDeclError {
				Point:    ErrorPointFrom(err_node),
				Concrete: *err,
			} }
			params_map[i] = params
			switch u := s.TypeValue.TypeValue.(type) {
			case ast.UnionType:
				for case_index, case_decl := range u.Cases {
					var case_i = uint(len(decls))
					decls = append(decls, case_decl)
					var params, err, err_node = TypeParams(case_decl.Params)
					if err != nil { return &TypeDeclError {
						Point:    ErrorPointFrom(err_node),
						Concrete: *err,
					} }
					params_map[case_i] = params
					case_index_map[case_i] = uint(case_index)
					case_parent_map[case_i] = i
				}
			}
		}
	}
	// 3. Go through all type declarations
	for i, d := range decls {
		// 3.1. Get the symbol of the declared type
		var type_sym = mod.SymbolFromDeclName(d.Name)
		// 3.2. Check if the symbol name is valid
		var sym_name = type_sym.SymbolName
		if strings.HasSuffix(sym_name, ForceExactSuffix) ||
			strings.HasPrefix(sym_name, CovariantPrefix) ||
			strings.HasPrefix(sym_name, ContravariantPrefix) ||
			IsReservedTypeName(sym_name) {
			return &TypeDeclError {
				Point:    ErrorPointFrom(d.Name.Node),
				Concrete: E_InvalidTypeName { sym_name },
			}
		}
		// 3.3. Check if the symbol is used
		var _, exists = raw.DeclMap[type_sym]
		if exists {
			// 3.3.1. If used, throw an error
			return &TypeDeclError {
				Point:    ErrorPointFrom(d.Name.Node),
				Concrete: E_DuplicateTypeDecl {
					TypeName: type_sym,
				},
			}
		}
		// 3.3.2. If not, register the declaration node to DeclMap
		//        and update ParamsMap and CaseIndexMap if necessary.
		//        If invalid parameters were declared on a case type,
		//        throw an error.
		raw.DeclMap[type_sym] = d
		var params = params_map[uint(i)]
		raw.ParamsMap[type_sym] = params
		var case_index, is_case_type = case_index_map[uint(i)]
		if is_case_type {
			var parent_i = case_parent_map[uint(i)]
			var parent = decls[parent_i]
			var parent_name = mod.SymbolFromDeclName(parent.Name)
			var parent_params = params_map[parent_i]
			var mapping = make([] uint, len(d.Params))
			for p, param := range params {
				var exists = false
				var corresponding = ^(uint(0))
				for parent_p, parent_param := range parent_params {
					if param.Name == parent_param.Name {
						exists = true
						corresponding = uint(parent_p)
					}
				}
				if !exists { return &TypeDeclError {
					Point:    ErrorPointFrom(d.Params[p].Node),
					Concrete: E_InvalidCaseTypeParam { param.Name },
				} }
				mapping[p] = corresponding
			}
			raw.CaseInfoMap[type_sym] = CaseInfo {
				IsCaseType: true,
				UnionName:  parent_name,
				UnionArity: uint(len(parent.Params)),
				CaseIndex:  case_index,
				CaseParams: mapping,
			}
		} // if is_case_type
	}
	// 4. Go through all imported modules, call self recursively
	for _, imported := range mod.ImpMap {
		// 4.1. If an error occurred, bubble it
		var err = RegisterRawTypes(imported, raw)
		if err != nil {
			return err
		}
	}
	// 5. Return nil
	return nil
}

type TypeContext struct {
	Module  *loader.Module
	Params  [] TypeParam
	Ireg    AbstractRegistry
}
type AbstractRegistry interface {
	LookupParams(loader.Symbol) ([] TypeParam, bool)
}
func (raw RawTypeRegistry) LookupParams(name loader.Symbol) ([] TypeParam, bool) {
	var params, exists = raw.ParamsMap[name]
	return params, exists
}
func (reg TypeRegistry) LookupParams(name loader.Symbol) ([] TypeParam, bool) {
	var t, exists = reg[name]
	if exists {
		return t.Params, true
	} else {
		return nil, false
	}
}
func (ctx TypeContext) Arity() uint {
	return uint(len(ctx.Params))
}
func (ctx TypeContext) DeduceVariance(params_v ([] TypeVariance), args_v ([][] TypeVariance)) ([] TypeVariance) {
	var ctx_arity = ctx.Arity()
	var n = uint(len(params_v))
	var result = make([] TypeVariance, ctx_arity)
	for i := uint(0); i < ctx_arity; i += 1 {
		var v = Bivariant
		for j := uint(0); j < n; j += 1 {
			v = CombineVariance(v, ApplyVariance(params_v[j], args_v[j][i]))
		}
		result[i] = TypeVariance(v)
	}
	return result
}
func TypeParams(raw_params ([] ast.Identifier)) ([] TypeParam, *E_InvalidTypeName, ast.Node) {
	var params = make([] TypeParam, len(raw_params))
	for p, raw_param := range raw_params {
		var raw_name = loader.Id2String(raw_param)
		var name = raw_name
		var v = Invariant
		if len(raw_name) > len(CovariantPrefix) &&
			strings.HasPrefix(raw_name, CovariantPrefix) {
			name = strings.TrimPrefix(raw_name, CovariantPrefix)
			v = Covariant
		} else if len(raw_name) > len(ContravariantPrefix) &&
			strings.HasPrefix(raw_name, ContravariantPrefix) {
			name = strings.TrimPrefix(raw_name, ContravariantPrefix)
			v = Contravariant
		}
		if strings.HasSuffix(name, ForceExactSuffix) ||
			IsReservedTypeName(name) {
			return nil, &E_InvalidTypeName {name}, raw_param.Node
		}
		params[p] = TypeParam {
			Name:     name,
			Variance: v,
		}
	}
	return params, nil, ast.Node{}
}
func MatchVariance(declared ([] TypeParam), deduced ([] TypeVariance)) (bool, ([] string)) {
	var bad_params = make([] string, 0)
	for i, _ := range declared {
		var v = deduced[i]
		var name = declared[i].Name
		var declared_v = declared[i].Variance
		var bad = false
		if declared_v == Covariant {
			if !(v == Covariant || v == Bivariant) {
				bad = true
			}
		} else if declared_v == Contravariant {
			if !(v == Contravariant || v == Bivariant) {
				bad = true
			}
		} else if declared_v == Bivariant {
			if v != Bivariant {
				bad = true
			}
		}
		if bad {
			bad_params = append(bad_params, name)
		}
	}
	if len(bad_params) > 0 {
		return false, bad_params
	} else {
		return true, nil
	}
}

func RegisterTypes (entry *loader.Module, idx loader.Index) (TypeRegistry, *TypeDeclError) {
	// 1. Build a raw registry
	var raw = MakeRawTypeRegistry()
	var err = RegisterRawTypes(entry, raw)
	if err != nil { return nil, err }
	// 2. Create a empty TypeRegistry
	var reg = make(TypeRegistry)
	// 3. Go through all types in the raw registry
	for name, t := range raw.DeclMap {
		var mod, mod_exists = idx[name.ModuleName]
		if !mod_exists { panic("mod " + name.ModuleName + " should exist") }
		// 3.1. Get parameters
		var params = raw.ParamsMap[name]
		// 3.2. Construct a TypeContext and pass it to TypeValFrom()
		//      to generate a TypeVal and bubble errors
		var val, err = TypeValFrom(t.TypeValue.TypeValue, TypeContext {
			Module: mod,
			Params: params,
			Ireg:   raw,
		}, raw)
		if err != nil { return nil, &TypeDeclError {
			Point:    err.Point,
			Concrete: E_InvalidTypeDecl {
				TypeName: name,
				Detail:   err,
			},
		} }
		// 3.3. Use the generated TypeVal to construct a GenericType
		//      and register it to the TypeRegistry
		reg[name] = &GenericType {
			Params:   params,
			Value:    val,
			Node:     t.Node,
			CaseInfo: raw.CaseInfoMap[name],
		}
	}
	// 4. Validate boxed types
	var check_cycle func(loader.Symbol, Boxed, []loader.Symbol) *TypeDeclError
	check_cycle = func (
		name loader.Symbol, t Boxed, path []loader.Symbol,
	) *TypeDeclError {
		for _, visited := range path {
			if visited == name {
				var node = reg[path[0]].Node
				var point = ErrorPointFrom(node)
				return &TypeDeclError {
					Point:    point,
					Concrete: E_TypeCircularDependency { path },
				}
			}
		}
		var named, is_named = t.InnerType.(NamedType)
		if is_named {
			var inner_name = named.Name
			var g = reg[inner_name]
			var inner_boxed, is_boxed = g.Value.(Boxed)
			if is_boxed {
				var inner_path = append(path, name)
				return check_cycle(inner_name, inner_boxed, inner_path)
			}
		}
		return nil
	}
	for name, g := range reg {
		var boxed, is_boxed = g.Value.(Boxed)
		if is_boxed {
			var err = check_cycle(name, boxed, make([]loader.Symbol, 0))
			if err != nil { return nil, err }
		}
	}
	// 5. Return the TypeRegistry
	return reg, nil
}

/* Transform: ast.TypeValue -> checker.TypeVal */
func TypeValFrom(tv ast.TypeValue, ctx TypeContext, raw RawTypeRegistry) (TypeVal, *TypeError) {
	switch val := tv.(type) {
	case ast.UnionType:
		var count = uint(len(val.Cases))
		var max = uint(common.SumMaxBranches)
		if count > max {
			return nil, &TypeError {
				Point:    ErrorPointFrom(val.Node),
				Concrete: E_TooManyUnionItems {
					Defined: count,
					Limit:   max,
				},
			}
		}
		var case_types = make([] CaseType, count)
		for i, case_decl := range val.Cases {
			var sym = ctx.Module.SymbolFromDeclName(case_decl.Name)
			case_types[i] = CaseType {
				Name:   sym,
				Params: raw.CaseInfoMap[sym].CaseParams,
			}
		}
		// TODO: check variance
		return Union {
			CaseTypes: case_types,
		}, nil
	case ast.BoxedType:
		var inner, specified = val.Inner.(ast.VariousType)
		if specified {
			var inner_type, inner_v, err = TypeFrom(inner.Type, ctx)
			if err != nil {
				return nil, err
			}
			var v_ok, bad_params = MatchVariance(ctx.Params, inner_v)
			if !(v_ok) { return nil, &TypeError {
				Point:    ErrorPointFrom(val.Node),
				Concrete: E_BoxedBadVariance { bad_params },
			} }
			return Boxed {
				InnerType: inner_type,
				AsIs:      val.AsIs,
				Protected: val.Protected,
				Opaque:    val.Opaque,
			}, nil
		} else {
			return Boxed {
				InnerType: AnonymousType { Unit{} },
				AsIs:      val.AsIs,
				Protected: val.Protected,
				Opaque:    val.Opaque,
			}, nil
		}
	case ast.NativeType:
		return Native{}, nil
	default:
		panic("impossible branch")
	}
}

/* Transform: ast.Type -> checker.Type */
func TypeFrom(type_ ast.Type, ctx TypeContext) (Type, ([] TypeVariance), *TypeError) {
	switch t := type_.(type) {
	case ast.TypeRef:
		var ref_mod = string(t.Module.Name)
		var ref_name = string(t.Id.Name)
		if ref_mod == "" {
			if ref_name == UnitAlias {
				var v = FilledVarianceVector(Bivariant, ctx.Arity())
				return AnonymousType { Unit{} }, v, nil
			}
			for i, param := range ctx.Params {
				if param.Name == ref_name {
					var v = FilledVarianceVector(Bivariant, ctx.Arity())
					return ParameterType { Index: uint(i) }, v, nil
				}
			}
		}
		var sym = ctx.Module.SymbolFromTypeRef(t)
		switch s := sym.(type) {
		case loader.Symbol:
			var params, exists = ctx.Ireg.LookupParams(s)
			if !exists { return nil, nil, &TypeError {
				Point:    ErrorPointFrom(t.Id.Node),
				Concrete: E_TypeNotFound {
					Name: s,
				},
			} }
			var arity = uint(len(params))
			var given_arity = uint(len(t.TypeArgs))
			if arity != given_arity { return nil, nil, &TypeError {
				Point:    ErrorPointFrom(t.Node),
				Concrete: E_WrongParameterQuantity {
					TypeName: s,
					Required: arity,
					Given:    given_arity,
				},
			} }
			var args = make([] Type, arity)
			var args_v = make([][] TypeVariance, arity)
			for i, arg_node := range t.TypeArgs {
				var arg, arg_v, err = TypeFrom(arg_node.Type, ctx)
				if err != nil { return nil, nil, err }
				args[i] = arg
				args_v[i] = arg_v
			}
			var params_v = ParamsVarianceVector(params)
			var v = ctx.DeduceVariance(params_v, args_v)
			return NamedType {
				Name: s,
				Args: args,
			}, v, nil
		default:
			return nil, nil, &TypeError {
				Point:    ErrorPointFrom(t.Module.Node),
				Concrete: E_ModuleOfTypeRefNotFound {
					Name: loader.Id2String(t.Module),
				},
			}
		}
	case ast.TypeLiteral:
		return TypeFromRepr(t.Repr.Repr, ctx)
	default:
		panic("impossible branch")
	}
}

/* Transform: ast.Repr -> checker.Type */
func TypeFromRepr(repr ast.Repr, ctx TypeContext) (Type, ([] TypeVariance), *TypeError) {
	switch r := repr.(type) {
	case ast.ReprTuple:
		var count = uint(len(r.Elements))
		var max = uint(common.ProductMaxSize)
		if count > max {
			return nil, nil, &TypeError {
				Point:    ErrorPointFrom(r.Node),
				Concrete: E_TooManyTupleBundleItems {
					Defined: count,
					Limit:   max,
				},
			}
		}
		if count == 0 {
			// there isn't an empty tuple,
			// assume it to be the unit type
			var v = FilledVarianceVector(Bivariant, ctx.Arity())
			return AnonymousType {
				Repr: Unit {},
			}, v, nil
		} else {
			var n = uint(len(r.Elements))
			var elements = make([] Type, n)
			var elements_v = make([][] TypeVariance, n)
			for i, el := range r.Elements {
				var e, ev, err = TypeFrom(el.Type, ctx)
				if err != nil { return nil, nil, err }
				elements[i] = e
				elements_v[i] = ev
			}
			if len(elements) == 1 {
				// there isn't a tuple with 1 element,
				// simply forward the inner type
				return elements[0], elements_v[0], nil
			} else {
				var tuple_v = FilledVarianceVector(Covariant, n)
				var v = ctx.DeduceVariance(tuple_v, elements_v)
				return AnonymousType {
					Repr: Tuple { Elements: elements },
				}, v, nil
			}
		}
	case ast.ReprBundle:
		var count = uint(len(r.Fields))
		var max = uint(common.ProductMaxSize)
		if count > max {
			return nil, nil, &TypeError {
				Point:    ErrorPointFrom(r.Node),
				Concrete: E_TooManyTupleBundleItems {
					Defined: count,
					Limit:   max,
				},
			}
		}
		if len(r.Fields) == 0 {
			// there isn't an empty bundle,
			// assume it to be the unit type
			var v = FilledVarianceVector(Bivariant, ctx.Arity())
			return AnonymousType {
				Repr: Unit {},
			}, v, nil
		} else {
			var n = uint(len(r.Fields))
			var fields = make(map[string] Field)
			var fields_v = make([][] TypeVariance, n)
			for i, f := range r.Fields {
				var f_name = loader.Id2String(f.Name)
				if f_name == IgnoreMark {
					return nil, nil, &TypeError {
						Point:    ErrorPointFrom(f.Name.Node),
						Concrete: E_InvalidFieldName { f_name },
					}
				}
				var _, exists = fields[f_name]
				if exists { return nil, nil, &TypeError {
					Point:    ErrorPointFrom(f.Name.Node),
					Concrete: E_DuplicateField { f_name },
				} }
				var f_type, f_v, err = TypeFrom(f.Type.Type, ctx)
				if err != nil { return nil, nil, err }
				fields[f_name] = Field {
					Type:  f_type,
					Index: uint(i),
				}
				fields_v[i] = f_v
			}
			var bundle_v = FilledVarianceVector(Covariant, n)
			var v = ctx.DeduceVariance(bundle_v, fields_v)
			return AnonymousType {
				Repr: Bundle {
					Fields: fields,
				},
			}, v, nil
		}
	case ast.ReprFunc:
		var input, input_v, err1 = TypeFrom(r.Input.Type, ctx)
		if err1 != nil { return nil, nil, err1 }
		var output, output_v, err2 = TypeFrom(r.Output.Type, ctx)
		if err2 != nil { return nil, nil, err2 }
		var func_v = [] TypeVariance { Contravariant, Covariant }
		var io_v = [][] TypeVariance { input_v, output_v }
		var v = ctx.DeduceVariance(func_v, io_v)
		return AnonymousType {
			Repr: Func {
				Input:  input,
				Output: output,
			},
		}, v, nil
	default:
		panic("impossible branch")
	}
}
