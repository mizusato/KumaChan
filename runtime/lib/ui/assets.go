package ui

import (
	"reflect"
	"kumachan/util"
	"kumachan/stdlib"
	"kumachan/runtime/lib/ui/qt"
)


type resources = func() (map[string] util.Resource)

type Font struct {
	File  stdlib.WebAsset
	Info  FontInfo
}
type FontInfo struct {
	Family  qt.Ucs4String
	Weight  qt.Ucs4String
	Style   qt.Ucs4String
}

func registerAssetFiles(assets resources) {
	for path, item := range assets() {
		var path_q, path_del = qt.NewString(([] rune)(path))
		var mime_q, mime_del = qt.NewString(([] rune)(item.MIME))
		qt.WebUiRegisterAsset(path_q, mime_q, item.Data)
		mime_del()
		path_del()
	}
}
func injectAssetFiles(files interface{}, inject func(interface{})(qt.String,[](func()))) {
	<-bridgeLoaded
	var wait = make(chan struct{})
	qt.CommitTask(func() {
		var rv = reflect.ValueOf(files)
		for i := 0; i < rv.Len(); i += 1 {
			var f = rv.Index(i).Interface()
			var uuid, deferred = inject(f)
			qt.DeleteString(uuid) // unused now
			for _, f := range deferred {
				f()
			}
		}
		wait <- struct{}{}
	})
	<- wait
}

