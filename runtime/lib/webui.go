package lib

import (
	"kumachan/qt"
	"kumachan/runtime/rx"
)


var WebUiFunctions = map[string] interface{} {
	"webui-debug": func() rx.Effect {
		return rx.CreateEffect(func(_ rx.Sender) {
			qt.MakeSureInitialized()
			qt.CommitTask(func() {
				var title, del_title = qt.NewStringFromRunes([]rune("WebUI Debug"))
				qt.WebUiDebug(title)
				del_title()
			})
		})
	},
}
