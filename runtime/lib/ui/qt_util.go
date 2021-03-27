package ui

import (
	"kumachan/runtime/lib/ui/qt"
	"kumachan/misc/rx"
	"kumachan/stdlib"
	. "kumachan/lang"
)


type QtSignal struct {
	Object      qt.Object
	Signature   string
	PropMapper  func(qt.Object) interface{}
}
func (signal QtSignal) Receive() rx.Observable {
	return rx.NewSubscriptionWithSender(func(sender rx.Sender) func() {
		return qt.Connect(signal.Object, signal.Signature, func() {
			sender.Next(signal.PropMapper(signal.Object))
		})
	})
}

type QtEvent struct {
	Object   qt.Object
	Kind     qt.EventKind
	Prevent  bool
}
func (event QtEvent) Receive() rx.Observable {
	return rx.NewSubscriptionWithSender(func(sender rx.Sender) func() {
		return qt.Listen(event.Object, event.Kind, event.Prevent, func(ev qt.Event) {
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
	})
}

func CreateQtTaskAction(action func() interface{}) rx.Observable {
	return rx.NewCallback(func(callback func(rx.Object)) {
		qt.CommitTask(func() {
			callback(action())
		})
	})
}

func QtFileDialogAdaptArgs (
	parent  SumValue,
	title   String,
	cwd     stdlib.Path,
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

