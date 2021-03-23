package ui

import (
	"reflect"
	"kumachan/lang"
	"kumachan/stdlib"
	"kumachan/runtime/lib/ui/qt"
)


type AssetIndex = (map[string] lang.Resource)

type TTF struct {
	File  stdlib.AssetFile
	Font  FontName
}

type FontName struct {
	Family  qt.Ucs4String
	Weight  qt.Ucs4String
	Style   qt.Ucs4String
}

func registerAssetFiles(view qt.Widget, assets AssetIndex) {
	for path, item := range assets {
		var path_q, path_del = qt.NewString(([] rune)(path))
		var mime_q, mime_del = qt.NewString(([] rune)(item.MIME))
		qt.WebViewRegisterAsset(view, path_q, mime_q, item.Data)
		mime_del()
		path_del()
	}
}

func injectAssetFiles(view qt.Widget, files interface{}, inject func(qt.Widget,interface{})(qt.String,[](func()))) {
	var inject_all = func() {
		var rv = reflect.ValueOf(files)
		for i := 0; i < rv.Len(); i += 1 {
			var f = rv.Index(i).Interface()
			var uuid, deferred = inject(view, f)
			qt.DeleteString(uuid) // unused now
			for _, clean := range deferred {
				clean()
			}
		}
	}
	// inject after loaded
	qt.Connect(view, "loadFinished()", inject_all)
	// inject immediately if already loaded
	var wait = make(chan struct{})
	qt.CommitTask(func() {
		if qt.WebViewIsContentLoaded(view) {
			inject_all()
		}
		wait <- struct{}{}
	})
	<-wait
}

func InjectTTF(view qt.Widget, fonts ([] TTF)) {
	injectAssetFiles(view, fonts, func(view qt.Widget, ttf_ interface{}) (qt.String, []func()) {
		var ttf = ttf_.(TTF)
		var path_q, del1 = qt.NewString(([] rune)(ttf.File.Path))
		var family_q, del2 = qt.NewString(ttf.Font.Family)
		var weight_q, del3  = qt.NewString(ttf.Font.Weight)
		var style_q, del4 = qt.NewString(ttf.Font.Style)
		var uuid = qt.WebViewInjectTTF(view, path_q, family_q, weight_q, style_q)
		return uuid, [] func() { del4, del3, del2, del1 }
	})
}

func InjectJS(view qt.Widget, files ([] stdlib.AssetFile)) {
	injectAssetFiles(view, files, func(view qt.Widget, file interface{}) (qt.String, [](func())) {
		var path = file.(stdlib.AssetFile).Path
		var path_q, del = qt.NewString(([] rune)(path))
		var uuid = qt.WebViewInjectJS(view, path_q)
		return uuid, [] func() { del }
	})
}

func InjectCSS(view qt.Widget, files ([] stdlib.AssetFile)) {
	injectAssetFiles(view, files, func(view qt.Widget, file interface{}) (qt.String, []func()) {
		var path = file.(stdlib.AssetFile).Path
		var path_q, del = qt.NewString(([] rune)(path))
		var uuid = qt.WebViewInjectCSS(view, path_q)
		return uuid, [] func() { del }
	})
}

