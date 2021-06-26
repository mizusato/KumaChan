package checker2

import (
	"strings"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/lang/common/source"
)


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


