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
var __WebUiEmptyAttrs = &vdom.Attrs { Data: WebUiEmptyMap{} }
var __WebUiEmptyStyles = &vdom.Styles { Data: WebUiEmptyMap{} }
var __WebUiEmptyEvents = &vdom.Events { Data: WebUiEmptyMap{} }
var __WebUiEmptyContent = &vdom.Children {}

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
				Context: rx.Background(),
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

type WebUiEmptyMap struct {}
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
func (_ WebUiEmptyMap) Has(key vdom.String) bool { return false }
func (m WebUiAdaptedMap) Has(key vdom.String) bool {
	var key_str = StringFromRuneSlice(key)
	var _, ok = m.Data.Lookup(key_str)
	return ok
}
func (_ WebUiEmptyMap) Lookup(key vdom.String) (interface{}, bool) { return nil, false }
func (m WebUiAdaptedMap) Lookup(key vdom.String) (interface{}, bool) {
	var key_str = StringFromRuneSlice(key)
	var v, ok = m.Data.Lookup(key_str)
	return WebUiMapAdaptValue(v), ok
}
func (_ WebUiEmptyMap) ForEach(_ func(key vdom.String, val interface{})) {}
func (m WebUiAdaptedMap) ForEach(f func(key vdom.String, val interface{})) {
	m.Data.ForEach(func(k Value, v Value) {
		f(RuneSliceFromString(k.(String)), WebUiMapAdaptValue(v))
	})
}

var __WebUiVirtualDomDeltaNotifier = &vdom.DeltaNotifier {
	ApplyStyle:  qt.WebUiApplyStyle,
	EraseStyle:  qt.WebUiEraseStyle,
	SetAttr:     qt.WebUiSetAttr,
	RemoveAttr:  qt.WebUiRemoveAttr,
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
	return rx.NewCallback(func(done func(rx.Object)) {
		qt.CommitTask(func() {
			var ctx = __WebUiVirtualDomDeltaNotifier
			var prev_root = __WebUiVirtualDom
			__WebUiVirtualDom = new_root
			vdom.Diff(ctx, nil, prev_root, new_root)
			done(nil)
		})
	})
}


var WebUiConstants = map[string] NativeConstant {
	"WebUi::GetWindow": func(_ InteropContext) Value {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			<- __WebUiLoaded
			return qt.WebUiGetWindow(), true
		})
	},
}

var WebUiFunctions = map[string] interface{} {
	"webui-init": func(title String, root rx.Effect, h InteropContext) rx.Effect {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			WebUiInitAndLoad(h.GetScheduler(), root, title)
			return nil, true
		})
	},
	"webui-dom-node": func (
		tag      String,
		styles   *vdom.Styles,
		attrs    *vdom.Attrs,
		events   *vdom.Events,
		content  vdom.Content,
	) *vdom.Node {
		var tag_ = RuneSliceFromString(tag)
		return &vdom.Node {
			Tag:     tag_,
			Props:   vdom.Props {
				Styles: styles,
				Attrs:  attrs,
				Events: events,
			},
			Content: content,
		}
	},
	"webui-dom-styles": func(styles container.Map) *vdom.Styles {
		return &vdom.Styles { Data: WebUiAdaptMap(styles) }
	},
	"webui-dom-styles-zero": func(_ Value) *vdom.Styles {
		return __WebUiEmptyStyles
	},
	"webui-dom-attrs": func(attrs container.Map) *vdom.Attrs {
		return &vdom.Attrs { Data: WebUiAdaptMap(attrs) }
	},
	"webui-dom-attrs-zero": func(_ Value) *vdom.Attrs {
		return __WebUiEmptyAttrs
	},
	"webui-dom-event": func(prevent SumValue, stop SumValue, sink rx.Sink) *vdom.EventOptions {
		return &vdom.EventOptions {
			Prevent: BoolFrom(prevent),
			Stop:    BoolFrom(stop),
			Handler: (vdom.EventHandler)(sink),
		}
	},
	"webui-dom-event-sink": func(sink rx.Sink, f Value, h InteropContext) rx.Sink {
		return &rx.AdaptedSink {
			Sink:    sink,
			Adapter: func(obj rx.Object) rx.Object {
				var ev = obj.(*qt.WebUiEventPayload)
				return qt.WebUiConsumeEventPayload(ev, func(ev *qt.WebUiEventPayload) interface{} {
					return h.Call(f, ev)
				})
			},
		}
	},
	"webui-dom-event-sink-latch": func(latch *rx.Latch, f Value, h InteropContext) rx.Sink {
		return &rx.AdaptedLatch {
			Latch:      latch,
			GetAdapter: func(state rx.Object) func(rx.Object) rx.Object {
				return func(obj rx.Object) rx.Object {
					var ev = obj.(*qt.WebUiEventPayload)
					return qt.WebUiConsumeEventPayload(ev, func(ev *qt.WebUiEventPayload) interface{} {
						return h.Call(h.Call(f, state), ev)
					})
				}
			},
		}
	},
	"webui-dom-events": func(events container.Map) *vdom.Events {
		return &vdom.Events { Data: WebUiAdaptMap(events) }
	},
	"webui-dom-events-zero": func(_ Value) *vdom.Events {
		return __WebUiEmptyEvents
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
	"webui-dom-content-zero": func(_ Value) vdom.Content {
		return __WebUiEmptyContent
	},
	"webui-event-payload-get-string": func(ev *qt.WebUiEventPayload, key String) String {
		return StringFromRuneSlice(qt.WebUiEventPayloadGetRunes(ev, RuneSliceFromString(key)))
	},
	"webui-event-payload-get-float": func(ev *qt.WebUiEventPayload, key String) float64 {
		return util.CheckFloat(qt.WebUiEventPayloadGetNumber(ev, RuneSliceFromString(key)))
	},
}
