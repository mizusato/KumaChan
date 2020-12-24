package kinds

import (
	"fmt"
	"errors"
	"strings"
	"path/filepath"
	"kumachan/parser"
	"kumachan/parser/ast"
	"kumachan/loader/common"
	"kumachan/stdlib"
)


type WebAssetFile struct {
	AbsPath  string
	Public   bool
}
func (f WebAssetFile) GetAST() (ast.Root, *parser.Error) {
	// TODO: should handle relative path to the entry module
	//       and provide custom url handler at Qt side
	var path = stdlib.ParsePath(f.AbsPath)
	var ext = filepath.Ext(f.AbsPath)
	var name_base = strings.TrimSuffix(filepath.Base(f.AbsPath), ext)
	var name_ext = strings.TrimPrefix(ext, ".")
	var name = strings.ReplaceAll(name_base, ".", "-") + "-" + name_ext
	var ast_root = common.CreateEmptyAST(f.AbsPath)
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

type WebAssetConfig struct {
	Public  bool   `json:"public"`
}
func LoadWebAsset(path string, content ([] byte), i_config interface{}) (common.UnitFile, error) {
	var config = i_config.(WebAssetConfig)
	var abs_path, err = filepath.Abs(path)
	if err != nil { return nil, errors.New(fmt.Sprintf(
		"cannot get absolute path of: %s", path)) }
	return WebAssetFile {
		AbsPath: abs_path,
		Public:  config.Public,
	}, nil
}

func WebAssetLoader() common.UnitFileLoader {
	return common.UnitFileLoader {
		Extensions: [] string {
			"html", "HTML",
			"css",  "CSS",
			"js",   "JS",
		},
		Name: "webAsset",
		Load: LoadWebAsset,
		ReadContent: false,
	}
}

