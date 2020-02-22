package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
	. "kumachan/error"
)

type GenericFunction struct {
	IsPublic    bool      // Public (a.k.a. exported) or not
	TypeParams  [] string // Generic Parameters
	FuncType    Func      // Including Input and Output Types
	Body        node.Body // Body
	Node        node.Node // To Generate ErrorPoint
}

type FunctionReference struct {
	Function    *GenericFunction  // Pointer to the underlying function
	IsImported  bool              // If it is imported from another module
}

// Map from module names to their corresponding function collections
type FunctionStore map[string] FunctionCollection

// Map from names of available functions in the module to their references
type FunctionCollection  map[string] []FunctionReference

// Procedure to collect all functions in a module hierarchy
func CollectFunctions (mod *loader.Module, reg TypeRegistry, store FunctionStore) (FunctionCollection, *FunctionError) {
	/**
	 *  Input: a root module, a type registry and an empty function store
	 *  Output: collected functions of the root module (or an error)
	 *  Effect: fill all collected functions into the function store
	 */
	// 1. Check if the module was visited, if so, return the existing result
	var mod_name = loader.Id2String(mod.Node.Name)
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
				if !ref.IsImported && ref.Function.IsPublic {
					// 2.2.1. If the function is exported, import it
					var _, exists = collection[name]
					if !exists { collection[name] = make([]FunctionReference, 0) }
					// Note: conflict (unsafe overload) must not happen there,
					//       since public functions have local types
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
			// 3.4. If the function is exported, check if it is autognostic
			var is_public = decl.IsPublic
			if is_public && !(IsLocalType(sig, mod_name)) { return nil, &FunctionError {
				Point:    ErrorPoint { AST: mod.AST, Node: decl.Repr.Node },
				Concrete: E_SignatureNonLocal {
					FuncName: name,
				},
			} }
			// 3.5. Construct a representation and a reference of the function
			var func_type = sig.(AnonymousType).Repr.(Func)
			var gf = &GenericFunction {
				IsPublic:   is_public,
				TypeParams: params,
				FuncType:   func_type,
				Body:       decl.Body.Body,
				Node:       decl.Node,
			}
			var ref = FunctionReference {
				IsImported: false,
				Function: gf,
			}
			// 3.6. Check if the name of the function is in use
			var _, exists = collection[name]
			if exists {
				// 3.6.1. If in use, try to overload it
				for _, existing := range collection[name] {
					var unsafe = AreTypesOverloadUnsafe (
						AnonymousType { Repr: existing.Function.FuncType },
						AnonymousType { Repr: func_type },
					)
					if unsafe { return nil, &FunctionError {
						Point:    ErrorPoint { AST: mod.AST, Node: decl.Name.Node },
						Concrete: E_InvalidOverload {
							FuncName:        name,
							IsLocalConflict: !(existing.IsImported),
						},
					} }
					collection[name] = append(collection[name], ref)
				}
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



