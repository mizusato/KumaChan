package ui

import (
	"kumachan/rx"
	"kumachan/stdlib"
	. "kumachan/lang"
	"kumachan/runtime/lib/ui/qt"
)


func Init(h InteropContext, root rx.Action, title String) {
	var assets = func() (map[string] Resource) {
		return h.GetResources("web_asset")
	}
	var debug_opts = h.GetDebugOptions()
	var debug = debug_opts.DebugUI
	var ok = load(debug, h.GetScheduler(), root, title, assets)
	if !(ok) {
		panic("UI: duplicate initialization operation")
	}
}

func GetWindow() qt.Widget {
	<-windowLoaded
	return qt.WebUiGetWindow()
}

func InjectJS(files ([] stdlib.WebAsset)) {
	injectAssetFiles(files, func(file interface{}) (qt.String, [](func())) {
		var path = file.(stdlib.WebAsset).Path
		var path_q, del = qt.NewString(([] rune)(path))
		var uuid = qt.WebUiInjectJS(path_q)
		return uuid, [] func() { del }
	})
}
func InjectCSS(files ([] stdlib.WebAsset)) {
	injectAssetFiles(files, func(file interface{}) (qt.String, []func()) {
		var path = file.(stdlib.WebAsset).Path
		var path_q, del = qt.NewString(([] rune)(path))
		var uuid = qt.WebUiInjectCSS(path_q)
		return uuid, [] func() { del }
	})
}
func InjectTTF(fonts ([] Font)) {
	injectAssetFiles(fonts, func(font_ interface{}) (qt.String, []func()) {
		var font = font_.(Font)
		var path_q, del1 = qt.NewString(([] rune)(font.File.Path))
		var family_q, del2 = qt.NewString(font.Info.Family)
		var weight_q, del3  = qt.NewString(font.Info.Weight)
		var style_q, del4 = qt.NewString(font.Info.Style)
		var uuid = qt.WebUiInjectTTF(path_q, family_q, weight_q, style_q)
		return uuid, [] func() { del4, del3, del2, del1 }
	})
	<-bridgeLoaded
	var wait = make(chan struct{})
	qt.CommitTask(func() {
		for _, font := range fonts {
			var path_q, del1 = qt.NewString(([] rune)(font.File.Path))
			var family_q, del2 = qt.NewString(font.Info.Family)
			var weight_q, del3  = qt.NewString(font.Info.Weight)
			var style_q, del4 = qt.NewString(font.Info.Style)
			var uuid = qt.WebUiInjectTTF(path_q, family_q, weight_q, style_q)
			qt.DeleteString(uuid) // unused now
			del1(); del2(); del3(); del4()
		}
		wait <- struct{}{}
	})
	<- wait
}

