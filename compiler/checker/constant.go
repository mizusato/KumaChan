package checker

import (
	"kumachan/compiler/loader"
	"kumachan/compiler/loader/parser/ast"
	. "kumachan/util/error"
)


type Constant struct {
	Section       string
	Node          ast.Node
	Doc           string
	Public        bool
	DeclaredType  Type
	Value         ast.ConstValue
}
type ConstantCollection  map[loader.Symbol] *Constant
type ConstantStore       map[string] ConstantCollection

var __NoParams = make([] TypeParam, 0)
var __NoBounds = TypeBounds {
	Sub:   make(map[uint] Type),
	Super: make(map[uint] Type),
}


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
	var section CurrentSection
	for _, stmt := range mod.AST.Statements {
		switch decl := stmt.Statement.(type) {
		case ast.Title:
			section.SetFrom(decl)
		case ast.DeclConst:
			var name = mod.SymbolFromDeclName(decl.Name)
			var doc = DocStringFromRaw(decl.Docs)
			if name.SymbolName == IgnoreMark {
				return nil, &ConstantError {
					Point:    ErrorPointFrom(decl.Name.Node),
					Concrete: E_InvalidConstName { name.SymbolName },
				}
			}
			var _, exists = collection[name]
			if exists { return nil, &ConstantError {
				Point:    ErrorPointFrom(decl.Name.Node),
				Concrete: E_DuplicateConstDecl {
					Name: name.SymbolName,
				},
			} }
			_, exists = reg[name]
			if exists { return nil, &ConstantError {
				Point: ErrorPointFrom(decl.Name.Node),
				Concrete: E_ConstConflictWithType {
					Name: name.SymbolName,
				},
			} }
			var ctx = TypeContext {
				TypeBoundsContext: TypeBoundsContext {
					TypeValidationContext: TypeValidationContext {
						TypeConstructContext: TypeConstructContext {
							Module:     mod,
							Parameters: __NoParams,
						},
						Registry: reg,
					},
					Bounds: __NoBounds,
				},
			}
			var is_public = decl.Public
			var declared_type, err = TypeFrom(decl.Type, ctx)
			if err != nil { return nil, &ConstantError {
				Point:    ErrorPointFrom(decl.Type.Node),
				Concrete: E_ConstTypeInvalid {
					ConstName: name.String(),
					TypeError: err,
				},
			} }
			var value = decl.Value.ConstValue
			if value == nil { return nil, &ConstantError {
				Point:    ErrorPointFrom(decl.Name.Node),
				Concrete: E_MissingConstantDefinition {
					ConstName: name.SymbolName,
				},
			} }
			var constant = &Constant {
				Section:      section.Get(),
				Node:         decl.Node,
				Doc:          doc,
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
