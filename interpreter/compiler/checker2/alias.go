package checker2

import (
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/loader"
	"kumachan/interpreter/lang/ast"
)


type AliasRegistry (map[name.Name] AliasDef)

type AliasDef struct {
	From      name.Name
	To        name.Name
	Section   *source.Section
	Location  source.Location
}

func collectAlias (
	entry  *loader.Module,
	mic    ModuleInfoCollection,
	sc     SectionCollection,
) (AliasRegistry, source.Errors) {
	var reg = make(AliasRegistry)
	var mvs = make(ModuleVisitedSet)
	{ var err = registerAlias(entry, mic, sc, mvs, reg)
	if err != nil {
		return nil, err
	} }
	var errs source.Errors
	for _, def := range reg {
		var _, to_alias = reg[def.To]
		if to_alias {
			var err = source.MakeError(def.Location,
				E_InvalidAlias { Which: def.From.String() })
			source.ErrorsJoin(&errs, err)
		}
	}
	if errs != nil {
		return nil, errs
	}
	// validation: alias to another alias
	return reg, nil
}

func registerAlias (
	mod  *loader.Module,
	mic  ModuleInfoCollection,
	sc   SectionCollection,
	mvs  ModuleVisitedSet,
	reg  AliasRegistry,
) source.Errors {
	return traverseStatements(mod, mic, sc, mvs, func(stmt ast.VariousStatement, sec *source.Section, mi *ModuleInfo) *source.Error {
		var alias, is_alias = stmt.Statement.(ast.Alias)
		if !(is_alias) { return nil }
		var from = name.MakeName(mod.Name, ast.Id2String(alias.Name))
		// TODO: validate 'from' name
		var to = NameFrom(alias.Module, alias.Item, mi)
		var _, exists = reg[from]
		if exists {
			return source.MakeError(alias.Name.Location,
				E_DuplicateAlias { Which: from.String() })
		}
		reg[from] = AliasDef {
			From:     from,
			To:       to,
			Section:  sec,
			Location: alias.Location,
		}
		return nil
	})
}


