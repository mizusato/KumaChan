package common

import (
	"kumachan/parser"
	"kumachan/parser/ast"
	"kumachan/parser/syntax"
	"kumachan/parser/transformer"
)


type UnitFileLoader struct {
	Extension  string
	Load       func(path string, content ([] byte), config interface{}) (UnitFile, error)
}

type UnitFile interface {
	GetAST() (ast.Root, *parser.Error)
}

func CreateEmptyAST(name string) ast.Root {
	var empty_cst, err = parser.Parse(([] rune)(""), syntax.RootPartName, name)
	if err != nil { panic("something went wrong") }
	var empty_ast = transformer.Transform(empty_cst)
	return empty_ast
}

func CreateConstant (
	dummy_node  ast.Node,
	public      bool,
	name        string,
	type_mod    string,
	type_name   string,
	value       interface{},
) (ast.VariousStatement) {
	var id = func(id_str string) ast.Identifier {
		return ast.Identifier {
			Node: dummy_node,
			Name: ([] rune)(id_str),
		}
	}
	return ast.VariousStatement {
		Node:      dummy_node,
		Statement: ast.DeclConst{
			Node:   dummy_node,
			Public: public,
			Name:   id(name),
			Type:   ast.VariousType {
				Node: dummy_node,
				Type: ast.TypeRef{
					Node:     dummy_node,
					Module:   id(type_mod),
					Specific: true,
					Id:       id(type_name),
					TypeArgs: make([] ast.VariousType, 0),
				},
			},
			Value:  ast.VariousConstValue {
				Node:       dummy_node,
				ConstValue: ast.PredefinedValue { Value: value },
			},
		},
	}
}
