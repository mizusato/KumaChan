package lib

import (
	"kumachan/qt"
	"kumachan/runtime/rx"
	. "kumachan/runtime/common"
	"kumachan/runtime/lib/container"
	"kumachan/qt/qtbinding/webui/vdom"
	"kumachan/util"
)


var __WebUiLoading = make(chan struct{}, 1)
var __WebUiLoaded = make(chan struct{})

func WebUiInitAndLoad(sched rx.Scheduler, root rx.Effect, title String) {
	select {
	case __WebUiLoading <- struct{}{}:
		qt.MakeSureInitialized()
		var wait = make(chan struct{})
		qt.CommitTask(func() {
			var title_runes = RuneSliceFromString(title)
			var title, del_title = qt.NewStringFromRunes(title_runes)
			defer del_title()
			qt.WebUiInit(title)
			wait <- struct{}{}
		})
		<- wait
		var window = qt.WebUiGetWindow()
		qt.Connect(window, "eventEmitted()", func() {
			var handler = qt.WebUiGetEventHandler()
			var payload = qt.WebUiGetEventPayload()
			var sink = handler.(rx.Sink)
			sched.RunTopLevel(sink.Send(payload), rx.Receiver {
				Context:   rx.Background(),
			})
		})
		qt.Connect(window, "loadFinished()", func() {
			var update = root.ConcatMap(func(node rx.Object) rx.Effect {
				return __WebUiUpdateDom(node.(*vdom.Node))
			})
			sched.RunTopLevel(update, rx.Receiver {
				Context: rx.Background(),
			})
		})
		qt.CommitTask(func() {
			qt.WebUiLoadView()
			wait <- struct{}{}
		})
		<- wait
		close(__WebUiLoaded)
	default:
		<-__WebUiLoaded
	}
}

type WebUiAdaptedMap struct {
	Data  container.Map
}
func WebUiAdaptMap(m container.Map) WebUiAdaptedMap {
	return WebUiAdaptedMap { m }
}
func WebUiMapAdaptValue(v Value) Value {
	var str, is_str = v.(String)
	if is_str {
		return RuneSliceFromString(str)
	} else {
		return v
	}
}
func (m WebUiAdaptedMap) Has(key vdom.String) bool {
	var key_str = StringFromRuneSlice(key)
	var _, ok = m.Data.Lookup(key_str)
	return ok
}
func (m WebUiAdaptedMap) Lookup(key vdom.String) (interface{}, bool) {
	var key_str = StringFromRuneSlice(key)
	var v, ok = m.Data.Lookup(key_str)
	return WebUiMapAdaptValue(v), ok
}
func (m WebUiAdaptedMap) ForEach(f func(key vdom.String, val interface{})) {
	m.Data.ForEach(func(k Value, v Value) {
		f(RuneSliceFromString(k.(String)), WebUiMapAdaptValue(v))
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


var WebUiFunctions = map[string] interface{} {
	"webui-init": func(title String, root rx.Effect, h MachineHandle) rx.Effect {
		return rx.CreateEffect(func(_ rx.Sender) {
			WebUiInitAndLoad(h.GetScheduler(), root, title)
		})
	},
	"webui-dom-node": func(tag String, styles *vdom.Styles, events *vdom.Events, content vdom.Content) *vdom.Node {
		var tag_ = RuneSliceFromString(tag)
		return &vdom.Node {
			Tag:     tag_,
			Props:   vdom.Props {
				Styles: styles,
				Events: events,
			},
			Content: content,
		}
	},
	"webui-dom-styles": func(styles container.Map) *vdom.Styles {
		return &vdom.Styles { Data: WebUiAdaptMap(styles) }
	},
	"webui-dom-event": func(prevent SumValue, stop SumValue, sink rx.Sink) *vdom.EventOptions {
		return &vdom.EventOptions {
			Prevent: BoolFrom(prevent),
			Stop:    BoolFrom(stop),
			Handler: (vdom.EventHandler)(sink),
		}
	},
	"webui-dom-events": func(events container.Map) *vdom.Events {
		return &vdom.Events { Data: WebUiAdaptMap(events) }
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
	"webui-event-payload-get-string": func(ev *qt.WebUiEventPayload, key String) String {
		return StringFromRuneSlice(qt.WebUiEventPayloadGetRunes(ev, RuneSliceFromString(key)))
	},
	"webui-event-payload-get-float": func(ev *qt.WebUiEventPayload, key String) float64 {
		return util.CheckFloat(qt.WebUiEventPayloadGetNumber(ev, RuneSliceFromString(key)))
	},
}
