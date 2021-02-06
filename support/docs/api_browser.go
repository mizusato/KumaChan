package docs

import (
	"os"
	"fmt"
	"sort"
	"reflect"
	"kumachan/util"
	"kumachan/runtime/lib/ui/qt"
)


const apiBrowserFontSize = 24

var apiBrowserUiXml =
	string(util.ReadInterpreterResource("api_browser.ui"))

var apiBrowserDocStyle =
	fmt.Sprintf("<style>body { font-size: %dpx; }</style><style>%s</style>",
		apiBrowserFontSize,
		string(util.ReadInterpreterResource("api_doc.css")))

var apiIcons = (func() (map[string] *qt.ImageData) {
	var result = make(map[string] *qt.ImageData)
	var names = [] string { "module", "type" }
	for _, name := range names {
		var file = fmt.Sprintf("icons/%s.png", name)
		result[name] = &qt.ImageData {
			Data:   util.ReadInterpreterResource(file),
			Format: qt.PNG,
		}
	}
	return result
})()

type ApiBrowser struct {
	Window  qt.Widget
	ApiBrowserWidgets
}

type ApiBrowserWidgets struct {
	ModuleList   qt.Widget
	ContentView  qt.Widget
	OutlineView  qt.Widget
}

func RunApiBrowser(doc ApiDocIndex) {
	qt.MakeSureInitialized()
	qt.CommitTask(func() {
		var ui_xml_dir = util.InterpreterResourceFolderPath()
		var window, ok = qt.LoadWidget(apiBrowserUiXml, ui_xml_dir)
		if !(ok) { panic("something went wrong") }
		var widgets = ApiBrowserWidgets {}
		var widgets_rv = reflect.ValueOf(&widgets).Elem()
		var widgets_t = widgets_rv.Type()
		for i := 0; i < widgets_t.NumField(); i += 1 {
			var widget_name = widgets_t.Field(i).Name
			var widget, ok = qt.FindChild(window, widget_name)
			if !(ok) { panic("something went wrong") }
			widgets_rv.Field(i).Set(reflect.ValueOf(widget))
		}
		var ui = ApiBrowser {
			Window: window,
			ApiBrowserWidgets: widgets,
		}
		apiBrowserUiLogic(ui, doc)
		go (func() {
			qt.Listen(window, qt.EventClose(), true, func(_ qt.Event) {
				os.Exit(0)
			})
		})()
		qt.MoveToScreenCenter(window)
		qt.Show(window)
	})
}

func apiBrowserUiLogic(ui ApiBrowser, doc ApiDocIndex) {
	var modules = make([] string, 0)
	for mod, _ := range doc {
		modules = append(modules, mod)
	}
	sort.Strings(modules)
	var get_mod_item = func(i uint) qt.ListWidgetItem {
		var mod = modules[i]
		var mod_ucs4 = ([] rune)(mod)
		return qt.ListWidgetItem {
			Key:   mod_ucs4,
			Label: mod_ucs4,
			Icon:  apiIcons["module"],
		}
	}
	var mod_count = uint(len(modules))
	qt.ListWidgetSetItems(ui.ModuleList, get_mod_item, mod_count, nil)
	qt.WebViewDisableContextMenu(ui.ContentView)
	qt.WebViewEnableLinkDelegation(ui.ContentView)
	go (func() {
		qt.Connect(ui.ModuleList, "currentRowChanged(int)", func() {
			var key = qt.ListWidgetGetCurrentItemKey(ui.ModuleList)
			var mod = string(key)
			var mod_data = doc[mod]
			var content = string(mod_data.Content)
			var styled_content = (apiBrowserDocStyle + content)
			var styled_content_, del = qt.NewStringFromGoString(styled_content)
			defer del()
			qt.WebViewSetHTML(ui.ContentView, styled_content_)
			var outline = mod_data.Outline
			var get_outline_item = func(i uint) qt.ListWidgetItem {
				var api = outline[i]
				return qt.ListWidgetItem {
					Key:   ([] rune)(api.Id),
					Label: ([] rune)(api.Name),
					Icon:  apiKindToIcon(api.Kind),
				}
			}
			var outline_count = uint(len(outline))
			qt.ListWidgetSetItems(ui.OutlineView, get_outline_item, outline_count, nil)
		})
		qt.Connect(ui.OutlineView, "currentRowChanged(int)", func() {
			var api_id = qt.ListWidgetGetCurrentItemKey(ui.OutlineView)
			var api_id_, del = qt.NewString(api_id)
			defer del()
			qt.WebViewScrollToAnchor(ui.ContentView, api_id_)
		})
	})()
}

func apiKindToIcon(kind ApiItemKind) *qt.ImageData {
	switch kind {
	case ApiType:
		return apiIcons["type"]
	default:
		return nil
	}
}

