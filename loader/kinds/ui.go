package kinds

import (
	"kumachan/loader/common"
	"encoding/xml"
	"kumachan/parser/ast"
	"kumachan/parser"
	"kumachan/qt"
)


type QtUiFile struct {
	Path     string
	Widgets  [] qt.Widget
	// TODO: config private/public
}
func (f QtUiFile) GetAST() (ast.Root, *parser.Error) {
	panic("not implemented")  // TODO
}

type UI struct {
	Widgets  [] Widget   `xml:"widget"`
}

type Widget struct {
	Name      string      `xml:"name,attr"`
	Class     string      `xml:"class,attr"`
	Children  [] Widget   `xml:"widget"`
}

func LoadQtUi(path string, content ([] byte)) (common.UnitFile, error) {
	var ui UI
	var err = xml.Unmarshal(content, &ui)
	if err != nil { return nil, err }
	qt.MakeSureInitialized()
	// TODO: QUiLoader::setWorkingDirectory
	return QtUiFile {
		Path:    path,
		// TODO
	}, nil
}

func QtUiLoader() common.UnitFileLoader {
	return common.UnitFileLoader{
		Extension: "ui",
		Load:      LoadQtUi,
	}
}
