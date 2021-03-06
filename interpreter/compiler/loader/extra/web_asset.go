package extra

import (
	"strings"
	"path/filepath"
	"kumachan/interpreter/lang/textual/parser"
	"kumachan/interpreter/lang/textual/ast"
	"kumachan/interpreter/compiler/loader/common"
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
		stdlib.Mod_ui,
		stdlib.AssetFile_T,
		stdlib.AssetFile { Path: path },
	)
	ast_root.Statements = append(ast_root.Statements, const_decl)
	return ast_root, nil
}

type WebAssetConfig struct {
	Public  bool   `json:"public"`
}
func LoadWebAsset(path string, _ ([] byte), config_ interface{}) (common.UnitFile, error) {
	var config = config_.(WebAssetConfig)
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
		Name: "web_asset",
		Load: LoadWebAsset,
		IsResource: true,
		GetMIME: func(path string) string {
			// NOTE: 1 hour wasted on the additional dot
			var ext = strings.TrimPrefix(filepath.Ext(path), ".")
			switch ext {
			case "css", "CSS":   return "text/css"
			case "js", "JS":     return "text/javascript"
			case "ttf", "TTF":   return "font/ttf"
			default:             return "text/plain"
			}
		},
	}
}

