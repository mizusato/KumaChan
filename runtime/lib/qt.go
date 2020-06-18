package lib

import (
	"kumachan/qt"
	"kumachan/runtime/rx"
)


var QtFunctions = map[string] interface{} {
	"qt-show": func(widget qt.Widget) rx.Effect {
		return rx.CreateEffect(func(sender rx.Sender) {
			qt.CommitTask(func() {
				qt.Show(widget)
				// TODO: wrap single-valued async effect
				sender.Next(nil)
				sender.Complete()
			})
		})
	},
}
