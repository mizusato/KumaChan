package kinds

import (
	"strings"
	"path/filepath"
	"kumachan/parser"
	"kumachan/parser/ast"
	"kumachan/loader/common"
	"kumachan/stdlib"
	"unicode/utf8"
)


type TextFile struct {
	Path    string
	Data    [] byte
	Public  bool
}
func (f TextFile) GetAST() (ast.Root, *parser.Error) {
	var bytes = f.Data
	var str = make([] uint32, 0, len(bytes) / 4)
	for len(bytes) > 0 {
		var char, size = utf8.DecodeRune(bytes)
		str = append(str, uint32(char))
		bytes = bytes[size:]
	}
	var ext = filepath.Ext(f.Path)
	var name_base = strings.TrimSuffix(filepath.Base(f.Path), ext)
	var name_ext = strings.TrimPrefix(ext, ".")
	var name = name_base + "-" + name_ext
	var ast_root = common.CreateEmptyAST(f.Path)
	var ast_root_node = ast_root.Node
	var const_decl = common.CreateConstant (
		ast_root_node,
		f.Public,
		name,
		stdlib.Core,
		stdlib.String,
		str,
	)
	ast_root.Statements = append(ast_root.Statements, const_decl)
	return ast_root, nil
}

type TextConfig struct {
	Public  bool   `json:"public"`
}
func LoadText(path string, content ([] byte), i_config interface{}) (common.UnitFile, error) {
	var config = i_config.(TextConfig)
	return TextFile {
		Path:   path,
		Data:   content,
		Public: config.Public,
	}, nil
}

func TextLoader() common.UnitFileLoader {
	return common.UnitFileLoader {
		Extensions: [] string {
			"txt",  "TXT",
			"html", "HTML",
			"css",  "CSS",
			"js",   "JS",
		},
		Name: "text",
		Load: LoadText,
	}
}

