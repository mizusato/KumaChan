package kinds

import (
	"bytes"
	"strings"
    "image/png"
	"path/filepath"
	"kumachan/loader/common"
	"kumachan/parser/ast"
	"kumachan/parser"
	"kumachan/stdlib"
)


type PNG_File struct {
	Path    string
	Data    *stdlib.PNG
	Public  bool
}
func (f PNG_File) GetAST() (ast.Root, *parser.Error) {
	var name = strings.TrimSuffix(filepath.Base(f.Path), filepath.Ext(f.Path))
	var ast_root = common.CreateEmptyAST(f.Path)
	var ast_root_node = ast_root.Node
	var const_decl = common.CreateConstant (
		ast_root_node,
		f.Public,
		name,
		stdlib.Image_M,
		stdlib.PNG_T,
		f.Data,
	)
	ast_root.Statements = append(ast_root.Statements, const_decl)
	return ast_root, nil
}

type PNG_Config struct {
	Public  bool   `json:"public"`
}
func LoadPNG(path string, content ([] byte), raw_config interface{}) (common.UnitFile, error) {
	var config = raw_config.(PNG_Config)
	var reader = bytes.NewReader(content)
	var decoded, err = png.Decode(reader)
	if err != nil { return nil, err }
	return PNG_File {
		Path:   path,
		Data:   &stdlib.PNG {
			RawData: content,
			Decoded: decoded,
		},
		Public: config.Public,
	}, nil
}

func PNG_Loader() common.UnitFileLoader {
	return common.UnitFileLoader {
		Extension: "png",
		Load:      LoadPNG,
	}
}
