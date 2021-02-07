package docs

import (
	"os"
	"fmt"
	"sort"
	"reflect"
	"strings"
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

type ApiBrowser struct {
	Window  qt.Widget
	ApiBrowserWidgets
	ApiBrowserActions
}
type ApiBrowserWidgets struct {
	ModuleList   qt.Widget
	ContentView  qt.Widget
	OutlineView  qt.Widget
}
type ApiBrowserActions struct {
	ActionBack     qt.Action
	ActionForward  qt.Action
}

type ApiRef struct {
	Module  string
	Id      string
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
			var name = widgets_t.Field(i).Name
			var widget, exists = qt.FindChild(window, name)
			if !(exists) { panic("something went wrong") }
			widgets_rv.Field(i).Set(reflect.ValueOf(widget))
		}
		var actions = ApiBrowserActions {}
		var actions_rv = reflect.ValueOf(&actions).Elem()
		var actions_t = actions_rv.Type()
		for i := 0; i < actions_t.NumField(); i += 1 {
			var name = actions_t.Field(i).Name
			var action, exists = qt.FindChildAction(window, name)
			if !(exists) { panic("something went wrong") }
			actions_rv.Field(i).Set(reflect.ValueOf(action))
		}
		var ui = ApiBrowser {
			Window: window,
			ApiBrowserWidgets: widgets,
			ApiBrowserActions: actions,
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
	var module_index = make(map[string] int)
	for i, mod := range modules {
		module_index[mod] = i
	}
	var get_mod_item = func(i uint) qt.ListWidgetItem {
		var mod = modules[i]
		var mod_ucs4 = ([] rune)(mod)
		return qt.ListWidgetItem {
			Key:   mod_ucs4,
			Label: mod_ucs4,
			Icon:  icons["module"],
		}
	}
	var mod_count = uint(len(modules))
	qt.ListWidgetSetItems(ui.ModuleList, get_mod_item, mod_count, nil)
	qt.WebViewDisableContextMenu(ui.ContentView)
	qt.WebViewEnableLinkDelegation(ui.ContentView)
	qt.WebViewRecordClickedLink(ui.ContentView)
	var current_ref = ApiRef { Module: "", Id: "" }
	var current_outline_index = make(map[string] int)
	var undo_stack = make([] ApiRef, 0)
	var redo_stack = make([] ApiRef, 0)
	var jump_state struct {
		jumping    bool
		has_step1  bool
		save       bool
	}
	var jump_init = func(save bool) {
		jump_state.jumping = true
		jump_state.has_step1 = false
		jump_state.save = save
	}
	var jump_clear = func() {
		jump_state.jumping = false
		jump_state.has_step1 = false
		jump_state.save = true
	}
	jump_clear()
	var is_first_update = true
	var update_current = func(ref ApiRef, is_step1 bool, is_step2 bool) {
		if is_first_update { defer (func() { is_first_update = false })() }
		var jumping = jump_state.jumping
		var save = jump_state.save
		if jumping && is_step1 {
			jump_state.has_step1 = true
		}
		var has_step1 = jump_state.has_step1
		if save &&
			!(jumping && has_step1 && !(is_step1)) &&
			!(is_first_update) {
			undo_stack = append(undo_stack, current_ref)
		}
		current_ref = ref
		if save {
			redo_stack = redo_stack[0:0]
		}
		fmt.Printf("undo: %+v\n", undo_stack)
		fmt.Printf("current: %+v\n", current_ref)
		fmt.Printf("redo: %+v\n\n", redo_stack)
	}
	var jump = func(ref ApiRef, save bool) {
		jump_init(save)
		defer jump_clear()
		var mod = ref.Module
		var id = ref.Id
		if mod != current_ref.Module {
			qt.SetPropInt(ui.ModuleList, "currentRow", module_index[mod])
		}
		if id != "" {
			qt.SetPropInt(ui.OutlineView, "currentRow", current_outline_index[id])
		}
	}
	var undo = func() {
		if len(undo_stack) == 0 { return }
		var ref = undo_stack[len(undo_stack)-1]
		undo_stack = undo_stack[:len(undo_stack)-1]
		redo_stack = append(redo_stack, current_ref)
		jump(ref, false)
	}
	var redo = func() {
		if len(redo_stack) == 0 { return }
		var ref = redo_stack[len(redo_stack)-1]
		redo_stack = redo_stack[:len(redo_stack)-1]
		undo_stack = append(undo_stack, current_ref)
		jump(ref, false)
	}
	go (func() {
		qt.Connect(ui.ModuleList, "currentRowChanged(int)", func() {
			if !(qt.ListWidgetHasCurrentItem(ui.ModuleList)) {
				return
			}
			var key = qt.ListWidgetGetCurrentItemKey(ui.ModuleList)
			var mod = string(key)
			var mod_data = doc[mod]
			var content = string(mod_data.Content)
			var styled_content = (apiBrowserDocStyle + content)
			var styled_content_, del = qt.NewStringFromGoString(styled_content)
			defer del()
			qt.WebViewSetHTML(ui.ContentView, styled_content_)
			var outline = mod_data.Outline
			current_outline_index = make(map[string] int)
			for i, item := range outline {
				current_outline_index[item.Id] = i
			}
			var get_outline_item = func(i uint) qt.ListWidgetItem {
				var api = outline[i]
				return qt.ListWidgetItem {
					Key:   ([]rune)(api.Id),
					Label: ([]rune)(api.Name),
					Icon:  apiKindToIcon(api.Kind),
				}
			}
			var outline_count = uint(len(outline))
			qt.ListWidgetSetItems(ui.OutlineView, get_outline_item, outline_count, nil)
			update_current(ApiRef {
				Module: mod,
			}, true, false)
		})
		qt.Connect(ui.OutlineView, "currentRowChanged(int)", func() {
			if !(qt.ListWidgetHasCurrentItem(ui.OutlineView)) {
				return
			}
			var api_id = qt.ListWidgetGetCurrentItemKey(ui.OutlineView)
			var api_id_, del = qt.NewString(api_id)
			defer del()
			qt.WebViewScrollToAnchor(ui.ContentView, api_id_)
			update_current(ApiRef {
				Module: current_ref.Module,
				Id:     string(api_id),
			}, false, true)
		})
		qt.Connect(ui.ContentView, "linkClicked(const QUrl&)", func() {
			var url = qt.GetPropString(ui.ContentView, "qtbindingClickedLinkUrl")
			var t = strings.Split(url, "#")
			if len(t) != 2 { return }
			var mod = t[0]
			var id = t[1]
			jump(ApiRef { Module: mod, Id: id }, true)
		})
		qt.Connect(ui.ActionBack, "triggered()", func() {
			undo()
		})
		qt.Connect(ui.ActionForward, "triggered()", func() {
			redo()
		})
	})()
}

