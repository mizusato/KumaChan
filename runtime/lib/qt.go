package lib

import (
	"fmt"
	"kumachan/qt"
	"kumachan/runtime/rx"
	. "kumachan/runtime/common"
	"kumachan/runtime/lib/container"
	"kumachan/stdlib"
)


type QtSignal struct {
	Object      qt.Object
	Signature   string
	PropMapper  func(qt.Object) interface{}
}
func (signal QtSignal) Receive() rx.Effect {
	return rx.NewGoroutine(func(sender rx.Sender) {
		var disconnect = qt.Connect(signal.Object, signal.Signature, func() {
			sender.Next(signal.PropMapper(signal.Object))
		})
		sender.Context().WaitDispose(func() {
			disconnect()
		})
	})
}

type QtEvent struct {
	Object   qt.Object
	Kind     qt.EventKind
	Prevent  bool
}
func (event QtEvent) Receive() rx.Effect {
	return rx.NewGoroutine(func(sender rx.Sender) {
		var cancel = qt.Listen(event.Object, event.Kind, event.Prevent, func(ev qt.Event) {
			var obj = (func() Value {
				switch event.Kind {
				case qt.EventResize():
					// Qt::EventResize
					return &ValProd { Elements: [] Value {
						ev.ResizeEventGetWidth(),
						ev.ResizeEventGetHeight(),
					} }
				case qt.EventClose():
					return nil
				default:
					panic("something went wrong")
				}
			})()
			sender.Next(obj)
		})
		sender.Context().WaitDispose(func() {
			cancel()
		})
	})
}

func QtFileDialogAdaptArgs (
	parent  SumValue,
	title   String,
	cwd     Path,
	filter  String,
) (qt.Widget, qt.FileDialogOptions) {
	var parent_val, ok = Unwrap(parent)
	var parent_widget qt.Widget
	if ok {
		parent_widget = parent_val.(qt.Widget)
	}
	var opts = qt.FileDialogOptions {
		Title:  RuneSliceFromString(title),
		Cwd:    ([] rune)(cwd.String()),
		Filter: RuneSliceFromString(filter),
	}
	return parent_widget, opts
}

func CreateQtTaskEffect(action func() interface{}) rx.Effect {
	return rx.NewCallback(func(callback func(rx.Object)) {
		qt.CommitTask(func() {
			callback(action())
		})
	})
}

var QtFunctions = map[string] interface{} {
	"qt-show": func(widget qt.Widget) rx.Effect {
		return CreateQtTaskEffect(func() interface{} {
			qt.Show(widget)
			return nil
		})
	},
	"qt-move-to-screen-center": func(widget qt.Widget) rx.Effect {
		return CreateQtTaskEffect(func() interface{} {
			qt.MoveToScreenCenter(widget)
			return nil
		})
	},
	"qt-signal": func(object qt.Object, signature String, mapper Value, h InteropContext) rx.Effect {
		var source = QtSignal {
			Object:     object,
			Signature:  GoStringFromString(signature),
			PropMapper: func(object qt.Object) interface{} {
				return h.Call(mapper, object)
			},
		}
		return source.Receive()
	},
	"qt-signal-no-payload": func(object qt.Object, signature String) rx.Effect {
		var source = QtSignal {
			Object:     object,
			Signature:  GoStringFromString(signature),
			PropMapper: func(object qt.Object) interface{} {
				return nil
			},
		}
		return source.Receive()
	},
	"qt-event": func(object qt.Object, kind String, prevent SumValue) rx.Effect {
		var event_kind = (func() qt.EventKind {
			var k = GoStringFromString(kind)
			switch k {
			case "resized":
				return qt.EventResize()
			case "closed":
				return qt.EventClose()
			default:
				panic(fmt.Sprintf("unsupported Qt event kind %s", k))
			}
		})()
		var prevent_default = FromBool(prevent)
		var source = QtEvent {
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
		default:
			panic(fmt.Sprintf("unsupported Qt property type %s", t))
		}
	},
	"qt-set-property": func(object qt.Object, prop_name String, prop_type String, value Value) rx.Effect {
		var prop = GoStringFromString(prop_name)
		var t = GoStringFromString(prop_type)
		return CreateQtTaskEffect(func() interface{} {
			switch t {
			case "String":
				qt.SetPropRuneString(object, prop, RuneSliceFromString(value.(String)))
			case "Bool":
				qt.SetPropBool(object, prop, FromBool(value.(SumValue)))
			default:
				panic(fmt.Sprintf("unsupported Qt property type %s", t))
			}
			return nil
		})
	},
	"qt-list-widget-set-items": func(list qt.Widget, av Value, current SumValue) rx.Effect {
		return CreateQtTaskEffect(func() interface{} {
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
	"qt-dialog-open": func(parent SumValue, title String, cwd Path, filter String) rx.Effect {
		var parent_widget, opts = QtFileDialogAdaptArgs(parent, title, cwd, filter)
		return rx.NewCallback(func(ok func(rx.Object)) {
			qt.CommitTask(func() {
				var path_runes = qt.FileDialogOpen(parent_widget, opts)
				var path_str = string(path_runes)
				if path_str != "" {
					var path = ParsePath(string(path_str))
					ok(Just(path))
				} else {
					ok(Na())
				}
			})
		})
	},
	"qt-dialog-open*": func(parent SumValue, title String, cwd Path, filter String) rx.Effect {
		var parent_widget, opts = QtFileDialogAdaptArgs(parent, title, cwd, filter)
		return rx.NewCallback(func(ok func(rx.Object)) {
			qt.CommitTask(func() {
				var str_list = qt.FileDialogOpenMultiple(parent_widget, opts)
				var path_list = make([] Path, len(str_list))
				for i, str := range str_list {
					path_list[i] = ParsePath(string(str))
				}
				ok(path_list)
			})
		})
	},
	"qt-dialog-open-dir": func(parent SumValue, title String, cwd Path) rx.Effect {
		var parent_widget, opts = QtFileDialogAdaptArgs(parent, title, cwd, String([] Char {}))
		return rx.NewCallback(func(ok func(rx.Object)) {
			qt.CommitTask(func() {
				var path_runes = qt.FileDialogSelectDirectory(parent_widget, opts)
				var path_str = string(path_runes)
				if path_str != "" {
					var path = ParsePath(string(path_str))
					ok(Just(path))
				} else {
					ok(Na())
				}
			})
		})
	},
	"qt-dialog-save": func(parent SumValue, title String, cwd Path, filter String) rx.Effect {
		var parent_widget, opts = QtFileDialogAdaptArgs(parent, title, cwd, filter)
		return rx.NewCallback(func(ok func(rx.Object)) {
			qt.CommitTask(func() {
				var path_runes = qt.FileDialogSave(parent_widget, opts)
				var path_str = string(path_runes)
				if path_str != "" {
					var path = ParsePath(string(path_str))
					ok(Just(path))
				} else {
					ok(Na())
				}
			})
		})
	},
}
