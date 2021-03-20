package api

import (
	"fmt"
	. "kumachan/lang"
	"kumachan/rx"
	"kumachan/stdlib"
	"kumachan/runtime/lib/container"
	"kumachan/runtime/lib/ui"
	"kumachan/runtime/lib/ui/qt"
)


var UiQtFunctions = map[string] interface{} {
	"qt-show": func(widget qt.Widget) rx.Action {
		return ui.CreateQtTaskAction(func() interface{} {
			qt.Show(widget)
			return nil
		})
	},
	"qt-show-at-center": func(widget qt.Widget) rx.Action {
		return ui.CreateQtTaskAction(func() interface{} {
			qt.Show(widget)
			qt.MoveToScreenCenter(widget)
			return nil
		})
	},
	"qt-signal": func(object qt.Object, signature String, mapper Value, h InteropContext) rx.Action {
		var source = ui.QtSignal {
			Object:     object,
			Signature:  GoStringFromString(signature),
			PropMapper: func(object qt.Object) interface{} {
				return h.Call(mapper, object)
			},
		}
		return source.Receive()
	},
	"qt-signal-no-payload": func(object qt.Object, signature String) rx.Action {
		var source = ui.QtSignal {
			Object:     object,
			Signature:  GoStringFromString(signature),
			PropMapper: func(object qt.Object) interface{} {
				return nil
			},
		}
		return source.Receive()
	},
	"qt-event": func(object qt.Object, kind String, prevent SumValue) rx.Action {
		var event_kind = (func() qt.EventKind {
			var k = GoStringFromString(kind)
			switch k {
			case "resize":
				return qt.EventResize()
			case "close":
				return qt.EventClose()
			default:
				panic(fmt.Sprintf("unsupported Qt event kind %s", k))
			}
		})()
		var prevent_default = FromBool(prevent)
		var source = ui.QtEvent {
			Object:  object,
			Kind:    event_kind,
			Prevent: prevent_default,
		}
		return source.Receive()
	},
	"qt-get-property": func(object qt.Object, prop_name String, prop_type String) Value {
		var prop = GoStringFromString(prop_name)
		var t = GoStringFromString(prop_type)
		switch t {
		case "String":
			return String(StringFromRuneSlice(qt.GetPropRuneString(object, prop)))
		case "Bool":
			return ToBool(qt.GetPropBool(object, prop))
		case "MaybeNumber":
			var v = qt.GetPropInt(object, prop)
			if v < 0 {
				return Na()
			} else {
				return Just(uint(v))
			}
		default:
			panic(fmt.Sprintf("unsupported Qt property type %s", t))
		}
	},
	"qt-set-property": func(object qt.Object, prop_name String, prop_type String, value Value) rx.Action {
		var prop = GoStringFromString(prop_name)
		var t = GoStringFromString(prop_type)
		return ui.CreateQtTaskAction(func() interface{} {
			var _, unblock = qt.BlockCallbacks(object)
			defer unblock()
			// workaround: block data reflow of 2-way bindings
			//             (read and compare before writing)
			switch t {
			case "String":
				var current_value = qt.GetPropRuneString(object, prop)
				var new_value = RuneSliceFromString(value.(String))
				if len(current_value) == len(new_value) {
					var L = len(current_value)
					var all_equal = true
					for i := 0; i < L; i += 1 {
						if current_value[i] != new_value[i] {
							all_equal = false
							break
						}
					}
					if all_equal {
						break
					}
				}
				qt.SetPropRuneString(object, prop, new_value)
			case "Bool":
				var current_value = qt.GetPropBool(object, prop)
				var new_value = FromBool(value.(SumValue))
				if new_value == current_value {
					break
				}
				qt.SetPropBool(object, prop, new_value)
			case "MaybeNumber":
				var current_value = qt.GetPropInt(object, prop)
				var new_value = (func() int {
					var v, ok = Unwrap(value.(SumValue))
					if ok {
						return int(v.(uint))
					} else {
						return -1
					}
				})()
				if new_value == current_value {
					break
				}
				qt.SetPropInt(object, prop, new_value)
			default:
				panic(fmt.Sprintf("unsupported Qt property type %s", t))
			}
			return nil
		})
	},
	"qt-list-widget-set-items": func(list qt.Widget, av Value, current SumValue) rx.Action {
		return ui.CreateQtTaskAction(func() interface{} {
			var arr = container.ArrayFrom(av)
			var current_key ([] rune)
			var the_current, has_current = Unwrap(current)
			if has_current {
				current_key = RuneSliceFromString(the_current.(String))
			} else {
				current_key = nil
			}
			qt.ListWidgetSetItems(list, func(i uint) qt.ListWidgetItem {
				return arr.GetItem(i).(qt.ListWidgetItem)
			}, arr.Length, current_key)
			return nil
		})
	},
	"qt-list-widget-item": func(key String, label String) qt.ListWidgetItem {
		return qt.ListWidgetItem {
			Key:   RuneSliceFromString(key),
			Label: RuneSliceFromString(label),
			Icon:  nil,
		}
	},
	"qt-list-widget-item-with-icon-png": func(key String, png *stdlib.PNG, label String) qt.ListWidgetItem {
		return qt.ListWidgetItem {
			Key:   RuneSliceFromString(key),
			Label: RuneSliceFromString(label),
			Icon:  &qt.ImageData {
				Data:   png.Data,
				Format: qt.PNG,
			},
		}
	},
	"qt-list-widget-get-current": func(list qt.Widget) Value {
		if qt.ListWidgetHasCurrentItem(list) {
			var key_runes = qt.ListWidgetGetCurrentItemKey(list)
			return Just(StringFromRuneSlice(key_runes))
		} else {
			return Na()
		}
	},
	"qt-dialog-open": func(parent SumValue, title String, cwd stdlib.Path, filter String) rx.Action {
		var parent_widget, opts = ui.QtFileDialogAdaptArgs(parent, title, cwd, filter)
		return rx.NewCallback(func(ok func(rx.Object)) {
			qt.CommitTask(func() {
				var path_runes = qt.FileDialogOpen(parent_widget, opts)
				var path_str = string(path_runes)
				if path_str != "" {
					var path = stdlib.ParsePath(string(path_str))
					ok(Just(path))
				} else {
					ok(Na())
				}
			})
		})
	},
	"qt-dialog-open-multiple": func(parent SumValue, title String, cwd stdlib.Path, filter String) rx.Action {
		var parent_widget, opts = ui.QtFileDialogAdaptArgs(parent, title, cwd, filter)
		return rx.NewCallback(func(ok func(rx.Object)) {
			qt.CommitTask(func() {
				var str_list = qt.FileDialogOpenMultiple(parent_widget, opts)
				var path_list = make([] stdlib.Path, len(str_list))
				for i, str := range str_list {
					path_list[i] = stdlib.ParsePath(string(str))
				}
				ok(path_list)
			})
		})
	},
	"qt-dialog-open-directory": func(parent SumValue, title String, cwd stdlib.Path) rx.Action {
		var parent_widget, opts = ui.QtFileDialogAdaptArgs(parent, title, cwd, String([] Char {}))
		return rx.NewCallback(func(ok func(rx.Object)) {
			qt.CommitTask(func() {
				var path_runes = qt.FileDialogSelectDirectory(parent_widget, opts)
				var path_str = string(path_runes)
				if path_str != "" {
					var path = stdlib.ParsePath(string(path_str))
					ok(Just(path))
				} else {
					ok(Na())
				}
			})
		})
	},
	"qt-dialog-save": func(parent SumValue, title String, cwd stdlib.Path, filter String) rx.Action {
		var parent_widget, opts = ui.QtFileDialogAdaptArgs(parent, title, cwd, filter)
		return rx.NewCallback(func(ok func(rx.Object)) {
			qt.CommitTask(func() {
				var path_runes = qt.FileDialogSave(parent_widget, opts)
				var path_str = string(path_runes)
				if path_str != "" {
					var path = stdlib.ParsePath(string(path_str))
					ok(Just(path))
				} else {
					ok(Na())
				}
			})
		})
	},
}

