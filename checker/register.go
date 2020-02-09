package checker

import (
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/transformer/node"
)

// Final Registry of Types
type TypeRegistry  map[loader.Symbol] *GenericType

// Intermediate Registry of Types, Used When Defining Types
type RawTypeRegistry struct {
	// a map from symbol to type declaration (transformed AST node)
	DeclMap       map[loader.Symbol] node.DeclType
	// a map from subtype to union type (e.g. Just -> Maybe, Null -> Maybe)
	// (parameters of subtypes are inherited from the out-most union type)
	UnionRootMap  map[loader.Symbol] loader.Symbol
	// a map from subtype to corresponding order
	OrderMap      map[loader.Symbol] uint
	// a context value to track all visited modules
	// (same module may appear many times when walking through the hierarchy)
	VisitedMod    map[string] bool
}
func MakeRawTypeRegistry() RawTypeRegistry {
	return RawTypeRegistry {
		DeclMap:      make(map[loader.Symbol] node.DeclType),
		UnionRootMap: make(map[loader.Symbol] loader.Symbol),
		OrderMap:     make(map[loader.Symbol]uint),
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
	// 1. Check if the module is visited, if visited, do nothing
	var mod_name = loader.Id2String(mod.Node.Name)
	var _, visited = raw.VisitedMod[mod_name]
	if visited { return nil }
	// 2. Extract all type declarations in the module,
	//    and record the root union types of all subtypes
	var decls = make([]node.DeclType, 0)
	var root_map = make(map[int]int)
	var order_map = make(map[int]int)
	for _, cmd := range mod.Node.Commands {
		switch c := cmd.Command.(type) {
		case node.DeclType:
			var cur_union_index = len(decls)
			decls = append(decls, c)
			var root_of_union, root_of_union_exists = root_map[cur_union_index]
			switch u := c.TypeValue.TypeValue.(type) {
			case node.UnionType:
				for order, item := range u.Items {
					var cur_sub_index = len(decls)
					decls = append(decls, item)
					order_map[cur_sub_index] = order
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
		// 3.2. Check if the symbol is used
		var _, exists = raw.DeclMap[type_sym]
		if exists || (mod_name != loader.CoreModule && loader.IsPreloadCoreSymbol(type_sym)) {
			// 3.2.1. If used, throw an error
			return &TypeDeclError {
				Point: ErrorPoint { AST: mod.AST, Node: d.Name.Node },
				Concrete: E_DuplicateTypeDecl {
					TypeName: type_sym,
				},
			}
		} else {
			// 3.2.2. If not, register the declaration node to DeclMap
			//        and update UnionRootMap and OrderMap if necessary.
			//        If parameters were declared on a subtype,
			//        throw an error.
			raw.DeclMap[type_sym] = d
			var root, exists = root_map[i]
			if exists {
				if len(d.Params) > 0 {
					return &TypeDeclError {
						Point: ErrorPoint { AST: mod.AST, Node: d.Name.Node },
						Concrete: E_GenericUnionSubType {
							TypeName: type_sym,
						},
					}
				}
				raw.UnionRootMap[type_sym] = mod.SymbolFromName(decls[root].Name)
				raw.OrderMap[type_sym] = uint(order_map[i])
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

type TypeExprContext struct {
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
		return true, t.Arity
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
		var params_t node.DeclType
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
		// 3.2. Construct a TypeExprContext and pass it to TypeValFrom()
		//      to generate a TypeVal and bubble errors
		var val, err = TypeValFrom(t.TypeValue.TypeValue, TypeExprContext {
			Module: mod,
			Params: params,
			Ireg:   raw,
		})
		if err != nil { return nil, &TypeDeclError {
			Point: ErrorPoint { AST: mod.AST, Node: t.Name.Node },
			Concrete: E_InvalidTypeDecl {
				TypeName:  name,
				ExprError: err,
			},
		} }
		// 3.3. Use the generated TypeVal to construct a GenericType
		//      and register it to the TypeRegistry
		reg[name] = &GenericType {
			Arity:     uint(len(params)),
			IsOpaque:  t.IsOpaque,
			Value:     val,
			Node:      t.Node,
		}
	}
	// 4. Return the TypeRegistry
	return reg, nil
}

/* Transform: node.TypeValue -> checker.TypeVal */
func TypeValFrom (tv node.TypeValue, ctx TypeExprContext) (TypeVal, *TypeExprError) {
	switch v := tv.(type) {
	case node.UnionType:
		var subtypes = make([]loader.Symbol, len(v.Items))
		for i, item := range v.Items {
			subtypes[i] = ctx.Module.SymbolFromName(item.Name)
		}
		return UnionTypeVal {
			SubTypes: subtypes,
		}, nil
	case node.CompoundType:
		var expr, err = TypeExprFromRepr(v.Repr.Repr, ctx)
		if err != nil { return nil, err }
		return SingleTypeVal {
			Expr: expr,
		}, nil
	case node.NativeType:
		return NativeTypeVal{}, nil
	default:
		panic("impossible branch")
	}
}

/* Transform: node.Type -> checker.TypeExpr */
func TypeExprFrom (type_ node.Type, ctx TypeExprContext) (TypeExpr, *TypeExprError) {
	switch t := type_.(type) {
	case node.TypeRef:
		var ref_mod = string(t.Ref.Module.Name)
		var ref_name = string(t.Ref.Id.Name)
		if ref_mod == "" {
			for i, param := range ctx.Params {
				if param == ref_name {
					return ParameterType { Index: uint(i) }, nil
				}
			}
		}
		var sym = ctx.Module.SymbolFromRef(t.Ref)
		switch s := sym.(type) {
		case loader.Symbol:
			var exists, arity = ctx.Ireg.LookupArity(s)
			if !exists { return nil, &TypeExprError {
				Point: ErrorPoint { AST: ctx.Module.AST, Node: t.Ref.Id.Node },
				Concrete: E_TypeNotFound {
					Name: s,
				},
			} }
			var given_arity = uint(len(t.Ref.TypeArgs))
			if arity != given_arity { return nil, &TypeExprError {
				Point: ErrorPoint { AST: ctx.Module.AST, Node: t.Ref.Node },
				Concrete: E_WrongParameterQuantity {
					TypeName: s,
					Required: arity,
					Given:    given_arity,
				},
			} }
			var args = make([]TypeExpr, arity)
			for i, arg_node := range t.Ref.TypeArgs {
				var arg, err = TypeExprFrom(arg_node.Type, ctx)
				if err != nil { return nil, err }
				args[i] = arg
			}
			return NamedType {
				Name: s,
				Args: args,
			}, nil
		default:
			return nil, &TypeExprError {
				Point: ErrorPoint { AST: ctx.Module.AST, Node: t.Ref.Module.Node },
				Concrete: E_ModuleOfTypeRefNotFound {
					Name: loader.Id2String(t.Ref.Module),
				},
			}
		}
	case node.TypeLiteral:
		return TypeExprFromRepr(t.Repr.Repr, ctx)
	default:
		panic("impossible branch")
	}
}

/* Transform: node.Repr -> checker.TypeExpr */
func TypeExprFromRepr (repr node.Repr, ctx TypeExprContext) (TypeExpr, *TypeExprError) {
	switch r := repr.(type) {
	case node.ReprTuple:
		if len(r.Elements) == 0 {
			// there isn't an empty tuple,
			// assume it to be the unit type
			return AnonymousType {
				Repr: Unit {},
			}, nil
		} else {
			var elements = make([]TypeExpr, len(r.Elements))
			for i, el := range r.Elements {
				var e, err = TypeExprFrom(el.Type, ctx)
				if err != nil { return nil, err }
				elements[i] = e
			}
			if len(elements) == 1 {
				// there isn't a tuple with 1 element,
				// simply forward the inner type
				return elements[0], nil
			} else {
				return AnonymousType {
					Repr: Tuple { Elements:elements },
				}, nil
			}
		}
	case node.ReprBundle:
		if len(r.Fields) == 0 {
			// there isn't an empty bundle,
			// assume it to be the unit type
			return AnonymousType {
				Repr: Unit {},
			}, nil
		} else {
			var fields = make(map[string]TypeExpr)
			for _, f := range r.Fields {
				var f_name = loader.Id2String(f.Name)
				var _, exists = fields[f_name]
				if exists { return nil, &TypeExprError {
					Point: ErrorPoint { AST: ctx.Module.AST, Node: f.Name.Node },
					Concrete: E_DuplicateField {
						FieldName: f_name,
					},
				} }
				var f_type, err = TypeExprFrom(f.Type.Type, ctx)
				if err != nil { return nil, err }
				fields[f_name] = f_type
			}
			return AnonymousType {
				Repr: Bundle { Fields: fields },
			}, nil
		}
	case node.ReprFunc:
		var input, err1 = TypeExprFrom(r.Input.Type, ctx)
		if err1 != nil { return nil, err1 }
		var output, err2 = TypeExprFrom(r.Output.Type, ctx)
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