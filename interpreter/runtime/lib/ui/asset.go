package ui

import (
	"kumachan/interpreter/base"
	"kumachan/stdlib"
	"kumachan/interpreter/runtime/lib/ui/qt"
)


type AssetIndex = func(path string) (base.Resource, bool)

func registerAssetFiles(view qt.Widget, assets AssetIndex, selected ([] Asset)) {
	for _, selection := range selected {
		var path = selection.GetPath()
		var item, exists = assets(path)
		if !(exists) { panic("something went wrong") }
		var path_q, path_del = qt.NewString(path)
		var mime_q, mime_del = qt.NewString(item.MIME)
		qt.WebViewRegisterAsset(view, path_q, mime_q, item.Data)
		mime_del()
		path_del()
	}
}

func injectAssetFiles(view qt.Widget, selected ([] Asset)) {
	for _, selection := range selected {
		selection.InjectTo(view)
	}
}


type Asset interface {
	GetPath() string
	InjectTo(view qt.Widget)
}

type CSS struct {
	File  stdlib.AssetFile
}
func (css CSS) GetPath() string { return css.File.Path }
func (css CSS) InjectTo(view qt.Widget) {
	var path = css.File.Path
	var path_q, del = qt.NewString(path)
	defer del()
	var uuid = qt.WebViewInjectCSS(view, path_q)
	qt.DeleteString(uuid)  // not used now
}

type JS struct {
	File  stdlib.AssetFile
}
func (js JS) GetPath() string { return js.File.Path }
func (js JS) InjectTo(view qt.Widget) {
	var path = js.File.Path
	var path_q, del = qt.NewString(path)
	defer del()
	var uuid = qt.WebViewInjectJS(view, path_q)
	qt.DeleteString(uuid)  // not used now
}

type TTF struct {
	File  stdlib.AssetFile
	Font  FontName
}
type FontName struct {
	Family  string
	Weight  string
	Style   string
}
func (ttf TTF) GetPath() string { return ttf.File.Path }
func (ttf TTF) InjectTo(view qt.Widget) {
	var path_q, del1 = qt.NewString(ttf.File.Path); defer del1()
	var family_q, del2 = qt.NewString(ttf.Font.Family); defer del2()
	var weight_q, del3  = qt.NewString(ttf.Font.Weight); defer del3()
	var style_q, del4 = qt.NewString(ttf.Font.Style); defer del4()
	var uuid = qt.WebViewInjectTTF(view, path_q, family_q, weight_q, style_q)
	qt.DeleteString(uuid)  // not used now
}


