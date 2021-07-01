package checker2

import (
	"strings"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/loader"
)


type SectionCollection (map[*ast.VariousStatement] *source.Section)

func collectSectionMapping(entry *loader.Module) SectionCollection {
	var sc = make(SectionCollection)
	registerSectionMapping(entry, sc)
	return sc
}

func registerSectionMapping(mod *loader.Module, sc SectionCollection) {
	var _ = mod.ForEachImported(func(imported *loader.Module) source.Errors {
		registerSectionMapping(imported, sc)
		return nil
	})
	var sb SectionBuffer
	for i, stmt := range mod.AST.Statements {
		var stmt_ptr = &(mod.AST.Statements[i])
		var title, is_title = stmt.Statement.(ast.Title)
		if is_title {
			sb.SetFrom(title)
		}
		var sec = sb.GetFrom(stmt.Location.File)
		sc[stmt_ptr] = sec
	}
}


type SectionBuffer struct {
	current  *source.Section
}

func (sb *SectionBuffer) SetFrom(title ast.Title) {
	var content = string(title.Content)
	content = strings.TrimPrefix(content, "##")
	content = strings.TrimSuffix(content, "\r")
	content = strings.Trim(content, " ")
	sb.current = &source.Section {
		Name:  content,
		Start: title.Location,
	}
}

func (sb *SectionBuffer) GetFrom(file source.File) *source.Section {
	if sb.current == nil {
		return nil
	} else {
		if file == sb.current.Start.File {
			return sb.current
		} else {
			sb.current = nil
			return nil
		}
	}
}


