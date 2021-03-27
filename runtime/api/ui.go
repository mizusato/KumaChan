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
	"ui-bind": func(view qt.Widget, root rx.Observable, h InteropContext) rx.Observable {
		var debug = h.GetDebugOptions().DebugUI
		var sched = h.Scheduler()
		var assets = h.GetResources("web_asset")
		return rx.NewSubscription(func(_ func(rx.Object)) func() {
			return ui.Bind(view, root, ui.BindOptions {
				Debug:  debug,
				Sched:  sched,
				Assets: assets,
			})
		})
	},
	"ui-inject-ttf": func(view qt.Widget, v Value) rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			var array = container.ArrayFrom(v)
			var fonts = make([] ui.TTF, array.Length)
			for i := uint(0); i < array.Length; i += 1 {
				var item = array.GetItem(i).(ProductValue)
				var info = item.Elements[0].(ProductValue)
				var family = RuneSliceFromString(info.Elements[0].(String))
				var weight = RuneSliceFromString(info.Elements[1].(String))
				var style = RuneSliceFromString(info.Elements[2].(String))
				var file = item.Elements[1].(stdlib.AssetFile)
				fonts[i] = ui.TTF {
					File: file,
					Font: ui.FontName {
						Family: family,
						Weight: weight,
						Style:  style,
					},
				}
			}
			ui.InjectTTF(view, fonts)
			return nil, true
		})
	},
	"ui-inject-css": func(view qt.Widget, v Value) rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			var array = container.ArrayFrom(v)
			var files = make([] stdlib.AssetFile, array.Length)
			for i := uint(0); i < array.Length; i += 1 {
				files[i] = array.GetItem(i).(stdlib.AssetFile)
			}
			ui.InjectCSS(view, files)
			return nil, true
		})
	},
	"ui-inject-js": func(view qt.Widget, v Value) rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			var array = container.ArrayFrom(v)
			var files = make([] stdlib.AssetFile, array.Length)
			for i := uint(0); i < array.Length; i += 1 {
				files[i] = array.GetItem(i).(stdlib.AssetFile)
			}
			ui.InjectJS(view, files)
			return nil, true
		})
	},
	"ui-static-component": func(thunk Value, h InteropContext) rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			return h.Call(thunk, nil), true
		})
	},
	"ui-dom-node": func(tag_ String) *vdom.Node {
		var tag = RuneSliceFromString(tag_)
		return &vdom.Node {
			Tag: tag,
			Props: vdom.Props {
				Styles: vdom.EmptyStyles,
				Attrs:  vdom.EmptyAttrs,
				Events: vdom.EmptyEvents,
			},
			Content: vdom.EmptyContent,
		}
	},
	"ui-dom-styles": func(styles container.Map) *vdom.Styles {
		if styles.IsEmpty() { return vdom.EmptyStyles }
		return &vdom.Styles { Data: ui.VdomAdaptMap(styles) }
	},
	"ui-dom-styles-zero": func(_ Value) *vdom.Styles {
		return vdom.EmptyStyles
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
	"ui-dom-event-handler": func(s rx.Sink, f Value, h InteropContext) *vdom.EventHandler {
		var adapter = func(obj rx.Object) rx.Object {
			var ev = obj.(*qt.WebViewEventPayload)
			return qt.WebViewConsumeEventPayload(ev, func(ev *qt.WebViewEventPayload) interface{} {
				return h.Call(f, ev)
			})
		}
		var sink = rx.SinkAdapt(s, adapter)
		return &vdom.EventHandler { Handler: sink }
	},
	"ui-dom-event-handler-reactive": func(r rx.Reactive, f Value, h InteropContext) *vdom.EventHandler {
		var in = func(state rx.Object) func(rx.Object) rx.Object {
			return func(obj rx.Object) rx.Object {
				var ev = obj.(*qt.WebViewEventPayload)
				return qt.WebViewConsumeEventPayload(ev, func(ev *qt.WebViewEventPayload) interface{} {
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
	"ui-with-content": func(node *vdom.Node, content vdom.Content) *vdom.Node {
		return &vdom.Node {
			Tag:     node.Tag,
			Props:   node.Props,
			Content: content,
		}
	},
	"ui-event-payload-get-string": func(ev *qt.WebViewEventPayload, key String) String {
		return StringFromRuneSlice(qt.WebViewEventPayloadGetRunes(ev, RuneSliceFromString(key)))
	},
	"ui-event-payload-get-float": func(ev *qt.WebViewEventPayload, key String) float64 {
		return util.CheckFloat(qt.WebViewEventPayloadGetFloat(ev, RuneSliceFromString(key)))
	},
	"ui-event-payload-get-number": func(ev *qt.WebViewEventPayload, key String) uint {
		var x = util.CheckFloat(qt.WebViewEventPayloadGetFloat(ev, RuneSliceFromString(key)))
		return uint(x)
	},
	"ui-event-payload-get-bool": func(ev *qt.WebViewEventPayload, key String) SumValue {
		return ToBool(qt.WebViewEventPayloadGetBool(ev, RuneSliceFromString(key)))
	},
}

