package extra

import (
	"errors"
	"strings"
	"encoding/xml"
	"path/filepath"
	"kumachan/lang/parser"
	"kumachan/lang/parser/ast"
	"kumachan/compiler/loader/common"
	"kumachan/runtime/lib/ui/qt"  // TODO: defer dependency to runtime packages
	"kumachan/stdlib"
)


var __QtWidgetDefaultNames = map[string] ([] string) {
	"Widget": { "widget", "centralWidget" },
	"MainWindow": { "MainWindow" },
	"Label": { "label" },
	"LineEdit": { "input" },
	"PlainTextEdit": { "plainTextEdit" },
	"PushButton": { "button" },
	"CheckBox": { "checkBox" },
	"ComboBox": { "comboBox" },
	"ListWidget": { "listWidget" },
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
func (f QtUiFile) GetAST() (ast.Root, *parser.Error) {
	var ast_root = common.CreateEmptyAST(f.Path)
	var ast_root_node = ast_root.Node
	outer: for name, widget := range f.Widgets {
		var normalized_class = strings.TrimPrefix(widget.Class, "Q")
		var default_names = __QtWidgetDefaultNames[normalized_class]
		for _, default_name := range default_names {
			if name == default_name || strings.HasPrefix(name, (default_name + ")")) {
				continue outer
			}
		}
		var type_name, exists = stdlib.GetQtWidgetTypeName(normalized_class)
		if !(exists) {
			continue
		}
		var const_decl = common.CreateConstant (
			ast_root_node,
			f.Config.Public,
			name,
			stdlib.Mod_ui,
			type_name,
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
	Layout    QtUiLayout      `xml:"layout"`
	Actions   [] QtUiAction   `xml:"action"`
}
type QtUiAction struct {
	Name  string   `xml:"name,attr"`
}
type QtUiLayout struct {
	Items  [] QtUiLayoutItem   `xml:"item"`
}
type QtUiLayoutItem struct {
	Widget  QtUiWidget   `xml:"widget"`
	Layout  QtUiLayout   `xml:"layout"`
}

type QtUiConfig struct {
	Public  bool   `json:"public"`
}
func LoadUiXml(path string, content ([] byte), i_config interface{}) (common.UnitFile, error) {
	var config, _ = i_config.(QtUiConfig)
	var ui QtUi
	var err = xml.Unmarshal(content, &ui)
	if err != nil { return nil, err }
	qt.MakeSureInitialized()
	var widgets = make(map[string] QtWidget)
	var result = make(chan error)
	qt.CommitTask(func() {
		var root_instance, ok = qt.LoadWidget(string(content), filepath.Dir(path))
		if !ok {
			result <- errors.New("error parsing UI file")
			return
		}
		widgets[ui.RootWidget.Name] = QtWidget {
			Class:    ui.RootWidget.Class,
			Instance: root_instance,
		}
		// TODO: collect ui.RootWidget.Actions
		var add_children func(def QtUiWidget, instance qt.Widget)
		add_children = func(def QtUiWidget, instance qt.Widget) {
			var all_children = make([] QtUiWidget, 0)
			for _, child := range def.Children {
				all_children = append(all_children, child)
			}
			var consume_layout func(QtUiLayout)
			consume_layout = func(layout QtUiLayout) {
				for _, item := range layout.Items {
					all_children = append(all_children, item.Widget)
					consume_layout(item.Layout)
				}
			}
			consume_layout(def.Layout)
			for _, child := range all_children {
				var name = child.Name
				var child_instance, ok = qt.FindChild(instance, name)
				if !ok { panic("something went wrong") }
				widgets[name] = QtWidget {
					Class:    child.Class,
					Instance: child_instance,
				}
				add_children(child, child_instance)
			}
		}
		add_children(ui.RootWidget, root_instance)
		result <- nil
	})
	err = <-result
	if err != nil { return QtUiFile{}, err }
	return QtUiFile {
		Path:    path,
		Widgets: widgets,
		Config:  config,
	}, nil
}

func UiXmlLoader() common.UnitFileLoader {
	return common.UnitFileLoader {
		Extensions: [] string { "ui" },
		Name:       "ui_xml",
		Load:       LoadUiXml,
		IsResource: false,
	}
}

