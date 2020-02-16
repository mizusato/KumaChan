package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
	. "kumachan/error"
)

type GenericFunction struct {
	IsGlobal    bool        // Global (a.k.a. exported) or not
	TypeParams  [] string   // Generic Parameters
	FuncType    Func        // Including Input and Output Types
	Body        node.Body   // Body
	Node        node.Node   // To Generate ErrorPoint
}

type FunctionReference struct {
	Function    *GenericFunction
	IsImported  bool
}

type FunctionStore map[string] FunctionCollection

type FunctionCollection  map[string] []FunctionReference

func CollectFunctions (mod *loader.Module, reg TypeRegistry, store FunctionStore) (FunctionCollection, *FunctionError) {
	var mod_name = loader.Id2String(mod.Node.Name)
	var existing, exists = store[mod_name]
	if exists {
		return existing, nil
	}
	var collection = make(FunctionCollection)
	for _, imported := range mod.ImpMap {
		var imp_col, err = CollectFunctions(imported, reg, store)
		if err != nil { return nil, err }
		for name, refs := range imp_col {
			for _, ref := range refs {
				if !ref.IsImported && ref.Function.IsGlobal {
					var _, exists = collection[name]
					if !exists { collection[name] = make([]FunctionReference, 0) }
					// conflict (unsafe overload) must not happen there,
					// since global functions have local types
					collection[name] = append(collection[name], FunctionReference {
						Function:   ref.Function,
						IsImported: true,
					})
				}
			}
		}
	}
	for _, cmd := range mod.Node.Commands {
		switch c := cmd.Command.(type) {
		case node.DeclFunction:
			var decl = c
			var name = loader.Id2String(decl.Name)
			var params = make([]string, len(decl.Params))
			for i, p := range decl.Params {
				params[i] = loader.Id2String(p)
			}
			var ctx = TypeExprContext {
				Module: mod,
				Params: params,
				Ireg:   reg,
			}
			var sig, err = TypeExprFromRepr(decl.Repr, ctx)
			if err != nil { return nil, &FunctionError {
				Point:    ErrorPoint { AST: mod.AST, Node: decl.Repr.Node },
				Concrete: E_SignatureInvalid {
					FuncName:  name,
					TypeError: err,
				},
			} }
			var is_global = decl.IsGlobal
			if is_global && !(IsLocalType(sig, mod_name)) { return nil, &FunctionError {
				Point:    ErrorPoint { AST: mod.AST, Node: decl.Repr.Node },
				Concrete: E_SignatureNonLocal {
					FuncName: name,
				},
			} }
			var func_type = sig.(AnonymousType).Repr.(Func)
			var gf = &GenericFunction {
				IsGlobal:   is_global,
				TypeParams: params,
				FuncType:   func_type,
				Body:       decl.Body.Body,
				Node:       decl.Node,
			}
			var ref = FunctionReference {
				IsImported: false,
				Function: gf,
			}
			var _, exists = collection[name]
			if exists {
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
				collection[name] = []FunctionReference { ref }
			}
		}
	}
	store[mod_name] = collection
	return collection, nil
}



