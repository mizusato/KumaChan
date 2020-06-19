package lib

import (
	"fmt"
	"kumachan/qt"
	"kumachan/runtime/rx"
	. "kumachan/runtime/common"
)


type QtSignal struct {
	Object      qt.Object
	Signature   string
	PropMapper  func(qt.Object) interface{}
}
func (signal QtSignal) Listen() rx.Effect {
	return rx.CreateEffect(func(sender rx.Sender) {
		var channel = make(chan func())
		qt.CommitTask(func() {
			var disconnect = qt.Connect(signal.Object, signal.Signature, func() {
				sender.Next(signal.PropMapper(signal.Object))
			})
			channel <- disconnect
		})
		var disconnect = <- channel
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
func (event QtEvent) Listen() rx.Effect {
	return rx.CreateEffect(func(sender rx.Sender) {
		var channel = make(chan func())
		qt.CommitTask(func() {
			var cancel = qt.Listen(event.Object, event.Kind, event.Prevent, func(ev qt.Event) {
				var obj = (func() Value {
					switch event.Kind {
					case qt.EventResize():
						// Qt::EventResize
						return &ValProd { Elements: [] Value {
							ev.ResizeEventGetWidth(),
							ev.ResizeEventGetHeight(),
						} }
					default:
						panic("something went wrong")
					}
				})()
				sender.Next(obj)
			})
			channel <- cancel
		})
		var cancel = <- channel
		sender.Context().WaitDispose(func() {
			cancel()
		})
	})
}

func CreateQtTaskEffect(action func() interface{}) rx.Effect {
	return rx.CreateValueCallbackEffect(func(callback func(rx.Object)) {
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
	"qt-signal": func(object qt.Object, signature String, mapper Value, h MachineHandle) QtSignal {
		return QtSignal {
			Object:     object,
			Signature:  GoStringFromString(signature),
			PropMapper: func(object qt.Object) interface{} {
				return h.Call(mapper, object)
			},
		}
	},
	"qt-signal-no-payload": func(object qt.Object, signature String) QtSignal {
		return QtSignal {
			Object:     object,
			Signature:  GoStringFromString(signature),
			PropMapper: func(object qt.Object) interface{} {
				return nil
			},
		}
	},
	"qt-event": func(object qt.Object, kind String, prevent SumValue) QtEvent {
		var event_kind = (func() qt.EventKind {
			var k = GoStringFromString(kind)
			switch k {
			case "Resize":
				return qt.EventResize()
			default:
				panic(fmt.Sprintf("unsupported Qt event kind %s", k))
			}
		})()
		var prevent_default = BoolFrom(prevent)
		return QtEvent {
			Object:  object,
			Kind:    event_kind,
			Prevent: prevent_default,
		}
	},
	"qt-get-property": func(object qt.Object, prop_name String, prop_type String) Value {
		var prop = GoStringFromString(prop_name)
		var t = GoStringFromString(prop_type)
		switch t {
		case "String":
			return String(StringFromRuneSlice(qt.GetPropRuneString(object, prop)))
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
			default:
				panic(fmt.Sprintf("unsupported Qt property type %s", t))
			}
			return nil
		})
	},
}