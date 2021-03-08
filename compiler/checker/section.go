package checker

import (
	"kumachan/compiler/loader/parser/ast"
	"strings"
)


type CurrentSection struct {
	title string
}

func (s *CurrentSection) SetFrom(title ast.Title) {
	var str = string(title.Content)
	str = strings.TrimPrefix(str, "#")
	str = strings.TrimSuffix(str, "\r")
	str = strings.Trim(str, " ")
	s.title = str
}

func (s *CurrentSection) Get() string {
	return s.title
}

