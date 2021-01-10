package checker

import (
	"strings"
	. "kumachan/util/error"
	"kumachan/compiler/loader"
	"kumachan/compiler/loader/parser/ast"
)


var __ReservedTypeNames = [...]string {
	IgnoreMark, UnitAlias,
	NeverTypeName, AnyTypeName, SuperTypeName,
}
func IsReservedTypeName(name string) bool {
	for _, reserved := range __ReservedTypeNames {
		if name == reserved {
			return true
		}
	}
	return false
}

type TypeConstructContext struct {
	Module      *loader.Module
	Parameters  [] TypeParam
}

type RawTypeContext struct {
	TypeConstructContext
	Registry    RawTypeRegistry
}

// Final Registry of Types
type TypeRegistry  map[loader.Symbol] *GenericType

// Intermediate Registry of Types, Used When Defining Types
type RawTypeRegistry struct {
	// a map from symbol to type declaration (AST node)
	DeclMap        map[loader.Symbol] ast.DeclType
	// a map from symbol to type parameters
	ParamsMap      map[loader.Symbol] ([] TypeParam)
	// a map from symbol to type parameter bounds
	BoundsMap      map[loader.Symbol] ([] ast.TypeBound)
	// a map from case types to their corresponding union information
	CaseInfoMap    map[loader.Symbol] CaseInfo
	// a context value to track all visited modules
	// (same module may appear many times when walking through the hierarchy)
	VisitedMod     map[string] bool
}
type CaseInfo struct {
	IsCaseType     bool
	UnionName      loader.Symbol
	UnionArity     uint
	UnionVariance  [] TypeVariance
	CaseIndex      uint
	CaseParams     [] uint
}

type TypeDeclNodeInfo  map[loader.Symbol] ast.Node
type TypeNodeInfo struct {
	ValNodeMap   map[TypeVal] ast.Node
	TypeNodeMap  map[Type] ast.Node
}


func GetVarianceFromRawTypeParams(raw_params ([] ast.TypeParam)) ([] TypeVariance) {
	var v = make([] TypeVariance, len(raw_params))
	for i, raw_param := range raw_params {
		var name = ast.Id2String(raw_param.Name)
		if len(name) > len(CovariantPrefix) &&
			strings.HasPrefix(name, CovariantPrefix) {
			v[i] = Covariant
		} else if len(name) > len(ContravariantPrefix) &&
			strings.HasPrefix(name, ContravariantPrefix) {
			v[i] = Contravariant
		} else {
			v[i] = Invariant
		}
	}
	return v
}

func CollectTypeParams(raw_params ([] ast.TypeParam)) (([] TypeParam), ([] ast.TypeBound), *E_InvalidTypeName, ast.Node) {
	var params = make([] TypeParam, len(raw_params))
	var bounds = make([] ast.TypeBound, len(raw_params))
	for i, raw_param := range raw_params {
		var raw_name = ast.Id2String(raw_param.Name)
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
			return nil, nil, &E_InvalidTypeName {name}, raw_param.Node
		}
		params[i] = TypeParam {
			Name:     name,
			Variance: v,
		}
		bounds[i] = raw_param.Bound.TypeBound
	}
	return params, bounds, nil, ast.Node{}
}

func MakeRawTypeRegistry() RawTypeRegistry {
	return RawTypeRegistry {
		DeclMap:       make(map[loader.Symbol] ast.DeclType),
		ParamsMap:     make(map[loader.Symbol] ([] TypeParam)),
		BoundsMap:     make(map[loader.Symbol] [] ast.TypeBound),
		CaseInfoMap:   make(map[loader.Symbol] CaseInfo),
		VisitedMod:    make(map[string] bool),
	}
}

func RegisterRawTypes(mod *loader.Module, raw RawTypeRegistry) *TypeDeclError {
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
	var decls = make([] ast.DeclType, 0)
	var params_map = make(map[uint] ([] TypeParam))
	var bounds_map = make(map[uint] ([] ast.TypeBound))
	var case_parent_map = make(map[uint]uint)
	var case_index_map = make(map[uint]uint)
	var process_decl_stmt func(ast.DeclType) *TypeDeclError
	process_decl_stmt = func(s ast.DeclType) *TypeDeclError {
		var i = uint(len(decls))
		decls = append(decls, s)
		var params, bounds, err, err_node = CollectTypeParams(s.Params)
		if err != nil { return &TypeDeclError {
			Point:    ErrorPointFrom(err_node),
			Concrete: *err,
		} }
		params_map[i] = params
		bounds_map[i] = bounds
		switch u := s.TypeValue.TypeValue.(type) {
		case ast.UnionType:
			for case_index, case_decl := range u.Cases {
				var case_i = uint(len(decls))
				var err = process_decl_stmt(case_decl)
				if err != nil { return err }
				case_index_map[case_i] = uint(case_index)
				case_parent_map[case_i] = i
			}
		}
		return nil
	}
	for _, stmt := range mod.AST.Statements {
		switch s := stmt.Statement.(type) {
		case ast.DeclType:
			var err = process_decl_stmt(s)
			if err != nil { return err }
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
		raw.BoundsMap[type_sym] = bounds_map[uint(i)]
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
				IsCaseType:    true,
				UnionName:     parent_name,
				UnionArity:    uint(len(parent.Params)),
				UnionVariance: GetVarianceFromRawTypeParams(parent.Params),
				CaseIndex:     case_index,
				CaseParams:    mapping,
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

func RegisterTypes(entry *loader.Module, idx loader.Index) (TypeRegistry, TypeDeclNodeInfo, ([] *TypeDeclError)) {
	var raise = func(name loader.Symbol, err *TypeError) *TypeDeclError {
		return &TypeDeclError {
			Point: err.Point,
			Concrete: E_InvalidTypeDecl {
				TypeName: name,
				Detail:   err,
			},
		}
	}
	var raise_all = func(name loader.Symbol, err *TypeError) ([] *TypeDeclError) {
		return [] *TypeDeclError { raise(name, err) }
	}
	var decl_info = make(TypeDeclNodeInfo)
	// 1. Build a raw registry
	var raw = MakeRawTypeRegistry()
	var err = RegisterRawTypes(entry, raw)
	if err != nil { return nil, nil, [] *TypeDeclError { err } }
	for sym, t := range raw.DeclMap {
		decl_info[sym] = t.Node
	}
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
		// 3.1. Get tags
		var tags, tags_err = ParseTypeTags(t.Tags)
		if tags_err != nil { return nil, nil, [] *TypeDeclError { {
			Point:    ErrorPointFrom(tags_err.Tag.Node),
			Concrete: E_InvalidTypeTag {
				Tag:  string(tags_err.Tag.RawContent),
				Info: tags_err.Info,
			},
		} } }
		// 3.2. Get parameters
		var params = raw.ParamsMap[name]
		// 3.3. Construct a TypeContext and pass it to TypeValFrom()
		//      to generate a TypeVal and bubble errors
		var cons_ctx = TypeConstructContext {
			Module:     mod,
			Parameters: params,
		}
		var ctx = RawTypeContext {
			TypeConstructContext: cons_ctx,
			Registry: raw,
		}
		var val, err = RawTypeValFrom(t.TypeValue, info, ctx)
		if err != nil { return nil, nil, raise_all(name, err) }
		// 3.4. Get bound types
		var bounds = TypeBounds {
			Sub:   make(map[uint] Type),
			Super: make(map[uint] Type),
		}
		var raw_bounds = raw.BoundsMap[name]
		for i, bound := range raw_bounds {
			switch b := bound.(type) {
			case ast.TypeLowerBound:
				var t, err = RawTypeFrom(b.BoundType, info.TypeNodeMap, ctx.TypeConstructContext)
				if err != nil { return nil, nil, raise_all(name, err) }
				bounds.Sub[uint(i)] = t
			case ast.TypeHigherBound:
				var t, err = RawTypeFrom(b.BoundType, info.TypeNodeMap, ctx.TypeConstructContext)
				if err != nil { return nil, nil, raise_all(name, err) }
				bounds.Super[uint(i)] = t
			}
		}
		// 3.5. Use the generated TypeVal to construct a GenericType
		//      and register it to the TypeRegistry
		reg[name] = &GenericType {
			Tags:     tags,
			Params:   params,
			Bounds:   bounds,
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
			var g, exists = reg[inner_name]
			if !(exists) {
				// refers to a type that does not exist
				return nil
			}
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
			if err != nil { return nil, nil, [] *TypeDeclError { err } }
		}
	}
	// 5. Validate types
	var errs ([] *TypeDeclError) = nil
	for name, g := range reg {
		var cons_ctx = TypeConstructContext {
			Module:     idx[name.ModuleName],
			Parameters: g.Params,
		}
		var val_ctx = TypeValidationContext {
			TypeConstructContext: cons_ctx,
			Registry:             reg,
		}
		// 5.1. Validate type values
		var err = ValidateTypeVal(g.Value, info, val_ctx)
		if err != nil {
			errs = append(errs, raise(name, err))
		}
		// 5.2. Validate bound types
		for _, bounds := range [](map[uint] Type) { g.Bounds.Sub, g.Bounds.Super } {
			for _, bound := range bounds {
				var param, is_param = bound.(*ParameterType)
				if is_param {
					errs = append(errs, raise(name, &TypeError {
						Point:    ErrorPointFrom(info.TypeNodeMap[bound]),
						Concrete: E_InvalidBoundType {
							Type: val_ctx.Parameters[param.Index].Name,
						},
					}))
				}
				var err = ValidateType(bound, info.TypeNodeMap, val_ctx)
				if err != nil {
					errs = append(errs, raise(name, err))
				}
			}
		}
	}
	// 6. Check type bounds
	for name, g := range reg {
		var cons_ctx = TypeConstructContext {
			Module:     idx[name.ModuleName],
			Parameters: g.Params,
		}
		var val_ctx = TypeValidationContext {
			TypeConstructContext: cons_ctx,
			Registry:             reg,
		}
		var bound_ctx = TypeBoundsContext {
			TypeValidationContext: val_ctx,
			Bounds:                g.Bounds,
		}
		for _, super := range g.Bounds.Super {
			var err = CheckTypeBounds(super, info.TypeNodeMap, bound_ctx)
			if err != nil { errs = append(errs, raise(name, err)) }
		}
		for _, sub := range g.Bounds.Sub {
			var err = CheckTypeBounds(sub, info.TypeNodeMap, bound_ctx)
			if err != nil { errs = append(errs, raise(name, err)) }
		}
		var err = CheckTypeValBounds(g.Value, info, bound_ctx)
		if err != nil { errs = append(errs, raise(name, err)) }
	}
	if errs != nil { return nil, nil, errs }
	// 7. Return the TypeRegistry
	return reg, decl_info, nil
}

/* Transform: ast.TypeDef -> checker.TypeVal */
func RawTypeValFrom(ast_val ast.VariousTypeDef, info TypeNodeInfo, ctx RawTypeContext) (TypeVal, *TypeError) {
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
			if err != nil { return nil, err }
			return got(&Boxed {
				InnerType: inner_type,
				Weak:      a.Weak,
				Protected: a.Protected,
				Opaque:    a.Opaque,
			})
		} else {
			return got(&Boxed {
				InnerType: &AnonymousType { Unit{} },
				Weak:      a.Weak,
				Protected: a.Protected,
				Opaque:    a.Opaque,
			})
		}
	case ast.ImplicitType:
		var repr = ast.VariousRepr {
			Node: a.Repr.Node,
			Repr: a.Repr,
		}
		var inner_type, err = RawTypeFromRepr(repr, info.TypeNodeMap, ctx.TypeConstructContext)
		if err != nil { return nil, err }
		return got(&Boxed {
			InnerType: inner_type,
			Implicit:  true,
		})
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
		if ref_mod == "" && len(a.TypeArgs) == 0 {
			if ref_name == UnitAlias {
				return got(&AnonymousType { Unit{} })
			} else if ref_name == NeverTypeName {
				return got(&NeverType {})
			} else if ref_name == AnyTypeName {
				return got(&AnyType {})
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
					Name: ast.Id2String(a.Module),
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
				var f_name = ast.Id2String(f.Name)
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

