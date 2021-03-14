package ui

import (
	"kumachan/rx"
	. "kumachan/lang"
	"kumachan/runtime/lib/ui/qt"
)


var loading = make(chan struct{}, 1)
var dialogLoaded = make(chan struct{})
var bridgeLoaded = make(chan struct{})
var singletonDialog qt.Widget
var singletonView qt.Widget

func load (
	debug   bool,
	sched   rx.Scheduler,
	root    rx.Action,
	title   String,
	assets  resources,
) bool {
	select {
	case loading <- struct{}{}:
		qt.MakeSureInitialized()
		var wait = make(chan struct{})
		var dialog qt.Widget
		var del_dialog func()
		qt.CommitTask(func() {
			var title_runes = RuneSliceFromString(title)
			var title, del_title = qt.NewString(title_runes)
			defer del_title()
			var icon, del_icon = qt.NewIconEmpty()
			defer del_icon()
			dialog, del_dialog = qt.WebDialogCreate(nil, icon, title, 40, 30, true)
			qt.DialogShowModal(dialog)
			wait <- struct{}{}
		})
		<- wait
		var view = qt.WebDialogGetWebView(dialog)
		qt.Connect(view, "eventEmitted()", func() {
			handleEvent(view, sched)
		})
		qt.Connect(view, "loadFinished()", func() {
			select { case <- bridgeLoaded: panic("unexpected page refresh"); default: }
			close(bridgeLoaded)
			scheduleUpdate(view, sched, root, debug)
		})
		qt.CommitTask(func() {
			registerAssetFiles(view, assets)
			qt.WebViewLoadContent(view)
			wait <- struct{}{}
		})
		<- wait
		singletonDialog = dialog
		singletonView = view
		close(dialogLoaded)
		return true
	default:
		<-dialogLoaded
		return false
	}
}

