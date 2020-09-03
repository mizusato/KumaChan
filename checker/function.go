package checker

import (
	"strings"
	. "kumachan/error"
	"kumachan/loader"
	"kumachan/parser/ast"
)


type GenericFunction struct {
	Node          ast.Node
	Public        bool
	TypeParams    [] TypeParam
	TypeBounds    TypeBounds
	Implicit      map[string] Field
	DeclaredType  Func
	Body          ast.Body
}

type FunctionReference struct {
	Function    *GenericFunction  // Pointer to the underlying function
	ModuleName  string
	Index       uint
	IsImported  bool              // If it is imported from another module
}

// Map from names of available functions in the module to their references
type FunctionCollection  map[string] []FunctionReference

// Map from module names to their corresponding function collections
type FunctionStore map[string] FunctionCollection


// Procedure to collect all functions in a module hierarchy
func CollectFunctions(mod *loader.Module, reg TypeRegistry, store FunctionStore) (FunctionCollection, *FunctionError) {
	/**
	 *  Input: a root module, a type registry and an empty function store
	 *  Output: collected functions of the root module (or an error)
	 *  Effect: fill all collected functions into the function store
	 */
	// 1. Check if the module was visited, if so, return the existing result
	var mod_name = mod.Name
	var existing, exists = store[mod_name]
	if exists {
		return existing, nil
	}
	// 2. Iterate over all modules imported by the current module
	var collection = make(FunctionCollection)
	for _, imported := range mod.ImpMap {
		// 2.1. Call self recursively to collect functions of imported modules
		var imp_col, err = CollectFunctions(imported, reg, store)
		if err != nil { return nil, err }
		// 2.2. Iterate over all functions in imported modules
		for name, refs := range imp_col {
			for _, ref := range refs {
				if !ref.IsImported && ref.Function.Public {
					// 2.2.1. If the function is exported, import it
					var _, exists = collection[name]
					if !exists {
						collection[name] = make([]FunctionReference, 0)
					}
					// Note: conflict (unsafe overload) must not happen there,
					//       since public functions have local signatures
					collection[name] = append(collection[name], FunctionReference {
						Function:   ref.Function,
						ModuleName: ref.ModuleName,
						Index:      ref.Index,
						IsImported: true,
					})
				}
			}
		}
	}
	var index_offset_map = make(map[string] uint)
	for name, refs := range collection {
		index_offset_map[name] = uint(len(refs))
	}
	// 3. Iterate over all function declarations in the current module
	for _, stmt := range mod.Node.Statements {
		switch decl := stmt.Statement.(type) {
		case ast.DeclFunction:
			// 3.1. Get the name of the function and its type parameters
			var name = loader.Id2String(decl.Name)
			if name == IgnoreMark || strings.HasSuffix(name, FuncSuffix) {
				// 3.1.1. If the function name is invalid, throw an error.
				return nil, &FunctionError {
					Point:    ErrorPointFrom(decl.Name.Node),
					Concrete: E_InvalidFunctionName { name },
				}
			}
			var params, raw_bounds, p_err, p_err_node = CollectTypeParams(decl.Params)
			if p_err != nil { return nil, &FunctionError {
				Point:    ErrorPointFrom(p_err_node),
				Concrete: E_FunctionInvalidTypeParameterName { p_err.Name },
			} }
			var bounds = TypeBounds {
				Sub:   make(map[uint] Type),
				Super: make(map[uint] Type),
			}
			// 3.2. Create a context for evaluating types
			var ctx = TypeContext {
				TypeBoundsContext: TypeBoundsContext {
					TypeValidationContext: TypeValidationContext {
						TypeConstructContext: TypeConstructContext {
							Module:     mod,
							Parameters: params,
						},
						Registry: reg,
					},
					Bounds: bounds,
				},
			}
			var bounds_info = make(map[Type] ast.Node)
			var got_bound = func(m (map[uint] Type), i int, b ast.VariousType) *FunctionError {
				var t, info, err = TypeNoBoundCheckFrom(b, ctx.TypeValidationContext)
				if err != nil { return &FunctionError {
					Point:    err.Point,
					Concrete: E_InvalidTypeInFunction {
						TypeError: err,
					},
				} }
				m[uint(i)] = t
				for k, v := range info {
					bounds_info[k] = v
				}
				return nil
			}
			for i, bound := range raw_bounds {
				switch b := bound.(type) {
				case ast.TypeLowerBound:
					var err = got_bound(bounds.Sub, i, b.BoundType)
					if err != nil { return nil, err }
				case ast.TypeHigherBound:
					var err = got_bound(bounds.Super, i, b.BoundType)
					if err != nil { return nil, err }
				}
			}
			for _, group := range [] (map[uint] Type) { bounds.Super, bounds.Sub } {
				for _, t := range group {
					var err = CheckTypeBounds(t, bounds_info, ctx.TypeBoundsContext)
					if err != nil { return nil, &FunctionError {
						Point:    err.Point,
						Concrete: E_InvalidTypeInFunction {
							TypeError: err,
						},
					} }
				}
			}
			// 3.3. Collect implicit context value definitions
			var implicit_fields = make(map[string] Field)
			var implicit_types = make([] Type, len(decl.Implicit))
			if len(decl.Implicit) > 0 {
				var _, is_native = decl.Body.Body.(ast.NativeRef)
				if is_native {
					return nil, &FunctionError {
						Point:    ErrorPointFrom(decl.Node),
						Concrete: E_ImplicitContextOnNativeFunction {},
					}
				}
			}
			for i, item := range decl.Implicit {
				var item_t, err = TypeFrom(item, ctx)
				if err != nil { return nil, &FunctionError {
					Point:    err.Point,
					Concrete: E_InvalidTypeInFunction {
						TypeError: err,
					},
				} }
				implicit_types[i] = item_t
			}
			for i, t := range implicit_types {
				var throw = func(problem string) *FunctionError {
					return &FunctionError {
						Point:    ErrorPointFrom(decl.Implicit[i].Node),
						Concrete: E_InvalidImplicitContextType {
							Reason: problem,
						},
					}
				}
				var named, is_named = t.(*NamedType)
				if !(is_named) { return nil, throw("should be a named type") }
				var g = reg[named.Name]
				var boxed, is_boxed = g.Value.(*Boxed)
				if !(is_boxed) { return nil, throw("should be a boxed type") }
				if !(boxed.Implicit) { return nil,
					throw("should be declared as a implicit context type") }
				var inner = FillTypeArgs(boxed.InnerType, named.Args)
				var bundle = inner.(*AnonymousType).Repr.(Bundle)
				var offset = uint(len(implicit_fields))
				for name, field := range bundle.Fields {
					var _, exists = implicit_fields[name]
					if exists { return nil, &FunctionError {
						Point:    ErrorPointFrom(decl.Node),
						Concrete: E_ConflictImplicitContextField { name },
					} }
					implicit_fields[name] = Field {
						Type:  field.Type,
						Index: (offset + field.Index),
					}
				}
			}
			// 3.4. Evaluate the function signature using the created context
			var sig, err = TypeFromRepr(ast.VariousRepr {
				Node: decl.Repr.Node,
				Repr: decl.Repr,
			}, ctx)
			if err != nil { return nil, &FunctionError {
				Point:    err.Point,
				Concrete: E_InvalidTypeInFunction {
					TypeError: err,
				},
			} }
			for i, param := range params {
				if param.Variance != Invariant {
					return nil, &FunctionError {
						Point:    ErrorPointFrom(decl.Params[i].Node),
						Concrete: E_FunctionVarianceDeclared {},
					}
				}
			}
			// 3.5. If the function is public, ensure its signature type
			//        to be a local type of this module.
			var is_public = decl.Public
			if is_public && !(IsExportable(mod_name, sig, implicit_fields, bounds)) {
				return nil, &FunctionError {
					Point:    ErrorPointFrom(decl.Repr.Node),
					Concrete: E_SignatureNonLocal { name },
				}
			}
			// 3.6. Construct a representation and a reference of the function
			var func_type = sig.(*AnonymousType).Repr.(Func)
			var f = &GenericFunction {
				Public:       is_public,
				TypeParams:   params,
				Implicit:     implicit_fields,
				DeclaredType: func_type,
				Body:         decl.Body.Body,
				Node:         decl.Node,
			}
			// 3.7. Check if the name of the function is in use
			var existing, exists = collection[name]
			if exists {
				// 3.7.1. If in use, try to overload it
				var index_offset = index_offset_map[name]
				var err_point = ErrorPointFrom(decl.Name.Node)
				var err = CheckOverload (
					existing,
					func_type, implicit_fields,
					name, TypeParamsNames(params),
					reg, err_point,
				)
				if err != nil { return nil, err }
				collection[name] = append(existing, FunctionReference {
					IsImported: false,
					Function:   f,
					ModuleName: mod_name,
					Index:      uint(len(existing)) - index_offset,
				})
			} else {
				// 3.7.2. If not, collect the function
				collection[name] = [] FunctionReference { {
					IsImported: false,
					Function:   f,
					ModuleName: mod_name,
					Index:      0,
				} }
			}
		}
	}
	// 4. Register all collected functions of the current module to the store
	store[mod_name] = collection
	// 5. Return all collected functions of the current module
	return collection, nil
}
