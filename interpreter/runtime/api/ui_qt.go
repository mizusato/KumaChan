package api

import (
	"fmt"
	"reflect"
	"math/big"
	. "kumachan/interpreter/def"
	"kumachan/standalone/rx"
	"kumachan/stdlib"
	"kumachan/interpreter/runtime/lib/container"
	"kumachan/interpreter/runtime/lib/ui"
	"kumachan/standalone/qt"
)


var UiQtFunctions = map[string] interface{} {
	"qt-show": func(widget qt.Widget) rx.Observable {
		return ui.CreateQtTaskAction(func() interface{} {
			qt.Show(widget)
			return nil
		})
	},
	"qt-show-at-center": func(widget qt.Widget) rx.Observable {
		return ui.CreateQtTaskAction(func() interface{} {
			qt.Show(widget)
			qt.MoveToScreenCenter(widget)
			return nil
		})
	},
	"qt-signal": func(object qt.Object, signature string, mapper Value, h InteropContext) rx.Observable {
		var source = ui.QtSignal {
			Object:     object,
			Signature:  signature,
			PropMapper: func(object qt.Object) interface{} {
				return h.Call(mapper, object)
			},
		}
		return source.Receive()
	},
	"qt-signal-no-payload": func(object qt.Object, signature string) rx.Observable {
		var source = ui.QtSignal {
			Object:     object,
			Signature:  signature,
			PropMapper: func(object qt.Object) interface{} {
				return nil
			},
		}
		return source.Receive()
	},
	"qt-event": func(object qt.Object, kind string, prevent EnumValue) rx.Observable {
		var event_kind = (func() qt.EventKind {
			switch kind {
			case "resize":
				return qt.EventResize()
			case "close":
				return qt.EventClose()
			default:
				panic(fmt.Sprintf("unsupported Qt event kind %s", kind))
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
	"qt-get-property": func(object qt.Object, prop string, t string) Value {
		switch t {
		case "String":
			return qt.GetPropString(object, prop)
		case "Bool":
			return ToBool(qt.GetPropBool(object, prop))
		case "MaybeNumber":
			var v = qt.GetPropInt(object, prop)
			if v < 0 {
				return None()
			} else {
				return Some(big.NewInt(int64(v)))
			}
		default:
			panic(fmt.Sprintf("unsupported Qt property type %s", t))
		}
	},
	"qt-set-property": func(object qt.Object, prop string, t string, value Value) rx.Observable {
		return ui.CreateQtTaskAction(func() interface{} {
			var _, unblock = qt.BlockCallbacks(object)
			defer unblock()
			// workaround: block data reflow of 2-way bindings
			//             (read and compare before writing)
			switch t {
			case "String":
				var current_value = qt.GetPropString(object, prop)
				var new_value = value.(string)
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
				qt.SetPropString(object, prop, new_value)
			case "Bool":
				var current_value = qt.GetPropBool(object, prop)
				var new_value = FromBool(value.(EnumValue))
				if new_value == current_value {
					break
				}
				qt.SetPropBool(object, prop, new_value)
			case "MaybeNumber":
				var current_value = qt.GetPropInt(object, prop)
				var new_value = (func() int {
					var v, ok = Unwrap(value.(EnumValue))
					if ok {
						return int(v.(*big.Int).Int64())
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
	"qt-list-widget-set-items": func(list qt.Widget, av Value, current EnumValue) rx.Observable {
		return ui.CreateQtTaskAction(func() interface{} {
			var current_key ([] rune)
			var the_current, has_current = Unwrap(current)
			if has_current {
				current_key = ([] rune)(the_current.(string))
			} else {
				current_key = nil
			}
			var items = container.ListFrom(av)
			var items_slice = items.CopyAsSlice()
			var items_slice_rv = reflect.ValueOf(items_slice)
			qt.ListWidgetSetItems(list, func(i uint) qt.ListWidgetItem {
				return items_slice_rv.Index(int(i)).Interface().(qt.ListWidgetItem)
			}, uint(items_slice_rv.Len()), current_key)
			return nil
		})
	},
	"qt-list-widget-item": func(key string, label string) qt.ListWidgetItem {
		return qt.ListWidgetItem {
			Key:   key,
			Label: label,
			Icon:  nil,
		}
	},
	"qt-list-widget-item-with-icon-png": func(key string, png *stdlib.PNG, label string) qt.ListWidgetItem {
		return qt.ListWidgetItem {
			Key:   key,
			Label: label,
			Icon:  &qt.ImageData {
				Data:   png.Data,
				Format: qt.PNG,
			},
		}
	},
	"qt-list-widget-get-current": func(list qt.Widget) Value {
		if qt.ListWidgetHasCurrentItem(list) {
			var key_runes = qt.ListWidgetGetCurrentItemKey(list)
			var key = string(key_runes)
			return Some(key)
		} else {
			return None()
		}
	},
	"qt-dialog-open": func(parent EnumValue, title string, cwd stdlib.Path, filter string) rx.Observable {
		var parent_widget, opts = ui.QtFileDialogAdaptArgs(parent, title, cwd, filter)
		return rx.NewCallback(func(ok func(rx.Object)) {
			qt.CommitTask(func() {
				var path_runes = qt.FileDialogOpen(parent_widget, opts)
				var path_str = string(path_runes)
				if path_str != "" {
					var path = stdlib.ParsePath(string(path_str))
					ok(Some(path))
				} else {
					ok(None())
				}
			})
		})
	},
	"qt-dialog-open-multiple": func(parent EnumValue, title string, cwd stdlib.Path, filter string) rx.Observable {
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
	"qt-dialog-open-directory": func(parent EnumValue, title string, cwd stdlib.Path) rx.Observable {
		var parent_widget, opts = ui.QtFileDialogAdaptArgs(parent, title, cwd, "")
		return rx.NewCallback(func(ok func(rx.Object)) {
			qt.CommitTask(func() {
				var path_runes = qt.FileDialogSelectDirectory(parent_widget, opts)
				var path_str = string(path_runes)
				if path_str != "" {
					var path = stdlib.ParsePath(string(path_str))
					ok(Some(path))
				} else {
					ok(None())
				}
			})
		})
	},
	"qt-dialog-save": func(parent EnumValue, title string, cwd stdlib.Path, filter string) rx.Observable {
		var parent_widget, opts = ui.QtFileDialogAdaptArgs(parent, title, cwd, filter)
		return rx.NewCallback(func(ok func(rx.Object)) {
			qt.CommitTask(func() {
				var path_runes = qt.FileDialogSave(parent_widget, opts)
				var path_str = string(path_runes)
				if path_str != "" {
					var path = stdlib.ParsePath(string(path_str))
					ok(Some(path))
				} else {
					ok(None())
				}
			})
		})
	},
}

