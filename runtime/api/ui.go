package api

import (
	"kumachan/rx"
	"kumachan/util"
	"kumachan/stdlib"
	. "kumachan/lang"
	"kumachan/runtime/lib/container"
	"kumachan/runtime/lib/ui"
	"kumachan/runtime/lib/ui/qt"
	"kumachan/runtime/lib/ui/vdom"
)


var UiFunctions = map[string] interface{} {
	"ui-init": func(title String, root rx.Action, h InteropContext) rx.Action {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			ui.Init(h, root, title)
			return nil, true
		})
	},
	"ui-get-window": func() rx.Action {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			return ui.GetWindow(), true
		})
	},
	"ui-inject-css": func(v Value) rx.Action {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			var array = container.ArrayFrom(v)
			var files = make([] stdlib.WebAsset, array.Length)
			for i := uint(0); i < array.Length; i += 1 {
				files[i] = array.GetItem(i).(stdlib.WebAsset)
			}
			ui.InjectCSS(files)
			return nil, true
		})
	},
	"ui-inject-js": func(v Value) rx.Action {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			var array = container.ArrayFrom(v)
			var files = make([] stdlib.WebAsset, array.Length)
			for i := uint(0); i < array.Length; i += 1 {
				files[i] = array.GetItem(i).(stdlib.WebAsset)
			}
			ui.InjectJS(files)
			return nil, true
		})
	},
	"ui-inject-ttf": func(v Value) rx.Action {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			var array = container.ArrayFrom(v)
			var fonts = make([] ui.Font, array.Length)
			for i := uint(0); i < array.Length; i += 1 {
				var item = array.GetItem(i).(ProductValue)
				var info = item.Elements[0].(ProductValue)
				var family = RuneSliceFromString(info.Elements[0].(String))
				var weight = RuneSliceFromString(info.Elements[1].(String))
				var style = RuneSliceFromString(info.Elements[2].(String))
				var file = item.Elements[1].(stdlib.WebAsset)
				fonts[i] = ui.Font {
					File: file,
					Info: ui.FontInfo {
						Family: family,
						Weight: weight,
						Style:  style,
					},
				}
			}
			ui.InjectTTF(fonts)
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
		return &vdom.Styles { Data: ui.VdomAdaptMap(styles) }
	},
	"ui-dom-styles-zero": func(_ Value) *vdom.Styles {
		return vdom.EmptyStyles
	},
	"ui-dom-styles-merge": func(v Value) *vdom.Styles {
		var list = container.ArrayFrom(v)
		return ui.VdomMergeStyles(list)
	},
	"ui-with-styles": func(node *vdom.Node, styles *vdom.Styles) *vdom.Node {
		return &vdom.Node {
			Tag:     node.Tag,
			Props:   vdom.Props {
				Attrs:  node.Attrs,
				Styles: ui.VdomMergeStyles(container.ArrayFrom([] *vdom.Styles {
					node.Styles, styles,
				})),
				Events: node.Events,
			},
			Content: node.Content,
		}
	},
	"ui-dom-attrs": func(attrs container.Map) *vdom.Attrs {
		if attrs.IsEmpty() { return vdom.EmptyAttrs }
		return &vdom.Attrs { Data: ui.VdomAdaptMap(attrs) }
	},
	"ui-dom-attrs-zero": func(_ Value) *vdom.Attrs {
		return vdom.EmptyAttrs
	},
	"ui-dom-attrs-merge": func(v Value) *vdom.Attrs {
		var list = container.ArrayFrom(v)
		return ui.VdomMergeAttrs(list)
	},
	"ui-with-attrs": func(node *vdom.Node, attrs *vdom.Attrs) *vdom.Node {
		return &vdom.Node {
			Tag:     node.Tag,
			Props:   vdom.Props {
				Attrs:  ui.VdomMergeAttrs(container.ArrayFrom([] *vdom.Attrs {
					node.Attrs, attrs,
				})),
				Styles: node.Styles,
				Events: node.Events,
			},
			Content: node.Content,
		}
	},
	"ui-dom-event": func(prevent SumValue, stop SumValue, capture SumValue, handler *vdom.EventHandler) *vdom.EventOptions {
		return &vdom.EventOptions {
			Prevent: FromBool(prevent),
			Stop:    FromBool(stop),
			Capture: FromBool(capture),
			Handler: handler,
		}
	},
	"ui-dom-event-handler": func(s rx.Sink, f Value, h InteropContext) *vdom.EventHandler {
		var adapter = func(obj rx.Object) rx.Object {
			var ev = obj.(*qt.WebUiEventPayload)
			return qt.WebUiConsumeEventPayload(ev, func(ev *qt.WebUiEventPayload) interface{} {
				return h.Call(f, ev)
			})
		}
		var sink = rx.SinkAdapt(s, adapter)
		return &vdom.EventHandler { Handler: sink }
	},
	"ui-dom-event-handler-reactive": func(r rx.Reactive, f Value, h InteropContext) *vdom.EventHandler {
		var in = func(state rx.Object) func(rx.Object) rx.Object {
			return func(obj rx.Object) rx.Object {
				var ev = obj.(*qt.WebUiEventPayload)
				return qt.WebUiConsumeEventPayload(ev, func(ev *qt.WebUiEventPayload) interface{} {
					return h.Call(h.Call(f, state), ev)
				})
			}
		}
		var sink = rx.ReactiveAdapt(r, in)
		return &vdom.EventHandler { Handler: sink }
	},
	"ui-dom-events": func(events container.Map) *vdom.Events {
		if events.IsEmpty() { return vdom.EmptyEvents }
		return &vdom.Events { Data: ui.VdomAdaptMap(events) }
	},
	"ui-dom-events-zero": func(_ Value) *vdom.Events {
		return vdom.EmptyEvents
	},
	"ui-dom-events-merge": func(v Value) *vdom.Events {
		var list = container.ArrayFrom(v)
		return ui.VdomMergeEvents(list)
	},
	"ui-with-events": func(node *vdom.Node, events *vdom.Events) *vdom.Node {
		return &vdom.Node {
			Tag:     node.Tag,
			Props:   vdom.Props {
				Styles: node.Styles,
				Attrs:  node.Attrs,
				Events: ui.VdomMergeEvents(container.ArrayFrom([] *vdom.Events {
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
	"ui-dom-children": func(children_ Value) vdom.Content {
		var arr = container.ArrayFrom(children_)
		if arr.Length == 0 { return vdom.EmptyContent }
		var children = make([] *vdom.Node, arr.Length)
		for i := uint(0); i < arr.Length; i += 1 {
			children[i] = arr.GetItem(i).(*vdom.Node)
		}
		var boxed = vdom.Children(children)
		return &boxed
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

