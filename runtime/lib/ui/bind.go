package ui

import (
	"kumachan/rx"
	"kumachan/runtime/lib/ui/qt"
)


type BindOptions struct {
	Debug   bool
	Sched   rx.Scheduler
	Assets  AssetIndex
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
	var wait = make(chan struct{})
	qt.CommitTask(func() {
		registerAssetFiles(view, assets)
		qt.WebViewLoadContent(view)
		wait <- struct{}{}
	})
	<- wait
	activeBindings[view] = struct{}{}
	return func() {
		cancel1()
		cancel2()
		delete(activeBindings, view)
	}
}

