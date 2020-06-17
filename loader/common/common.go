package common

import (
	"kumachan/parser/ast"
	"kumachan/parser"
)


type UnitFileLoader struct {
	Extension  string
	Load       func(path string, content ([] byte)) (UnitFile, error)
}

type UnitFile interface {
	GetAST() (ast.Root, *parser.Error)
}
