package checker2

import (
	"kumachan/interpreter/compiler/loader"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/source"
)


type ModuleInfo struct {
	ModName      string
	ModImported  map[string] *ModuleInfo
}
type ModuleInfoCollection (map[*loader.Module] *ModuleInfo)

func collectModuleInfo(entry *loader.Module) (*ModuleInfo, ModuleInfoCollection) {
	var mic = make(ModuleInfoCollection)
	var entry_info = registerModuleInfo(entry, mic)
	return entry_info, mic
}
func registerModuleInfo(mod *loader.Module, mic ModuleInfoCollection) *ModuleInfo {
	var existing, exists = mic[mod]
	if exists {
		return existing
	}
	var mod_name = mod.Name
	var all_imported = make(map[string] *ModuleInfo)
	for preferred_name, imported := range mod.ImpMap {
		all_imported[preferred_name] = registerModuleInfo(imported, mic)
	}
	return &ModuleInfo {
		ModName:     mod_name,
		ModImported: all_imported,
	}
}


type ModuleVisitedSet (map[*loader.Module] struct{})
func (mvs ModuleVisitedSet) Visited(mod *loader.Module) bool {
	var _, visited = mvs[mod]
	if !(visited) {
		mvs[mod] = struct{}{}
	}
	return visited
}

func TraverseStatements (
	mod  *loader.Module,
	mic  ModuleInfoCollection,
	sc   SectionCollection,
	mvs  ModuleVisitedSet,
	f    func(stmt ast.VariousStatement, sec *source.Section, mi *ModuleInfo) *source.Error,
) source.Errors {
	if mvs.Visited(mod) {
		return nil
	}
	{ var err = mod.ForEachImported(func(imported *loader.Module) source.Errors {
		return TraverseStatements(imported, mic, sc, mvs, f)
	})
	if err != nil {
		return err
	} }
	var errs source.Errors
	var mi = mic[mod]
	for i, stmt := range mod.AST.Statements {
		var _, is_title = stmt.Statement.(ast.Title)
		if !(is_title) {
			var stmt_ptr = &(mod.AST.Statements[i])
			var sec, sec_exists = sc[stmt_ptr]
			if !(sec_exists) { panic("something went wrong") }
			source.ErrorsJoin(&errs, f(stmt, sec, mi))
		}
	}
	return errs
}


