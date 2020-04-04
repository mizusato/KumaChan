package checker

import (
	"kumachan/loader"
	"kumachan/transformer/ast"
	. "kumachan/error"
)


type Constant struct {
	Node          ast.Node
	Public        bool
	DeclaredType  Type
	Value         ast.ConstValue
}

type ConstantCollection  map[loader.Symbol] *Constant

type ConstantStore  map[string] ConstantCollection


func CollectConstants(mod *loader.Module, reg TypeRegistry, store ConstantStore) (ConstantCollection, *ConstantError) {
	var mod_name = mod.Name
	var existing, exists = store[mod_name]
	if exists {
		return existing, nil
	}
	var collection = make(ConstantCollection)
	for _, imported := range mod.ImpMap {
		var imp_mod_name = imported.Name
		var imp_col, err = CollectConstants(imported, reg, store)
		if err != nil { return nil, err }
		for name, constant := range imp_col {
			if name.ModuleName == imp_mod_name && constant.Public {
				var _, exists = collection[name]
				if exists { panic("something went wrong") }
				collection[name] = constant
			}
		}
	}
	for _, cmd := range mod.Node.Commands {
		switch decl := cmd.Command.(type) {
		case ast.DeclConst:
			var name = mod.SymbolFromName(decl.Name)
			if name.SymbolName == IgnoreMark {
				return nil, &ConstantError {
					Point:    ErrorPoint { CST: mod.CST, Node: decl.Name.Node },
					Concrete: E_InvalidConstName { name.SymbolName },
				}
			}
			var _, exists = collection[name]
			if exists { return nil, &ConstantError {
				Point:    ErrorPoint { CST: mod.CST, Node: decl.Name.Node },
				Concrete: E_DuplicateConstDecl {
					Name: name.SymbolName,
				},
			} }
			exists, _ = reg.LookupArity(name)
			if exists { return nil, &ConstantError {
				Point: ErrorPoint { CST: mod.CST, Node: decl.Name.Node, },
				Concrete: E_ConstConflictWithType {
					Name: name.SymbolName,
				},
			} }
			var ctx = TypeContext {
				Module: mod,
				Params: make([]string, 0),
				Ireg:   reg,
			}
			var is_public = decl.IsPublic
			var declared_type, err = TypeFrom(decl.Type.Type, ctx)
			if err != nil { return nil, &ConstantError {
				Point:    ErrorPoint { CST: mod.CST, Node: decl.Type.Node },
				Concrete: E_ConstTypeInvalid {
					ConstName: name.String(),
					TypeError: err,
				},
			} }
			var value = decl.Value.ConstValue
			var constant = &Constant {
				Node:         decl.Node,
				Public:       is_public,
				DeclaredType: declared_type,
				Value:        value,
			}
			collection[name] = constant
		}
	}
	store[mod_name] = collection
	return collection, nil
}
