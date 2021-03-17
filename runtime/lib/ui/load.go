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

type BindOptions struct {
	Debug   bool
	Sched   rx.Scheduler
	Assets  resources
}

var activeBindings = make(map[qt.Widget] struct{})

func Bind(view qt.Widget, root rx.Action, opts BindOptions) func() {
	var _, exists = activeBindings[view]
	if exists {
		panic("Cannot bind a WebView to multiple root components. " +
			"The previous binding is required to be cancelled before " +
			"the establishment of a new binding.")
	}
	var debug = opts.Debug
	var sched = opts.Sched
	var assets = opts.Assets
	cancel1 := qt.Connect(view, "eventEmitted()", func() {
		handleEvent(view, sched)
	})
	cancel2 := qt.Connect(view, "loadFinished()", func() {
		scheduleUpdate(view, sched, root, debug)
	})
	qt.CommitTask(func() {
		registerAssetFiles(view, assets)
		qt.WebViewLoadContent(view)
	})
	activeBindings[view] = struct{}{}
	return func() {
		cancel1()
		cancel2()
		delete(activeBindings, view)
	}
}

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
		qt.CommitTask(func() {
			var title_runes = RuneSliceFromString(title)
			var title, del_title = qt.NewString(title_runes)
			defer del_title()
			var icon, del_icon = qt.NewIconEmpty()
			defer del_icon()
			dialog, _ = qt.WebDialogCreate(nil, icon, title, 40, 30, true)
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

