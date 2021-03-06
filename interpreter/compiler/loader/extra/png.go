package extra

import (
	"bytes"
	"strings"
    "image/png"
	"path/filepath"
	"kumachan/interpreter/compiler/loader/common"
	"kumachan/interpreter/lang/textual/ast"
	"kumachan/interpreter/lang/textual/parser"
	"kumachan/stdlib"
)


type PNG_File struct {
	Path    string
	Data    interface {}
	Public  bool
	Decode  bool
}
func (f PNG_File) GetAST() (ast.Root, *parser.Error) {
	var name = strings.TrimSuffix(filepath.Base(f.Path), filepath.Ext(f.Path))
	name = strings.ReplaceAll(name, ".", "-")
	var ast_root = common.CreateEmptyAST(f.Path)
	var ast_root_node = ast_root.Node
	var type_name string
	if f.Decode {
		type_name = stdlib.RawImage_T
	} else {
		type_name = stdlib.PNG_T
	}
	var const_decl = common.CreateConstant (
		ast_root_node,
		f.Public,
		name,
		stdlib.Image_M,
		type_name,
		f.Data,
	)
	ast_root.Statements = append(ast_root.Statements, const_decl)
	return ast_root, nil
}

type PNG_Config struct {
	Public  bool   `json:"public"`
	Decode  bool   `json:"decode"`
}
func LoadPNG(path string, content ([] byte), config_ interface{}) (common.UnitFile, error) {
	var config = config_.(PNG_Config)
	if config.Decode {
		var reader = bytes.NewReader(content)
		var decoded, err = png.Decode(reader)
		if err != nil { return nil, err }
		return PNG_File {
			Path:   path,
			Data:   &stdlib.RawImage { Data: decoded },
			Public: config.Public,
			Decode: config.Decode,
		}, nil
	} else {
		return PNG_File {
			Path:   path,
			Data:   &stdlib.PNG { Data: content },
			Public: config.Public,
			Decode: config.Decode,
		}, nil
	}
}

func PNG_Loader() common.UnitFileLoader {
	return common.UnitFileLoader {
		Extensions: [] string { "png" },
		Name:       "png",
		Load:       LoadPNG,
		IsResource: true,
		GetMIME:    func(_ string) string { return "image/png" },
	}
}
