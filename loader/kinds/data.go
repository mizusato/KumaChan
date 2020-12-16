package kinds

import (
	"strings"
	"path/filepath"
	"kumachan/parser"
	"kumachan/parser/ast"
	"kumachan/loader/common"
	"kumachan/stdlib"
)


type DataFile struct {
	Path    string
	Public  bool
}
func (f DataFile) GetAST() (ast.Root, *parser.Error) {
	var path = stdlib.ParsePath(f.Path)
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
		stdlib.OS,
		stdlib.PathT,
		path,
	)
	ast_root.Statements = append(ast_root.Statements, const_decl)
	return ast_root, nil
}

type DataConfig struct {
	Public  bool   `json:"public"`
}
func LoadData(path string, i_config interface{}) (common.UnitFile, error) {
	var config = i_config.(DataConfig)
	return DataFile {
		Path:   path,
		Public: config.Public,
	}, nil
}

func DataLoader() common.UnitFileLoader {
	return common.UnitFileLoader {
		Extensions: [] string {
			"bin",  "BIN",
			"txt",  "TXT",
			"html", "HTML",
			"css",  "CSS",
			"js",   "JS",
		},
		Name: "data",
		Load: LoadData,
	}
}

