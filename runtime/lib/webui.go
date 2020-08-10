package lib

import (
	"sync"
	"kumachan/qt"
	"kumachan/runtime/rx"
	. "kumachan/runtime/common"
	"kumachan/runtime/lib/container"
	"kumachan/qt/qtbinding/webui/vdom"
)


type WebUiAdaptedMap struct {
	Data  container.Map
}
func WebUiAdaptMap(mv Value) WebUiAdaptedMap {
	return WebUiAdaptedMap{mv.(container.Map) }
}
func (m WebUiAdaptedMap) Has(key vdom.String) bool {
	var key_str = StringFromRuneSlice(key)
	var _, ok = m.Data.Lookup(key_str)
	return ok
}
func (m WebUiAdaptedMap) Lookup(key vdom.String) (interface{}, bool) {
	var key_str = StringFromRuneSlice(key)
	return m.Data.Lookup(key_str)
}
func (m WebUiAdaptedMap) ForEach(f func(key vdom.String, val interface{})) {
	m.Data.ForEach(func(k Value, v Value) {
		f(RuneSliceFromString(k.(String)), v)
	})
}

var __WebUiVirtualDomDeltaNotifier = &vdom.DeltaNotifier {
	ApplyStyle:  qt.WebUiApplyStyle,
	EraseStyle:  qt.WebUiEraseStyle,
	AttachEvent: qt.WebUiAttachEvent,
	ModifyEvent: qt.WebUiModifyEvent,
	DetachEvent: qt.WebUiDetachEvent,
	SetText:     qt.WebUiSetText,
	AppendNode:  qt.WebUiAppendNode,
	RemoveNode:  qt.WebUiRemoveNode,
	UpdateNode:  qt.WebUiUpdateNode,
	ReplaceNode: qt.WebUiReplaceNode,
}
var __WebUiVirtualDom *vdom.Node = nil
var __WebUiUpdateDom = func(new_root *vdom.Node) rx.Effect {
	return rx.CreateValueCallbackEffect(func(done func(rx.Object)) {
		qt.CommitTask(func() {
			var ctx = __WebUiVirtualDomDeltaNotifier
			var prev_root = __WebUiVirtualDom
			__WebUiVirtualDom = new_root
			vdom.Diff(ctx, nil, prev_root, new_root)
			done(nil)
		})
	})
}

var __WebUiLoaded = false
var __WebUiLoadedMutex sync.Mutex

var WebUiFunctions = map[string] interface{} {
	"webui-debug": func(root rx.Effect, h MachineHandle) rx.Effect {
		return rx.CreateEffect(func(_ rx.Sender) {
			__WebUiLoadedMutex.Lock()
			if __WebUiLoaded {
				__WebUiLoadedMutex.Unlock()
				return
			}
			qt.MakeSureInitialized()
			var wait = make(chan struct{})
			qt.CommitTask(func() {
				var title, del_title = qt.NewStringFromRunes([]rune("WebUI Debug"))
				defer del_title()
				qt.WebUiInit(title)
				wait <- struct{}{}
			})
			<- wait
			__WebUiLoaded = true
			__WebUiLoadedMutex.Unlock()
			var window = qt.WebUiGetWindow()
			qt.Connect(window, "loadFinished()", func() {
				var update = root.ConcatMap(func(node rx.Object) rx.Effect {
					return __WebUiUpdateDom(node.(*vdom.Node))
				})
				h.GetScheduler().RunTopLevel(update, rx.Receiver {
					Context: rx.Background(),
				})
			})
		})
	},
	"webui-dom-node": func(tag String, styles Value, events Value, content vdom.Content) *vdom.Node {
		var tag_ = RuneSliceFromString(tag)
		var styles_ = WebUiAdaptMap(styles)
		var events_ = WebUiAdaptMap(events)
		return &vdom.Node {
			Tag:     tag_,
			Props:   vdom.Props {
				Styles: styles_,
				Events: events_,
			},
			Content: content,
		}
	},
	"webui-dom-event": func(prevent SumValue, stop SumValue, sink rx.Sink) *vdom.EventOptions {
		return &vdom.EventOptions {
			Prevent: BoolFrom(prevent),
			Stop:    BoolFrom(stop),
			Handler: (vdom.EventHandler)(sink),
		}
	},
	"webui-dom-text": func(text String) vdom.Content {
		var t = vdom.Text(RuneSliceFromString(text))
		return &t
	},
	"webui-dom-children": func(children Value) vdom.Content {
		var arr = container.ArrayFrom(children)
		var children_ = make([] *vdom.Node, arr.Length)
		for i := uint(0); i < arr.Length; i += 1 {
			children_[i] = arr.GetItem(i).(*vdom.Node)
		}
		var c = vdom.Children(children_)
		return &c
	},
}
