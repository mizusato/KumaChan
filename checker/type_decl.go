package checker

import (
	"strings"
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/parser/ast"
)


var __ReservedTypeNames = [...]string {
	IgnoreMark, UnitAlias, WildcardRhsTypeName,
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

type TypeConstructContext struct {
	Module      *loader.Module
	Parameters  [] TypeParam
}
type RawTypeContext struct {
	TypeConstructContext
	Registry    RawTypeRegistry
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

func RegisterTypes (entry *loader.Module, idx loader.Index) (TypeRegistry, ([] *TypeDeclError)) {
	// 1. Build a raw registry
	var raw = MakeRawTypeRegistry()
	var err = RegisterRawTypes(entry, raw)
	if err != nil { return nil, [] *TypeDeclError { err } }
	// 2. Create a empty TypeRegistry
	var reg = make(TypeRegistry)
	// 3. Go through all types in the raw registry
	var info = TypeNodeInfo {
		ValNodeMap:  make(map[TypeVal] ast.Node),
		TypeNodeMap: make(map[Type] ast.Node),
	}
	for name, t := range raw.DeclMap {
		var mod, mod_exists = idx[name.ModuleName]
		if !mod_exists { panic("mod " + name.ModuleName + " should exist") }
		// 3.1. Get parameters
		var params = raw.ParamsMap[name]
		// 3.2. Construct a TypeContext and pass it to TypeValFrom()
		//      to generate a TypeVal and bubble errors
		var val, err = RawTypeValFrom(t.TypeValue, info, RawTypeContext {
			TypeConstructContext: TypeConstructContext {
				Module:     mod,
				Parameters: params,
			},
			Registry: raw,
		})
		if err != nil { return nil, [] *TypeDeclError { &TypeDeclError {
			Point:    err.Point,
			Concrete: E_InvalidTypeDecl {
				TypeName: name,
				Detail:   err,
			},
		} } }
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
	var check_cycle func(loader.Symbol, *Boxed, []loader.Symbol) *TypeDeclError
	check_cycle = func (
		name loader.Symbol, t *Boxed, path []loader.Symbol,
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
		var named, is_named = t.InnerType.(*NamedType)
		if is_named {
			var inner_name = named.Name
			var g = reg[inner_name]
			var inner_boxed, is_boxed = g.Value.(*Boxed)
			if is_boxed {
				var inner_path = append(path, name)
				return check_cycle(inner_name, inner_boxed, inner_path)
			}
		}
		return nil
	}
	for name, g := range reg {
		var boxed, is_boxed = g.Value.(*Boxed)
		if is_boxed {
			var err = check_cycle(name, boxed, make([]loader.Symbol, 0))
			if err != nil { return nil, [] *TypeDeclError { err } }
		}
	}
	// 5. Validate types
	var errs ([] *TypeDeclError) = nil
	for name, g := range reg {
		var cons_ctx = TypeConstructContext {
			Module:     idx[name.ModuleName],
			Parameters: g.Params,
		}
		var err = ValidateTypeVal(g.Value, info, TypeContext{
			TypeConstructContext: cons_ctx,
			Registry:             reg,
		})
		if err != nil {
			errs = append(errs, &TypeDeclError {
				Point: err.Point,
				Concrete: E_InvalidTypeDecl {
					TypeName: name,
					Detail:   err,
				},
			})
		}
	}
	if errs != nil { return nil, errs }
	// 6. Return the TypeRegistry
	return reg, nil
}

/* Transform: ast.TypeValue -> checker.TypeVal */
func RawTypeValFrom(ast_val ast.VariousTypeValue, info TypeNodeInfo, ctx RawTypeContext) (TypeVal, *TypeError) {
	var got = func(val TypeVal) (TypeVal, *TypeError) {
		info.ValNodeMap[val] = ast_val.Node
		return val, nil
	}
	switch a := ast_val.TypeValue.(type) {
	case ast.UnionType:
		var raw_reg = ctx.Registry
		var case_types = make([] CaseType, len(a.Cases))
		for i, case_decl := range a.Cases {
			var sym = ctx.Module.SymbolFromDeclName(case_decl.Name)
			var mapping = raw_reg.CaseInfoMap[sym].CaseParams
			case_types[i] = CaseType {
				Name:   sym,
				Params: mapping,
			}
		}
		return got(&Union { case_types })
	case ast.BoxedType:
		var inner, specified = a.Inner.(ast.VariousType)
		if specified {
			var inner_type, err = RawTypeFrom(inner, info.TypeNodeMap, ctx.TypeConstructContext)
			if err != nil {
				return nil, err
			}
			return got(&Boxed {
				InnerType: inner_type,
				AsIs:      a.AsIs,
				Protected: a.Protected,
				Opaque:    a.Opaque,
			})
		} else {
			return got(&Boxed {
				InnerType: &AnonymousType { Unit{} },
				AsIs:      a.AsIs,
				Protected: a.Protected,
				Opaque:    a.Opaque,
			})
		}
	case ast.NativeType:
		return got(&Native{})
	default:
		panic("impossible branch")
	}
}

/* Transform: ast.Type -> checker.Type */
func RawTypeFrom(ast_type ast.VariousType, info (map[Type] ast.Node), ctx TypeConstructContext) (Type, *TypeError) {
	var got = func(t Type) (Type, *TypeError) {
		info[t] = ast_type.Node
		return t, nil
	}
	switch a := ast_type.Type.(type) {
	case ast.TypeRef:
		var ref_mod = string(a.Module.Name)
		var ref_name = string(a.Id.Name)
		if ref_mod == "" {
			if ref_name == UnitAlias {
				return got(&AnonymousType { Unit{} })
			}
			if ref_name == WildcardRhsTypeName {
				return got(&WildcardRhsType {})
			}
			for i, param := range ctx.Parameters {
				if param.Name == ref_name {
					return got(&ParameterType { Index: uint(i) })
				}
			}
		}
		var sym = ctx.Module.SymbolFromTypeRef(a)
		switch s := sym.(type) {
		case loader.Symbol:
			var args = make([] Type, len(a.TypeArgs))
			for i, ast_arg := range a.TypeArgs {
				var arg, err = RawTypeFrom(ast_arg, info, ctx)
				if err != nil { return nil, err }
				args[i] = arg
			}
			return got(&NamedType {
				Name: s,
				Args: args,
			})
		default:
			return nil, &TypeError {
				Point:    ErrorPointFrom(a.Module.Node),
				Concrete: E_ModuleOfTypeRefNotFound {
					Name: loader.Id2String(a.Module),
				},
			}
		}
	case ast.TypeLiteral:
		return RawTypeFromRepr(a.Repr, info, ctx)
	default:
		panic("impossible branch")
	}
}

/* Transform: ast.Repr -> checker.Type */
func RawTypeFromRepr(ast_repr ast.VariousRepr, info (map[Type] ast.Node), ctx TypeConstructContext) (Type, *TypeError) {
	var got = func(t Type) (Type, *TypeError) {
		info[t] = ast_repr.Node
		return t, nil
	}
	switch a := ast_repr.Repr.(type) {
	case ast.ReprTuple:
		var count = uint(len(a.Elements))
		if count == 0 {
			// there isn't an empty tuple,
			// assume it to be the unit type
			return got(&AnonymousType { Unit{} })
		} else {
			var n = uint(len(a.Elements))
			var elements = make([] Type, n)
			for i, el := range a.Elements {
				var e, err = RawTypeFrom(el, info, ctx)
				if err != nil { return nil, err }
				elements[i] = e
			}
			if len(elements) == 1 {
				// there isn't a tuple with 1 element,
				// simply forward the inner type
				return got(elements[0])
			} else {
				return got(&AnonymousType { Tuple { elements } })
			}
		}
	case ast.ReprBundle:
		if len(a.Fields) == 0 {
			// there isn't an empty bundle,
			// assume it to be the unit type
			return got(&AnonymousType { Unit{} })
		} else {
			var fields = make(map[string] Field)
			for i, f := range a.Fields {
				var f_name = loader.Id2String(f.Name)
				if f_name == IgnoreMark {
					return nil, &TypeError {
						Point:    ErrorPointFrom(f.Name.Node),
						Concrete: E_InvalidFieldName { f_name },
					}
				}
				var _, exists = fields[f_name]
				if exists { return nil, &TypeError {
					Point:    ErrorPointFrom(f.Name.Node),
					Concrete: E_DuplicateField { f_name },
				} }
				var f_type, err = RawTypeFrom(f.Type, info, ctx)
				if err != nil { return nil, err }
				fields[f_name] = Field {
					Type:  f_type,
					Index: uint(i),
				}
			}
			return got(&AnonymousType { Bundle { fields } })
		}
	case ast.ReprFunc:
		var input, err1 = RawTypeFrom(a.Input, info, ctx)
		if err1 != nil { return nil, err1 }
		var output, err2 = RawTypeFrom(a.Output, info, ctx)
		if err2 != nil { return nil, err2 }
		return got(&AnonymousType { Func {
			Input:  input,
			Output: output,
		} })
	default:
		panic("impossible branch")
	}
}
