package kinds

import (
	"strings"
	"path/filepath"
	"kumachan/parser"
	"kumachan/parser/ast"
	"kumachan/loader/common"
	"kumachan/stdlib"
)


type WebAssetFile struct {
	Path    string
	Public  bool
}
func (f WebAssetFile) GetAST() (ast.Root, *parser.Error) {
	var path = f.Path
	var ext = filepath.Ext(path)
	var name_base = strings.TrimSuffix(filepath.Base(path), ext)
	var name_ext = strings.TrimPrefix(ext, ".")
	var name = strings.ReplaceAll(name_base, ".", "-") + "-" + name_ext
	var ast_root = common.CreateEmptyAST(path)
	var ast_root_node = ast_root.Node
	var const_decl = common.CreateConstant (
		ast_root_node,
		f.Public,
		name,
		stdlib.WebUi,
		stdlib.WebUi_ResourceFile,
		stdlib.WebUiResourceFile { Path: path },
	)
	ast_root.Statements = append(ast_root.Statements, const_decl)
	return ast_root, nil
}

type WebAssetConfig struct {
	Public  bool   `json:"public"`
}
func LoadWebAsset(path string, _ ([] byte), i_config interface{}) (common.UnitFile, error) {
	var config = i_config.(WebAssetConfig)
	return WebAssetFile {
		Path:   path,
		Public: config.Public,
	}, nil
}

func WebAssetLoader() common.UnitFileLoader {
	return common.UnitFileLoader {
		Extensions: [] string {
			"css",   "CSS",
			"js",    "JS",
			"ttf",   "TTF",
		},
		Name: "webAsset",
		Load: LoadWebAsset,
		IsResource: true,
		GetMIME: func(path string) string {
			var ext = filepath.Ext(path)
			switch ext {
			case "css", "CSS":   return "text/css"
			case "js", "JS":     return "text/javascript"
			case "ttf", "TTF":   return "font/ttf"
			default:             return "text/plain"
			}
		},
	}
}

