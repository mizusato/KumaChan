package extra

import (
	"strings"
	"encoding/xml"
	"path/filepath"
	"kumachan/lang"
	"kumachan/lang/parser"
	"kumachan/lang/parser/ast"
	"kumachan/compiler/loader/common"
	"kumachan/stdlib"
)


var __UiXmlWidgetDefaultNames = map[string] ([] string) {
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

type UiXmlFile struct {
	Path     string
	Content  string
	Root     string
	Widgets  map[string] UiXmlWidgetInfo
	Actions  map[string] struct {}
	Config   UiXmlConfig
}
type UiXmlWidgetInfo struct {
	Class     string
}
func (f UiXmlFile) GetAST() (ast.Root, *parser.Error) {
	var widget_list = make([] string, 0)
	for name, _ := range f.Widgets {
		if name != f.Root {
			widget_list = append(widget_list, name)
		}
	}
	var action_list = make([] string, 0)
	for name, _ := range f.Actions {
		action_list = append(action_list, name)
	}
	var group = &lang.UiObjectGroup {
		GroupName: f.Path,
		BaseDir:   filepath.Dir(f.Path),
		XmlDef:    f.Content,
		RootName:  f.Root,
		Widgets:   widget_list,
		Actions:   action_list,
	}
	var ast_root = common.CreateEmptyAST(f.Path)
	var ast_root_node = ast_root.Node
	outer: for name, widget := range f.Widgets {
		var normalized_class = strings.TrimPrefix(widget.Class, "Q")
		var default_names = __UiXmlWidgetDefaultNames[normalized_class]
		for _, default_name := range default_names {
			if name == default_name ||
				strings.HasPrefix(name, (default_name + "_")) {
				continue outer
			}
		}
		var type_name, exists = stdlib.GetQtWidgetTypeName(normalized_class)
		if !(exists) {
			continue
		}
		var thunk = lang.UiObjectThunk {
			Object: name,
			Group:  group,
		}
		var const_decl = common.CreateConstant (
			ast_root_node,
			f.Config.Public,
			name,
			stdlib.Mod_ui,
			type_name,
			thunk,
		)
		ast_root.Statements = append(ast_root.Statements, const_decl)
	}
	for name, _ := range f.Actions {
		var thunk = lang.UiObjectThunk {
			Object: name,
			Group:  group,
		}
		var const_decl = common.CreateConstant (
			ast_root_node,
			f.Config.Public,
			name,
			stdlib.Mod_ui,
			stdlib.QtActionType,
			thunk,
		)
		ast_root.Statements = append(ast_root.Statements, const_decl)
	}
	return ast_root, nil
}

type UiXml struct {
	RootWidget  UiXmlWidget   `xml:"widget"`
}
type UiXmlWidget struct {
	Name      string           `xml:"name,attr"`
	Class     string           `xml:"class,attr"`
	Children  [] UiXmlWidget   `xml:"widget"`
	Layout    UiXmlLayout      `xml:"layout"`
	Actions   [] UiXmlAction   `xml:"action"`
}
type UiXmlAction struct {
	Name  string   `xml:"name,attr"`
}
type UiXmlLayout struct {
	Items  [] UiXmlLayoutItem   `xml:"item"`
}
type UiXmlLayoutItem struct {
	Widget  UiXmlWidget   `xml:"widget"`
	Layout  UiXmlLayout   `xml:"layout"`
}

type UiXmlConfig struct {
	Public  bool   `json:"public"`
}
func LoadUiXml(path string, content ([] byte), config_ interface{}) (common.UnitFile, error) {
	var config, _ = config_.(UiXmlConfig)
	var ui UiXml
	var err = xml.Unmarshal(content, &ui)
	if err != nil { return nil, err }
	var widgets = make(map[string] UiXmlWidgetInfo)
	var root_name = ui.RootWidget.Name
	widgets[root_name] = UiXmlWidgetInfo {
		Class: ui.RootWidget.Class,
	}
	var actions = make(map[string] struct{})
	for _, item := range ui.RootWidget.Actions {
		actions[item.Name] = struct{}{}
	}
	// TODO: collect ui.RootWidget.Actions
	var add_children func(def UiXmlWidget)
	add_children = func(def UiXmlWidget) {
		var all_children = make([] UiXmlWidget, 0)
		for _, child := range def.Children {
			all_children = append(all_children, child)
		}
		var consume_layout func(UiXmlLayout)
		consume_layout = func(layout UiXmlLayout) {
			for _, item := range layout.Items {
				all_children = append(all_children, item.Widget)
				consume_layout(item.Layout)
			}
		}
		consume_layout(def.Layout)
		for _, child := range all_children {
			var name = child.Name
			widgets[name] = UiXmlWidgetInfo {
				Class: child.Class,
			}
			add_children(child)
		}
	}
	add_children(ui.RootWidget)
	var content_string = string(content)
	return UiXmlFile {
		Path:    path,
		Content: content_string,
		Root:    root_name,
		Widgets: widgets,
		Actions: actions,
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

