package ui

import (
	"kumachan/standalone/rx"
	"kumachan/interpreter/runtime/lib/ui/qt"
)


type BindOptions struct {
	Debug       bool
	Sched       rx.Scheduler
	AssetIndex  AssetIndex
	AssetsUsed  [] Asset
}

var activeBindings = make(map[qt.Widget] struct{})

func Bind(view qt.Widget, root rx.Observable, opts BindOptions) func() {
	var _, exists = activeBindings[view]
	if exists {
		panic("Cannot bind a WebView to multiple root components. " +
			"The previous binding is required to be cancelled before " +
			"the establishment of a new binding.")
	}
	var debug = opts.Debug
	var sched = opts.Sched
	var asset_index = opts.AssetIndex
	var assets_used = opts.AssetsUsed
	cancel1 := qt.Connect(view, "eventEmitted()", func() {
		handleEvent(view, sched)
	})
	cancel2 := qt.Connect(view, "loadFinished()", func() {
		injectAssetFiles(view, assets_used)
		scheduleUpdate(view, sched, root, debug)
	})
	var wait = make(chan struct{})
	qt.CommitTask(func() {
		registerAssetFiles(view, asset_index, assets_used)
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

