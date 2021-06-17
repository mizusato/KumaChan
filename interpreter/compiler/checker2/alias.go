package checker2

import (
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/loader"
	"kumachan/interpreter/lang/ast"
	"kumachan/stdlib"
)


type AliasRegistry (map[name.Name] AliasDef)

type AliasDef struct {
	From      name.Name
	To        name.Name
	Location  source.Location
}

func NameFrom(id_mod ast.Identifier, id_item ast.Identifier, mod *loader.Module) name.Name {
	var ref_mod = ast.Id2String(id_mod)
	var ref_item = ast.Id2String(id_item)
	if ref_mod == "" {
		var _, is_core_type = coreTypes[ref_item]
		if is_core_type {
			return name.MakeName(stdlib.Mod_core, ref_item)
		} else {
			return name.MakeName(mod.Name, ref_item)
		}
	} else {
		var imported, found = mod.ImpMap[ref_mod]
		if found {
			return name.MakeName(imported.Name, ref_item)
		} else {
			return name.MakeName(ref_mod, ref_item)
		}
	}
}

func collectAlias(entry *loader.Module) (AliasRegistry, *source.Error) {
	var reg = make(AliasRegistry)
	var err = registerAlias(entry, reg)
	if err != nil { return nil, err }
	for _, def := range reg {
		var _, to_alias = reg[def.To]
		if to_alias {
			return nil, source.MakeError(def.Location, E_InvalidAlias {
				Which: def.From.String(),
			})
		}
	}
	// validation: alias to another alias
	return reg, nil
}

func registerAlias(mod *loader.Module, reg AliasRegistry) *source.Error {
	for _, stmt := range mod.AST.Statements {
		var alias, is_alias = stmt.Statement.(ast.Alias)
		if is_alias {
			var from = name.MakeName(mod.Name, ast.Id2String(alias.Name))
			var to = NameFrom(alias.Module, alias.Item, mod)
			var _, exists = reg[from]
			if exists {
				return source.MakeError(alias.Name.Location, E_DuplicateAlias {
					Which: from.String(),
				})
			}
			reg[from] = AliasDef {
				From:     from,
				To:       to,
				Location: alias.Location,
			}
		}
	}
	for _, imported := range mod.ImpMap {
		var err = registerAlias(imported, reg)
		if err != nil { return err }
	}
	return nil
}


