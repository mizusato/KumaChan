package loader

import (
	"kumachan/loader/parser"
	"kumachan/loader/parser/ast"
	"kumachan/loader/parser/syntax"
	"kumachan/loader/parser/transformer"
	"kumachan/loader/common"
)


type Structure interface {
	Load(ctx Context)  (ast.Root, *Error)
}

type PredefinedAST struct {
	Root  ast.Root
}
func (m PredefinedAST) Load(Context) (ast.Root, *Error) {
	return m.Root, nil
}

type StandaloneScript struct {
	File  SourceFile
}
func (mod StandaloneScript) Load(ctx Context) (ast.Root, *Error) {
	var root, err = mod.File.GetAST()
	if err != nil { return ast.Root{}, &Error {
		Context:  ctx,
		Concrete: E_ParseFailed {
			ParserError: err,
		},
	} }
	return root, nil
}

type ModuleFolder struct {
	Files  [] common.UnitFile
}
type SourceFile struct /* implements common.UnitFile */ {
	Path     string
	Content  [] byte
}
func (mod ModuleFolder) Load(ctx Context) (ast.Root, *Error) {
	var ast_root = common.CreateEmptyAST("(Module Folder)")
	for _, f := range mod.Files {
		var f_root, err = f.GetAST()
		if err != nil { return ast.Root{}, &Error {
			Context:  ctx,
			Concrete: E_ParseFailed {
				ParserError: err,
			},
		} }
		for _, cmd := range f_root.Statements {
			ast_root.Statements = append(ast_root.Statements, cmd)
		}
	}
	return ast_root, nil
}
func (sf SourceFile) GetAST() (ast.Root, *parser.Error) {
	var code_string = string(sf.Content)
	var code = ([] rune)(code_string)
	var tree, err = parser.Parse(code, syntax.RootPartName, sf.Path)
	if err != nil { return ast.Root{}, err }
	return transformer.Transform(tree).(ast.Root), nil
}

