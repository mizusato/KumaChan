package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
)

type GenericFunction struct {
	IsGlobal    bool        // Global (a.k.a. exported) or not
	TypeParams  [] string   // Generic Parameters
	FuncType    Func        // Including Input and Output Types
	Body        node.Expr   // Body Expression
	Node        node.Node
}

type FunctionReference struct {
	Function    *GenericFunction
	IsImported  bool
}

type FunctionStore map[string] FunctionCollection

type FunctionCollection  map[string] []FunctionReference

func CollectFunctions (mod *loader.Module, store FunctionStore) (FunctionCollection, *FunctionError) {
	var mod_name = loader.Id2String(mod.Node.Name)
	var existing, exists = store[mod_name]
	if exists {
		return existing, nil
	}
	var collection = make(FunctionCollection)
	for _, imported := range mod.ImpMap {
		var imp_col, err = CollectFunctions(imported, store)
		if err != nil { return nil, err }
		for name, refs := range imp_col {
			var _, exists = collection[name]
			if !exists { collection[name] = make([]FunctionReference, 0) }
			for _, ref := range refs {
				if !ref.IsImported && ref.Function.IsGlobal {
					// conflict must not happen there
					// since global functions have local types
					collection[name] = append(collection[name], FunctionReference {
						Function:   ref.Function,
						IsImported: true,
					})
				}
			}
		}
	}

	store[mod_name] = collection
	return collection, nil
}



