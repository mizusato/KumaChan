package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
	. "kumachan/error"
)

type GenericFunction struct {
	Node          node.Node
	Public        bool
	TypeParams    [] string
	DeclaredType  Func
	Body          node.Body
}

type FunctionReference struct {
	Function    *GenericFunction  // Pointer to the underlying function
	IsImported  bool              // If it is imported from another module
	ModuleName  string
}

// Map from module names to their corresponding function collections
type FunctionStore map[string] FunctionCollection

// Map from names of available functions in the module to their references
type FunctionCollection  map[string] []FunctionReference

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
						IsImported: true,
					})
				}
			}
		}
	}
	// 3. Iterate over all function declarations in the current module
	for _, cmd := range mod.Node.Commands {
		switch c := cmd.Command.(type) {
		case node.DeclFunction:
			// 3.1. Get the names of the function and its type parameters
			var decl = c
			var name = loader.Id2String(decl.Name)
			if name == IgnoreMark {
				// 3.1.1. If the function name is invalid, throw an error.
				return nil, &FunctionError {
					Point:    ErrorPoint { AST: mod.AST, Node: decl.Name.Node },
					Concrete: E_InvalidFunctionName { name },
				}
			}
			var params = make([]string, len(decl.Params))
			for i, p := range decl.Params {
				params[i] = loader.Id2String(p)
			}
			// 3.2. Create a context for evaluating types
			var ctx = TypeContext {
				Module: mod,
				Params: params,
				Ireg:   reg,
			}
			// 3.3. Evaluate the function signature using the created context
			var sig, err = TypeFromRepr(decl.Repr, ctx)
			if err != nil { return nil, &FunctionError {
				Point:    ErrorPoint { AST: mod.AST, Node: decl.Repr.Node },
				Concrete: E_SignatureInvalid {
					FuncName:  name,
					TypeError: err,
				},
			} }
			// 3.4. If the function is public, ensure its signature type
			//        to be a local type of this module.
			var is_public = decl.Public
			if is_public && !(IsLocalType(sig, mod_name)) {
				return nil, &FunctionError {
					Point:    ErrorPoint {
						AST: mod.AST, Node: decl.Repr.Node,
					},
					Concrete: E_SignatureNonLocal {
						FuncName: name,
					},
				}
			}
			// 3.5. Construct a representation and a reference of the function
			var func_type = sig.(AnonymousType).Repr.(Func)
			var gf = &GenericFunction {
				Public:       is_public,
				TypeParams:   params,
				DeclaredType: func_type,
				Body:         decl.Body.Body,
				Node:         decl.Node,
			}
			var ref = FunctionReference {
				IsImported: false,
				Function:   gf,
				ModuleName: mod_name,
			}
			// 3.6. Check if the name of the function is in use
			var _, exists = collection[name]
			if exists {
				// 3.6.1. If in use, try to overload it
				var err_point = ErrorPoint {
					AST: mod.AST, Node: decl.Name.Node,
				}
				var err = CheckOverload (
					collection[name], func_type, name, err_point,
				)
				if err != nil { return nil, err }
				collection[name] = append(collection[name], ref)
			} else {
				// 3.6.2. If not, collect the function
				collection[name] = []FunctionReference { ref }
			}
		}
	}
	// 4. Register all collected functions of the current module to the store
	store[mod_name] = collection
	// 5. Return all collected functions of the current module
	return collection, nil
}


func CheckOverload (
	functions   [] FunctionReference,
	added_type  Func,
	added_name  string,
	err_point   ErrorPoint,
) *FunctionError {
	for _, existing := range functions {
		var cannot_overload = AreTypesOverloadUnsafe (
			AnonymousType { existing.Function.DeclaredType },
			AnonymousType { added_type },
		)
		if cannot_overload { return &FunctionError {
			Point:    err_point,
			Concrete: E_InvalidOverload {
				FunctionName: added_name,
				ModuleName:   existing.ModuleName,
				BetweenLocal: !(existing.IsImported),
			},
		} }
	}
	return nil
}
