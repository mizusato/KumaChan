package kinds

import (
	"errors"
	"encoding/xml"
	"path/filepath"
	"kumachan/loader/common"
	"kumachan/parser"
	"kumachan/parser/ast"
	"kumachan/qt"
	"strings"
)


const QtModName = "Qt"
var __QtSupportedWidgetClassList = [] string {
	"Widget", "MainWindow", "Label", "LineEdit", "PushButton",
}
var __QtSupportedWidgetClassMap = (func() map[string] bool {
	var m = make(map[string] bool)
	for _, name := range __QtSupportedWidgetClassList {
		m[name] = true
	}
	return m
})()
func QtIsSupportedWidgetClass(name string) bool {
	return __QtSupportedWidgetClassMap[name]
}

type QtUiFile struct {
	Path     string
	Widgets  map[string] QtWidget
	Config   QtUiConfig
}
type QtWidget struct {
	Class     string
	Instance  qt.Widget
}
type QtUiConfig struct {
	Public  bool   `json:"public"`
}
func (f QtUiFile) GetAST() (ast.Root, *parser.Error) {
	var ast_root = common.CreateEmptyAST(f.Path)
	var ast_root_node = ast_root.Node
	for name, widget := range f.Widgets {
		var normalized_class = strings.TrimPrefix(widget.Class, "Q")
		if (!(QtIsSupportedWidgetClass(normalized_class))) {
			continue
		}
		var const_decl = common.CreateConstant (
			ast_root_node,
			f.Config.Public,
			name,
			QtModName,
			normalized_class,
			widget.Instance,
		)
		ast_root.Statements = append(ast_root.Statements, const_decl)
	}
	return ast_root, nil
}

type QtUi struct {
	RootWidget  QtUiWidget   `xml:"widget"`
}

type QtUiWidget struct {
	Name      string          `xml:"name,attr"`
	Class     string          `xml:"class,attr"`
	Children  [] QtUiWidget   `xml:"widget"`
}

func LoadQtUi(path string, content ([] byte), raw_config interface{}) (common.UnitFile, error) {
	var config, _ = raw_config.(QtUiConfig)
	var ui QtUi
	var err = xml.Unmarshal(content, &ui)
	if err != nil { return nil, err }
	qt.MakeSureInitialized()
	var root_instance, ok = qt.LoadWidget(string(content), filepath.Dir(path))
	if !ok { return QtUiFile{}, errors.New("error parsing UI file") }
	var widgets = make(map[string] QtWidget)
	widgets[ui.RootWidget.Name] = QtWidget {
		Class:    ui.RootWidget.Class,
		Instance: root_instance,
	}
	var add_children func(def QtUiWidget, instance qt.Widget)
	add_children = func(def QtUiWidget, instance qt.Widget) {
		for _, child := range def.Children {
			var name = child.Name
			var child_instance, ok = qt.FindChild(instance, name)
			if !ok { panic("something went wrong") }
			widgets[name] = QtWidget {
				Class:    def.Class,
				Instance: child_instance,
			}
			add_children(child, child_instance)
		}
	}
	add_children(ui.RootWidget, root_instance)
	return QtUiFile {
		Path:    path,
		Widgets: widgets,
		Config:  config,
	}, nil
}

func QtUiLoader() common.UnitFileLoader {
	return common.UnitFileLoader {
		Extension: "ui",
		Load:      LoadQtUi,
	}
}
