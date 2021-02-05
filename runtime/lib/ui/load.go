package ui

import (
	"kumachan/rx"
	. "kumachan/lang"
	"kumachan/runtime/lib/ui/qt"
)


var loading = make(chan struct{}, 1)
var windowLoaded = make(chan struct{})
var bridgeLoaded = make(chan struct{})

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
		qt.CommitTask(func() {
			var title_runes = RuneSliceFromString(title)
			var title, del_title = qt.NewString(title_runes)
			defer del_title()
			qt.WebUiInit(title, debug)
			wait <- struct{}{}
		})
		<- wait
		var window = qt.WebUiGetWindow()
		qt.Connect(window, "eventEmitted()", func() {
			handleEvent(sched)
		})
		qt.Connect(window, "loadFinished()", func() {
			close(bridgeLoaded)
			scheduleUpdate(sched, root, debug)
		})
		qt.CommitTask(func() {
			registerAssetFiles(assets)
			qt.WebUiLoadView()
			wait <- struct{}{}
		})
		<- wait
		close(windowLoaded)
		return true
	default:
		<- windowLoaded
		return false
	}
}

