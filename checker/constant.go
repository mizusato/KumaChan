package checker

import (
	"kumachan/loader"
	"kumachan/transformer/node"
	. "kumachan/error"
)

type Constant struct {
	Node          node.Node
	IsPublic      bool
	DeclaredType  Type
	Value         node.ConstValue
}

type ConstantCollection  map[loader.Symbol] *Constant

type ConstantStore  map[string] ConstantCollection

func CollectConstants(mod *loader.Module, reg TypeRegistry, store ConstantStore) (ConstantCollection, *ConstantError) {
	var mod_name = loader.Id2String(mod.Node.Name)
	var existing, exists = store[mod_name]
	if exists {
		return existing, nil
	}
	var collection = make(ConstantCollection)
	for _, imported := range mod.ImpMap {
		var imp_mod_name = loader.Id2String(imported.Node.Name)
		var imp_col, err = CollectConstants(imported, reg, store)
		if err != nil { return nil, err }
		for name, constant := range imp_col {
			if name.ModuleName == imp_mod_name && constant.IsPublic {
				var _, exists = collection[name]
				if exists { panic("something went wrong") }
				collection[name] = constant
			}
		}
	}
	for _, cmd := range mod.Node.Commands {
		switch c := cmd.Command.(type) {
		case node.DeclConst:
			var decl = c
			var name = mod.SymbolFromName(decl.Name)
			if name.SymbolName == IgnoreMark {
				return nil, &ConstantError {
					Point:    ErrorPoint { AST: mod.AST, Node: c.Name.Node },
					Concrete: E_InvalidConstName { name.SymbolName },
				}
			}
			var _, exists = collection[name]
			if exists { return nil, &ConstantError {
				Point:    ErrorPoint { AST: mod.AST, Node: c.Name.Node },
				Concrete: E_DuplicateConstDecl {
					Name: name.SymbolName,
				},
			} }
			exists, _ = reg.LookupArity(name)
			if exists { return nil, &ConstantError {
				Point: ErrorPoint { AST:  mod.AST, Node: c.Name.Node, },
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
				Point:    ErrorPoint { AST: mod.AST, Node: c.Type.Node },
				Concrete: E_ConstTypeInvalid {
					ConstName: name.String(),
					TypeError: err,
				},
			} }
			var value = decl.Value.ConstValue
			var constant = &Constant {
				Node:         c.Node,
				IsPublic:     is_public,
				DeclaredType: declared_type,
				Value:        value,
			}
			collection[name] = constant
		}
	}
	store[mod_name] = collection
	return collection, nil
}
