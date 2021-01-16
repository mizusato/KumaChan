package api

import (
	"kumachan/rx"
	"kumachan/util"
	"kumachan/stdlib"
	. "kumachan/lang"
	"kumachan/runtime/lib/container"
	"kumachan/runtime/lib/ui/qt"
	"kumachan/runtime/lib/ui/vdom"
)


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
func WebUiMergeStyles(list container.Array) *vdom.Styles {
	var styles = container.NewStrMap()
	for i := uint(0); i < list.Length; i += 1 {
		var part = list.GetItem(i).(*vdom.Styles)
		part.Data.ForEach(func(k_runes vdom.String, v_obj interface{}) {
			var v_runes = v_obj.([] rune)
			var k_str = StringFromRuneSlice(k_runes)
			var v_str = StringFromRuneSlice(v_runes)
			styles, _ = styles.Inserted(k_str, v_str)
		})
	}
	return &vdom.Styles { Data: WebUiAdaptMap(styles) }
}
func WebUiMergeAttrs(list container.Array) *vdom.Attrs {
	var class = StringFromRuneSlice([] rune ("class"))
	var attrs = container.NewStrMap()
	for i := uint(0); i < list.Length; i += 1 {
		var part = list.GetItem(i).(*vdom.Attrs)
		part.Data.ForEach(func(k_runes vdom.String, v_obj interface{}) {
			var k_str = StringFromRuneSlice(k_runes)
			var v_runes = v_obj.([] rune)
			if container.StringCompare(k_str, class) == Equal {
				var existing, exists = attrs.Lookup(k_str)
				if exists {
					// merge class list
					var existing_runes = RuneSliceFromString(existing.(String))
					var buf = make([] rune, 0)
					buf = append(buf, existing_runes...)
					buf = append(buf, ' ')
					buf = append(buf, v_runes...)
					var v_merged_str = StringFromRuneSlice(buf)
					attrs, _ = attrs.Inserted(k_str, v_merged_str)
				} else {
					var v_str = StringFromRuneSlice(v_runes)
					attrs, _ = attrs.Inserted(k_str, v_str)
				}
			} else {
				var v_str = StringFromRuneSlice(v_runes)
				attrs, _ = attrs.Inserted(k_str, v_str)
			}
		})
	}
	return &vdom.Attrs { Data: WebUiAdaptMap(attrs) }
}
func WebUiMergeEvents(list container.Array) *vdom.Events {
	var events = container.NewStrMap()
	for i := uint(0); i < list.Length; i += 1 {
		var part = list.GetItem(i).(*vdom.Events)
		part.Data.ForEach(func(k_ vdom.String, v interface{}) {
			var k = StringFromRuneSlice(k_)
			events, _ = events.Inserted(k, v)
		})
	}
	return &vdom.Events { Data: WebUiAdaptMap(events) }
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

var __WebUiLoading = make(chan struct{}, 1)
var __WebUiWindowLoaded = make(chan struct{})
var __WebUiBridgeLoaded = make(chan struct{})

func WebUiInitAndLoad (
	sched  rx.Scheduler,
	root   rx.Effect,
	title  String,
	res    map[string] util.Resource,
) {
	select {
	case __WebUiLoading <- struct{}{}:
		qt.MakeSureInitialized()
		var wait = make(chan struct{})
		qt.CommitTask(func() {
			var title_runes = RuneSliceFromString(title)
			var title, del_title = qt.NewString(title_runes)
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
			sched.RunTopLevel(sink.Emit(payload), rx.Receiver {
				Context: rx.Background(),
			})
		})
		qt.Connect(window, "loadFinished()", func() {
			close(__WebUiBridgeLoaded)
			var update = root.ConcatMap(func(node rx.Object) rx.Effect {
				return __WebUiUpdateDom(node.(*vdom.Node))
			})
			sched.RunTopLevel(update, rx.Receiver {
				Context: rx.Background(),
			})
		})
		qt.CommitTask(func() {
			__WebUiRegisterAssetFiles(res)
			qt.WebUiLoadView()
			wait <- struct{}{}
		})
		<- wait
		close(__WebUiWindowLoaded)
	default:
		<-__WebUiWindowLoaded
	}
}

func __WebUiRegisterAssetFiles(res (map[string] util.Resource)) {
	for path, item := range res {
		var path_q, path_del = qt.NewString(([] rune)(path))
		var mime_q, mime_del = qt.NewString(([] rune)(item.MIME))
		qt.WebUiRegisterAsset(path_q, mime_q, item.Data)
		mime_del()
		path_del()
	}
}

func WebUiInjectAssetFiles (
	files   ([] stdlib.WebAsset),
	inject  func(qt.String)(qt.String),
) {
	<- __WebUiBridgeLoaded
	var wait = make(chan struct{})
	qt.CommitTask(func() {
		for _, f := range files {
			var path_q, del = qt.NewString(([] rune)(f.Path))
			var uuid = inject(path_q)
			qt.DeleteString(uuid) // unused now
			del()
		}
		wait <- struct{}{}
	})
	<- wait
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
	SwapNode:    qt.WebUiSwapNode,
}
var __WebUiVirtualDom *vdom.Node = nil
var __WebUiUpdateDom = func(new_root *vdom.Node) rx.Effect {
	return rx.NewCallback(func(done func(rx.Object)) {
		qt.CommitTask(func() {
			var ctx = __WebUiVirtualDomDeltaNotifier
			var prev_root = __WebUiVirtualDom
			__WebUiVirtualDom = new_root
			vdom.Diff(ctx, nil, prev_root, new_root)
			qt.WebUiPerformActualRendering()
			done(nil)
		})
	})
}


var WebUiConstants = map[string] NativeConstant {
	"UI::GetWindow": func(_ InteropContext) Value {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			<-__WebUiWindowLoaded
			return qt.WebUiGetWindow(), true
		})
	},
}

var WebUiFunctions = map[string] interface{} {
	"ui-init": func(title String, root rx.Effect, h InteropContext) rx.Effect {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			// TODO: handle duplicate load (throw an error)
			var res = h.GetResources("web_asset")
			WebUiInitAndLoad(h.GetScheduler(), root, title, res)
			return nil, true
		})
	},
	"ui-inject-css": func(v Value) rx.Effect {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			var array = container.ArrayFrom(v)
			var files = make([] stdlib.WebAsset, array.Length)
			for i := uint(0); i < array.Length; i += 1 {
				files[i] = array.GetItem(i).(stdlib.WebAsset)
			}
			WebUiInjectAssetFiles(files, qt.WebUiInjectCSS)
			return nil, true
		})
	},
	"ui-inject-js": func(v Value) rx.Effect {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			var array = container.ArrayFrom(v)
			var files = make([] stdlib.WebAsset, array.Length)
			for i := uint(0); i < array.Length; i += 1 {
				files[i] = array.GetItem(i).(stdlib.WebAsset)
			}
			WebUiInjectAssetFiles(files, qt.WebUiInjectJS)
			return nil, true
		})
	},
	"ui-inject-ttf": func(v Value) rx.Effect {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			<- __WebUiBridgeLoaded
			var array = container.ArrayFrom(v)
			var wait = make(chan struct{})
			qt.CommitTask(func() {
				for i := uint(0); i < array.Length; i += 1 {
					var item = array.GetItem(i).(ProductValue)
					var info = item.Elements[0].(ProductValue)
					var family = RuneSliceFromString(info.Elements[0].(String))
					var weight = RuneSliceFromString(info.Elements[1].(String))
					var style = RuneSliceFromString(info.Elements[2].(String))
					var f = item.Elements[1].(stdlib.WebAsset)
					var path_q, del1 = qt.NewString(([] rune)(f.Path))
					var family_q, del2 = qt.NewString(family)
					var weight_q, del3  = qt.NewString(weight)
					var style_q, del4 = qt.NewString(style)
					var uuid = qt.WebUiInjectTTF(path_q, family_q, weight_q, style_q)
					qt.DeleteString(uuid) // unused now
					del1(); del2(); del3(); del4()
				}
				wait <- struct{}{}
			})
			<- wait
			return nil, true
		})
	},
	"ui-dom-node": func (
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
	"ui-dom-styles": func(styles container.Map) *vdom.Styles {
		if styles.IsEmpty() { return vdom.EmptyStyles }
		return &vdom.Styles { Data: WebUiAdaptMap(styles) }
	},
	"ui-dom-styles-zero": func(_ Value) *vdom.Styles {
		return vdom.EmptyStyles
	},
	"ui-dom-styles-merge": func(v Value) *vdom.Styles {
		var list = container.ArrayFrom(v)
		return WebUiMergeStyles(list)
	},
	"ui-with-styles": func(node *vdom.Node, styles *vdom.Styles) *vdom.Node {
		return &vdom.Node {
			Tag:     node.Tag,
			Props:   vdom.Props {
				Attrs:  node.Attrs,
				Styles: WebUiMergeStyles(container.ArrayFrom([] *vdom.Styles {
					node.Styles, styles,
				})),
				Events: node.Events,
			},
			Content: node.Content,
		}
	},
	"ui-dom-attrs": func(attrs container.Map) *vdom.Attrs {
		if attrs.IsEmpty() { return vdom.EmptyAttrs }
		return &vdom.Attrs { Data: WebUiAdaptMap(attrs) }
	},
	"ui-dom-attrs-zero": func(_ Value) *vdom.Attrs {
		return vdom.EmptyAttrs
	},
	"ui-dom-attrs-merge": func(v Value) *vdom.Attrs {
		var list = container.ArrayFrom(v)
		return WebUiMergeAttrs(list)
	},
	"ui-with-attrs": func(node *vdom.Node, attrs *vdom.Attrs) *vdom.Node {
		return &vdom.Node {
			Tag:     node.Tag,
			Props:   vdom.Props {
				Attrs:  WebUiMergeAttrs(container.ArrayFrom([] *vdom.Attrs {
					node.Attrs, attrs,
				})),
				Styles: node.Styles,
				Events: node.Events,
			},
			Content: node.Content,
		}
	},
	"ui-dom-event": func(prevent SumValue, stop SumValue, capture SumValue, sink rx.Sink) *vdom.EventOptions {
		return &vdom.EventOptions {
			Prevent: FromBool(prevent),
			Stop:    FromBool(stop),
			Capture: FromBool(capture),
			Handler: (vdom.EventHandler)(sink),
		}
	},
	"ui-dom-event-sink": func(s rx.Sink, f Value, h InteropContext) rx.Sink {
		var adapter = func(obj rx.Object) rx.Object {
			var ev = obj.(*qt.WebUiEventPayload)
			return qt.WebUiConsumeEventPayload(ev, func(ev *qt.WebUiEventPayload) interface{} {
				return h.Call(f, ev)
			})
		}
		return rx.SinkAdapt(s, adapter)
	},
	"ui-dom-event-sink-reactive": func(r rx.Reactive, f Value, h InteropContext) rx.Sink {
		var in = func(state rx.Object) func(rx.Object) rx.Object {
			return func(obj rx.Object) rx.Object {
				var ev = obj.(*qt.WebUiEventPayload)
				return qt.WebUiConsumeEventPayload(ev, func(ev *qt.WebUiEventPayload) interface{} {
					return h.Call(h.Call(f, state), ev)
				})
			}
		}
		return rx.ReactiveAdapt(r, in)
	},
	"ui-dom-events": func(events container.Map) *vdom.Events {
		if events.IsEmpty() { return vdom.EmptyEvents }
		return &vdom.Events { Data: WebUiAdaptMap(events) }
	},
	"ui-dom-events-zero": func(_ Value) *vdom.Events {
		return vdom.EmptyEvents
	},
	"ui-dom-events-merge": func(v Value) *vdom.Events {
		var list = container.ArrayFrom(v)
		return WebUiMergeEvents(list)
	},
	"ui-with-events": func(node *vdom.Node, events *vdom.Events) *vdom.Node {
		return &vdom.Node {
			Tag:     node.Tag,
			Props:   vdom.Props {
				Styles: node.Styles,
				Attrs:  node.Attrs,
				Events: WebUiMergeEvents(container.ArrayFrom([] *vdom.Events {
					node.Events, events,
				})),
			},
			Content: node.Content,
		}
	},
	"ui-dom-text": func(text String) vdom.Content {
		var t = vdom.Text(RuneSliceFromString(text))
		return &t
	},
	"ui-dom-children": func(children Value) vdom.Content {
		var arr = container.ArrayFrom(children)
		if arr.Length == 0 { return vdom.EmptyContent }
		var children_ = make([] *vdom.Node, arr.Length)
		for i := uint(0); i < arr.Length; i += 1 {
			children_[i] = arr.GetItem(i).(*vdom.Node)
		}
		var c = vdom.Children(children_)
		return &c
	},
	"ui-dom-content-zero": func(_ Value) vdom.Content {
		return vdom.EmptyContent
	},
	"ui-event-payload-get-string": func(ev *qt.WebUiEventPayload, key String) String {
		return StringFromRuneSlice(qt.WebUiEventPayloadGetRunes(ev, RuneSliceFromString(key)))
	},
	"ui-event-payload-get-float": func(ev *qt.WebUiEventPayload, key String) float64 {
		return util.CheckFloat(qt.WebUiEventPayloadGetFloat(ev, RuneSliceFromString(key)))
	},
	"ui-event-payload-get-number": func(ev *qt.WebUiEventPayload, key String) uint {
		var x = util.CheckFloat(qt.WebUiEventPayloadGetFloat(ev, RuneSliceFromString(key)))
		return uint(x)
	},
	"ui-event-payload-get-bool": func(ev *qt.WebUiEventPayload, key String) SumValue {
		return ToBool(qt.WebUiEventPayloadGetBool(ev, RuneSliceFromString(key)))
	},
}
