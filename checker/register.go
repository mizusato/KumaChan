package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/runtime/common"
	"kumachan/stdlib"
	"kumachan/transformer/ast"
)

// Final Registry of Types
type TypeRegistry  map[loader.Symbol] *GenericType

// Intermediate Registry of Types, Used When Defining Types
type RawTypeRegistry struct {
	// a map from symbol to type declaration (AST node)
	DeclMap        map[loader.Symbol] ast.DeclType
	// a map from case type to union type (e.g. Just -> Maybe, Null -> Maybe)
	// (parameters of cases are inherited from the out-most union type)
	UnionRootMap   map[loader.Symbol] loader.Symbol
	// a map from case types to their corresponding indexes
	CaseIndexMap   map[loader.Symbol] uint
	// a context value to track all visited modules
	// (same module may appear many times when walking through the hierarchy)
	VisitedMod     map[string] bool
}
func MakeRawTypeRegistry() RawTypeRegistry {
	return RawTypeRegistry {
		DeclMap:      make(map[loader.Symbol] ast.DeclType),
		UnionRootMap: make(map[loader.Symbol] loader.Symbol),
		CaseIndexMap: make(map[loader.Symbol] uint),
		VisitedMod:   make(map[string] bool),
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
	var root_map = make(map[int]int)
	var index_map = make(map[int]int)
	for _, cmd := range mod.Node.Commands {
		switch c := cmd.Command.(type) {
		case ast.DeclType:
			var cur_union_index = len(decls)
			decls = append(decls, c)
			var root_of_union, root_of_union_exists = root_map[cur_union_index]
			switch u := c.TypeValue.TypeValue.(type) {
			case ast.UnionType:
				for index, item := range u.Items {
					var cur_sub_index = len(decls)
					decls = append(decls, item)
					index_map[cur_sub_index] = index
					if root_of_union_exists {
						root_map[cur_sub_index] = root_of_union
					} else {
						root_map[cur_sub_index] = cur_union_index
					}
				}
			}
		}
	}
	// 3. Go through all type declarations
	for i, d := range decls {
		// 3.1. Get the symbol of the declared type
		var type_sym = mod.SymbolFromName(d.Name)
		// 3.2. Check if the symbol name is valid
		if type_sym.SymbolName == IgnoreMark {
			return &TypeDeclError {
				Point:    ErrorPoint { CST: mod.CST, Node: d.Name.Node },
				Concrete: E_InvalidTypeName { type_sym.SymbolName },
			}
		}
		// 3.3. Check if the symbol is used
		var _, exists = raw.DeclMap[type_sym]
		if exists || (mod_name != stdlib.Core && loader.IsPreloadCoreSymbol(type_sym)) {
			// 3.3.1. If used, throw an error
			return &TypeDeclError {
				Point: ErrorPoint { CST: mod.CST, Node: d.Name.Node },
				Concrete: E_DuplicateTypeDecl {
					TypeName: type_sym,
				},
			}
		} else {
			// 3.3.2. If not, register the declaration node to DeclMap
			//        and update UnionRootMap and CaseIndexMap if necessary.
			//        If parameters were declared on a case type,
			//        throw an error.
			raw.DeclMap[type_sym] = d
			var root, exists = root_map[i]
			if exists {
				// TODO: this behaviour is bad
				//       parameters of a case type should be a subset
				//       of the parameters of its union type
				if len(d.Params) > 0 {
					return &TypeDeclError {
						Point: ErrorPoint { CST: mod.CST, Node: d.Name.Node },
						Concrete: E_GenericUnionSubType {
							TypeName: type_sym,
						},
					}
				}
				raw.UnionRootMap[type_sym] = mod.SymbolFromName(decls[root].Name)
				raw.CaseIndexMap[type_sym] = uint(index_map[i])
			}
		}
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
	Params  [] string
	Ireg    AbstractRegistry
}
type AbstractRegistry interface {
	// to check
	// 1. if a type with given symbol exists
	// 2. if its arity is correct
	LookupArity(loader.Symbol) (bool, uint)
}
func (raw RawTypeRegistry) LookupArity(name loader.Symbol) (bool, uint) {
	var t, exists = raw.DeclMap[name]
	if exists {
		var ur_name, exists = raw.UnionRootMap[name]
		if exists {
			// if union root exists, use the arity of the union root
			var ur = raw.DeclMap[ur_name]  // the value must exist,
										   // thus omit checking
			return true, uint(len(ur.Params))
		} else {
			// else, use the arity of the type itself
			return true, uint(len(t.Params))
		}
	} else {
		return false, 0
	}
}
func (reg TypeRegistry) LookupArity(name loader.Symbol) (bool, uint) {
	var t, exists = reg[name]
	if exists {
		return true, uint(len(t.Params))
	} else {
		return false, 0
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
		// 3.1. Determine parameters
		var params_t ast.DeclType
		var root, exists = raw.UnionRootMap[name]
		if exists {
			// 3.1.1. If union root exists, use the parameters of it
			params_t = raw.DeclMap[root]
		} else {
			// 3.1.2. Otherwise, use the parameters of the type itself
			params_t = t
		}
		var params = make([]string, len(params_t.Params))
		for i, param := range params_t.Params {
			params[i] = loader.Id2String(param)
		}
		// 3.2. Construct a TypeContext and pass it to TypeValFrom()
		//      to generate a TypeVal and bubble errors
		var val, err = TypeValFrom(t.TypeValue.TypeValue, TypeContext {
			Module: mod,
			Params: params,
			Ireg:   raw,
		})
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
			Params:    params,
			Value:     val,
			Node:      t.Node,
			CaseIndex: raw.CaseIndexMap[name],
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
				var cst = idx[path[0].ModuleName].CST
				var point = ErrorPoint { Node: node, CST: cst }
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
func TypeValFrom (tv ast.TypeValue, ctx TypeContext) (TypeVal, *TypeError) {
	switch v := tv.(type) {
	case ast.UnionType:
		var count = uint(len(v.Items))
		var max = uint(common.SumMaxBranches)
		if count > max {
			return nil, &TypeError {
				Point:    ErrorPoint { CST: ctx.Module.CST, Node: v.Node },
				Concrete: E_TooManyUnionItems {
					Defined: count,
					Limit:   max,
				},
			}
		}
		var case_types = make([]loader.Symbol, count)
		for i, item := range v.Items {
			case_types[i] = ctx.Module.SymbolFromName(item.Name)
		}
		return Union {
			CaseTypes: case_types,
		}, nil
	case ast.BoxedType:
		var inner, specified = v.Inner.(ast.VariousType)
		if specified {
			var inner_type, err = TypeFrom(inner.Type, ctx)
			if err != nil {
				return nil, err
			}
			return Boxed {
				InnerType: inner_type,
				Protected: v.Protected,
				Opaque:    v.Opaque,
			}, nil
		} else {
			return Boxed {
				InnerType: AnonymousType { Unit{} },
				Protected: v.Protected,
				Opaque:    v.Opaque,
			}, nil
		}
	case ast.NativeType:
		return Native{}, nil
	default:
		panic("impossible branch")
	}
}

/* Transform: ast.Type -> checker.Type */
func TypeFrom (type_ ast.Type, ctx TypeContext) (Type, *TypeError) {
	switch t := type_.(type) {
	case ast.TypeRef:
		var ref_mod = string(t.Ref.Module.Name)
		var ref_name = string(t.Ref.Id.Name)
		if ref_mod == "" {
			for i, param := range ctx.Params {
				if param == ref_name {
					return ParameterType { Index: uint(i) }, nil
				}
			}
		}
		var sym = ctx.Module.TypeSymbolFromRef(t.Ref)
		switch s := sym.(type) {
		case loader.Symbol:
			var exists, arity = ctx.Ireg.LookupArity(s)
			if !exists { return nil, &TypeError {
				Point:    ErrorPoint { CST: ctx.Module.CST, Node: t.Ref.Id.Node },
				Concrete: E_TypeNotFound {
					Name: s,
				},
			} }
			var given_arity = uint(len(t.Ref.TypeArgs))
			if arity != given_arity { return nil, &TypeError {
				Point:    ErrorPoint { CST: ctx.Module.CST, Node: t.Ref.Node },
				Concrete: E_WrongParameterQuantity {
					TypeName: s,
					Required: arity,
					Given:    given_arity,
				},
			} }
			var args = make([]Type, arity)
			for i, arg_node := range t.Ref.TypeArgs {
				var arg, err = TypeFrom(arg_node.Type, ctx)
				if err != nil { return nil, err }
				args[i] = arg
			}
			return NamedType {
				Name: s,
				Args: args,
			}, nil
		default:
			return nil, &TypeError {
				Point: ErrorPoint { CST: ctx.Module.CST, Node: t.Ref.Module.Node },
				Concrete: E_ModuleOfTypeRefNotFound {
					Name: loader.Id2String(t.Ref.Module),
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
func TypeFromRepr (repr ast.Repr, ctx TypeContext) (Type, *TypeError) {
	switch r := repr.(type) {
	case ast.ReprTuple:
		var count = uint(len(r.Elements))
		var max = uint(common.ProductMaxSize)
		if count > max {
			return nil, &TypeError {
				Point:    ErrorPoint { CST: ctx.Module.CST, Node: r.Node },
				Concrete: E_TooManyTupleBundleItems {
					Defined: count,
					Limit:   max,
				},
			}
		}
		if count == 0 {
			// there isn't an empty tuple,
			// assume it to be the unit type
			return AnonymousType {
				Repr: Unit {},
			}, nil
		} else {
			var elements = make([]Type, len(r.Elements))
			for i, el := range r.Elements {
				var e, err = TypeFrom(el.Type, ctx)
				if err != nil { return nil, err }
				elements[i] = e
			}
			if len(elements) == 1 {
				// there isn't a tuple with 1 element,
				// simply forward the inner type
				return elements[0], nil
			} else {
				return AnonymousType {
					Repr: Tuple { Elements: elements },
				}, nil
			}
		}
	case ast.ReprBundle:
		var count = uint(len(r.Fields))
		var max = uint(common.ProductMaxSize)
		if count > max {
			return nil, &TypeError {
				Point:    ErrorPoint { CST: ctx.Module.CST, Node: r.Node },
				Concrete: E_TooManyTupleBundleItems {
					Defined: count,
					Limit:   max,
				},
			}
		}
		if len(r.Fields) == 0 {
			// there isn't an empty bundle,
			// assume it to be the unit type
			return AnonymousType {
				Repr: Unit {},
			}, nil
		} else {
			var fields = make(map[string]Field)
			for i, f := range r.Fields {
				var f_name = loader.Id2String(f.Name)
				if f_name == IgnoreMark {
					return nil, &TypeError {
						Point:    ErrorPoint {
							CST: ctx.Module.CST, Node: f.Name.Node,
						},
						Concrete: E_InvalidFieldName { f_name },
					}
				}
				var _, exists = fields[f_name]
				if exists { return nil, &TypeError {
					Point: ErrorPoint {
						CST: ctx.Module.CST, Node: f.Name.Node,
					},
					Concrete: E_DuplicateField { f_name },
				} }
				var f_type, err = TypeFrom(f.Type.Type, ctx)
				if err != nil { return nil, err }
				fields[f_name] = Field {
					Type:  f_type,
					Index: uint(i),
				}
			}
			return AnonymousType {
				Repr: Bundle {
					Fields: fields,
				},
			}, nil
		}
	case ast.ReprFunc:
		var input, err1 = TypeFrom(r.Input.Type, ctx)
		if err1 != nil { return nil, err1 }
		var output, err2 = TypeFrom(r.Output.Type, ctx)
		if err2 != nil { return nil, err2 }
		return AnonymousType {
			Repr: Func {
				Input:  input,
				Output: output,
			},
		}, nil
	default:
		panic("impossible branch")
	}
}
