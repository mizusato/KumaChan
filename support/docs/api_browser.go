package docs

import (
	"os"
	"fmt"
	"sort"
	"reflect"
	"strings"
	"context"
	"runtime"
	"kumachan/misc/util"
	"kumachan/runtime/lib/ui/qt"
)


const apiBrowserFontSize = 24

var apiBrowserUiXml =
	string(util.ReadInterpreterResource("api_browser.ui"))

var apiSearchDialogUiXml =
	string(util.ReadInterpreterResource("api_search.ui"))

var apiBrowserDocStyle =
	fmt.Sprintf("<style>body { font-size: %dpx; }</style><style>%s</style>",
		apiBrowserFontSize,
		string(util.ReadInterpreterResource("api_doc.css")))

type ApiBrowser struct {
	Window  qt.Widget
	ApiBrowserWidgets
	ApiBrowserActions
	ApiBrowserDialogs
}
type ApiBrowserWidgets struct {
	ModuleList   qt.Widget
	ContentView  qt.Widget
	OutlineView  qt.Widget
}
type ApiBrowserActions struct {
	ActionBack     qt.Action
	ActionForward  qt.Action
	ActionSearch   qt.Action
}
type ApiBrowserDialogs struct {
	SearchDialog  ApiSearchDialog
}
type ApiSearchDialog struct {
	Dialog  qt.Widget
	ApiSearchDialogWidgets
}
type ApiSearchDialogWidgets struct {
	KindSelect    qt.Widget
	ContentInput  qt.Widget
	ResultList    qt.Widget
}

type ApiDocPosition struct {
	ApiRef
	Scroll  qt.Point
}
type ApiRef struct {
	Module  string
	Id      string
}
func ApiRefFromHyperRef(href string) (ApiRef, bool) {
	var t = strings.Split(href, "#")
	if len(t) != 2 {
		return ApiRef{}, false
	}
	var mod = t[0]
	var id = t[1]
	return ApiRef { Module: mod, Id: id }, true
}
func (ref ApiRef) HyperRef() string {
	return fmt.Sprintf("%s#%s", ref.Module, ref.Id)
}

func RunApiBrowser(doc ApiDocIndex) {
	qt.MakeSureInitialized()
	qt.CommitTask(func() {
		var ui_xml_dir = util.InterpreterResourceFolderPath()
		window, ok := qt.LoadWidget(apiBrowserUiXml, ui_xml_dir)
		if !(ok) { panic("something went wrong") }
		dialog, ok := qt.LoadWidget(apiSearchDialogUiXml, ui_xml_dir)
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
		var dialog_widgets = ApiSearchDialogWidgets {}
		var dialog_widgets_rv = reflect.ValueOf(&dialog_widgets).Elem()
		var dialog_widgets_t = dialog_widgets_rv.Type()
		for i := 0; i < dialog_widgets_t.NumField(); i += 1 {
			var name = dialog_widgets_t.Field(i).Name
			var widget, exists = qt.FindChild(dialog, name)
			if !(exists) { panic("something went wrong") }
			dialog_widgets_rv.Field(i).Set(reflect.ValueOf(widget))
		}
		var dialogs = ApiBrowserDialogs {
			SearchDialog: ApiSearchDialog {
				Dialog: dialog,
				ApiSearchDialogWidgets: dialog_widgets,
			},
		}
		var ui = ApiBrowser {
			Window: window,
			ApiBrowserWidgets: widgets,
			ApiBrowserActions: actions,
			ApiBrowserDialogs: dialogs,
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
	qt.BaseWebViewDisableContextMenu(ui.ContentView)
	qt.BaseWebViewEnableLinkDelegation(ui.ContentView)
	qt.BaseWebViewRecordClickedLink(ui.ContentView)
	var current_ref = ApiRef { Module: "", Id: "" }
	var current_outline_index = make(map[string] int)
	var undo_stack = make([] ApiDocPosition, 0)
	var redo_stack = make([] ApiDocPosition, 0)
	var get_current_pos = func() ApiDocPosition {
		return ApiDocPosition {
			ApiRef: current_ref,
			Scroll: qt.BaseWebViewGetScroll(ui.ContentView),
		}
	}
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
			undo_stack = append(undo_stack, get_current_pos())
		}
		current_ref = ref
		if save {
			redo_stack = redo_stack[0:0]
		}
		// fmt.Printf("undo: %+v\n", undo_stack)
		// fmt.Printf("current: %+v\n", current_ref)
		// fmt.Printf("redo: %+v\n\n", redo_stack)
	}
	type JumpOptions struct {
		SaveToUndo     bool
		SpecifyScroll  bool
		Scroll         qt.Point
	}
	var jump = func(ref ApiRef, opts JumpOptions) {
		jump_init(opts.SaveToUndo)
		defer jump_clear()
		var mod = ref.Module
		var id = ref.Id
		var mod_unchanged = false
		if mod != current_ref.Module {
			qt.SetPropInt(ui.ModuleList, "currentRow", module_index[mod])
		} else {
			mod_unchanged = true
		}
		if id != "" {
			qt.SetPropInt(ui.OutlineView, "currentRow", current_outline_index[id])
		} else {
			if mod_unchanged {
				qt.SetPropInt(ui.OutlineView, "currentRow", -1)
				update_current(ApiRef { Module: mod, Id: id }, false, true)
			}
		}
		if opts.SpecifyScroll {
			qt.BaseWebViewSetScroll(ui.ContentView, opts.Scroll)
		}
	}
	var undo = func() {
		if len(undo_stack) == 0 { return }
		var pos = undo_stack[len(undo_stack)-1]
		undo_stack = undo_stack[:len(undo_stack)-1]
		redo_stack = append(redo_stack, get_current_pos())
		jump(pos.ApiRef, JumpOptions {
			SaveToUndo:    false,
			SpecifyScroll: true,
			Scroll:        pos.Scroll,
		})
	}
	var redo = func() {
		if len(redo_stack) == 0 { return }
		var pos = redo_stack[len(redo_stack)-1]
		redo_stack = redo_stack[:len(redo_stack)-1]
		undo_stack = append(undo_stack, get_current_pos())
		jump(pos.ApiRef, JumpOptions {
			SaveToUndo:    false,
			SpecifyScroll: true,
			Scroll:        pos.Scroll,
		})
	}
	go (func() {
		qt.Connect(ui.ModuleList, "currentRowChanged(int)", func() {
			if !(qt.ListWidgetHasCurrentItem(ui.ModuleList)) {
				return
			}
			var key = qt.ListWidgetGetCurrentItemKey(ui.ModuleList)
			var mod = string(key)
			update_current(ApiRef { Module: mod }, true, false)
			var mod_data = doc[mod]
			var content = string(mod_data.Content)
			var styled_content = (apiBrowserDocStyle + content)
			var styled_content_, del = qt.NewStringFromGoString(styled_content)
			defer del()
			qt.BaseWebViewSetHTML(ui.ContentView, styled_content_)
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
		})
		qt.Connect(ui.OutlineView, "currentRowChanged(int)", func() {
			if !(qt.ListWidgetHasCurrentItem(ui.OutlineView)) {
				return
			}
			var key = qt.ListWidgetGetCurrentItemKey(ui.OutlineView)
			var id = string(key)
			var mod = current_ref.Module
			update_current(ApiRef { Module: mod, Id: id }, false, true)
			var id_, del = qt.NewString(key)
			defer del()
			qt.BaseWebViewScrollToAnchor(ui.ContentView, id_)
		})
		qt.Connect(ui.ContentView, "linkClicked(const QUrl&)", func() {
			var url = qt.GetPropString(ui.ContentView, "qtbindingClickedLinkUrl")
			var ref, ok = ApiRefFromHyperRef(url)
			if !(ok) { return }
			jump(ref, JumpOptions { SaveToUndo: true })
		})
		qt.Connect(ui.ActionBack, "triggered()", func() {
			undo()
		})
		qt.Connect(ui.ActionForward, "triggered()", func() {
			redo()
		})
		qt.Connect(ui.ActionSearch, "triggered()", func() {
			qt.DialogExec(ui.SearchDialog.Dialog)
		})
		if runtime.GOOS == "windows" {
			// workaround: no focus by default on Windows
			qt.CommitTask(func() {
				qt.SetPropInt(ui.ModuleList, "currentRow", -1)
			})
		}
	})()
	var search_kind = "All"
	var search_content = ""
	var dialog = ui.SearchDialog
	var latest_search_number = uint64(0)
	var search_ctx, search_cancel = context.WithCancel(context.Background())
	var search = func() {
		var filter = search_kind
		var target = search_content
		latest_search_number += 1
		var this_search_number = latest_search_number
		search_cancel()
		search_ctx, search_cancel = context.WithCancel(context.Background())
		var ctx = search_ctx
		go (func() {
			if target == "" {
				qt.CommitTask(func() {
					if this_search_number != latest_search_number {
						return
					}
					qt.ListWidgetSetItems(dialog.ResultList, nil, 0, nil)
				})
				return
			}
			var result = make([] ApiItem, 0)
			for _, mod := range modules {
				var mod_data = doc[mod]
				for _, item := range mod_data.Outline {
					select {
					case <- ctx.Done():
						return
					default:
					}
					if filter == "Type" && item.Kind != TypeDecl { continue }
					if filter == "Constant" && item.Kind != ConstDecl { continue }
					if filter == "Function" && item.Kind != FuncDecl { continue }
					var normalized_id = strings.ToLower(item.Id)
					var normalized_target = strings.ToLower(target)
					var ok = strings.Contains(normalized_id, normalized_target)
					if ok {
						result = append(result, item)
					} else if item.Sec != nil {
						for _, s := range item.Sec {
							var normalized_s = strings.ToLower(s)
							var ok = strings.Contains(normalized_s, normalized_target)
							if ok {
								result = append(result, item)
								break
							}
						}
					}
				}
 			}
 			qt.CommitTask(func() {
 				if this_search_number != latest_search_number {
 					return
				}
 				var get_result_item = func(i uint) qt.ListWidgetItem {
 					var item = result[i]
 					var ref = ApiRef { Module: item.Mod, Id: item.Id }
 					var href = ref.HyperRef()
 					var label = (func() string {
 						var display_name = fmt.Sprintf("%s (%s)", item.Name, item.Mod)
 						if item.Sec != nil {
 							var t = make([] string, len(item.Sec))
 							for i, s := range item.Sec {
 								t[i] = fmt.Sprintf("[%s]", s)
							}
 							var sec_desc = strings.Join(t, " ")
 							return fmt.Sprintf("%s %s", display_name, sec_desc)
						} else {
							return display_name
						}
					})()
 					return qt.ListWidgetItem {
						Key:   ([] rune)(href),
						Label: ([] rune)(label),
						Icon:  apiKindToIcon(item.Kind),
					}
				}
				var result_size = uint(len(result))
				qt.ListWidgetSetItems(dialog.ResultList,
					get_result_item, result_size, nil)
			})
		})()
	}
	go (func() {
		qt.Connect(dialog.KindSelect, "currentTextChanged(const QString&)", func() {
			search_kind = qt.GetPropString(dialog.KindSelect, "currentText")
			search()
		})
		qt.Connect(dialog.ContentInput, "textEdited(const QString&)", func() {
			search_content = qt.GetPropString(dialog.ContentInput, "text")
			search()
		})
		qt.Connect(dialog.ResultList, "itemActivated(QListWidgetItem*)", func() {
			if !(qt.ListWidgetHasCurrentItem(dialog.ResultList)) {
				return
			}
			var key = qt.ListWidgetGetCurrentItemKey(dialog.ResultList)
			var href = string(key)
			var ref, ok = ApiRefFromHyperRef(href)
			if !(ok) { return }
			qt.DialogAccept(dialog.Dialog)
			jump(ref, JumpOptions { SaveToUndo: true })
		})
	})()
}

