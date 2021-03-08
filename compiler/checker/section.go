package checker

import (
	"strings"
	"kumachan/compiler/loader/parser/cst"
	"kumachan/compiler/loader/parser/ast"
)


type CurrentSection struct {
	cst    *cst.Tree
	title  string
}

func (s *CurrentSection) SetFrom(title ast.Title) {
	var str = string(title.Content)
	str = strings.TrimPrefix(str, "##")
	str = strings.TrimSuffix(str, "\r")
	str = strings.Trim(str, " ")
	s.title = str
	s.cst = title.Node.CST
}

func (s *CurrentSection) GetAt(node ast.Node) string {
	if node.CST != s.cst {
		s.cst = node.CST
		s.title = ""
	}
	return s.title
}

