package loader

import (
	"io/ioutil"
	"kumachan/parser"
	"kumachan/transformer"
	"kumachan/transformer/node"
	"path/filepath"
	."kumachan/error"
	"strings"
)


type Module struct {
	Node    node.Module
	AST     *parser.Tree
	ImpMap  map[string] *Module
}


func LoadModule(path string, ctx ErrorContext) (*Module, Error) {
	var file_content, err1 = ioutil.ReadFile(path)
	if err1 != nil { return nil, E_ReadFileFailed {
		FilePath: path,
		Message: err1.Error(),
		Context: ctx,
	} }
	var code_string = string(file_content)
	var code = []rune(code_string)
	var ast, err2 = parser.Parse(code, "module", path)
	if err2 != nil { return nil, E_ParseFailed {
		PartialAST: ast,
		ParserError: err2,
		Context: ctx,
	} }
	var module_node = transformer.Transform(ast)
	var imported_map = make(map[string] *Module)
	for _, cmd := range module_node.Commands {
		switch c := cmd.Command.(type) {
		case node.Import:
			var local_alias = string(c.Name.Name)
			var relative_path = strings.Trim(string(c.Path.Value), "'")  // fixme: ugly, should be handled before loader phase
			var imported_path = filepath.Join (
				filepath.Dir(path), relative_path,
			)
			var imported_module, err = LoadModule (
				imported_path,
				ErrorContext {
					ImportPoint: ErrorPoint { AST: ast, Node: c.Node },
					LocalAlias: local_alias,
				},
			)
			if err != nil { return nil, err }
			imported_map[local_alias] = imported_module
		default:
			// do nothing
		}
	}
	return &Module {
		Node: module_node,
		ImpMap: imported_map,
		AST: ast,
	}, nil
}
