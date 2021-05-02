package api

import (
	"kumachan/misc/rx"
	"kumachan/misc/util"
	"kumachan/stdlib"
	. "kumachan/lang"
	"kumachan/runtime/lib/container"
	"kumachan/runtime/lib/ui"
	"kumachan/runtime/lib/ui/qt"
	"kumachan/runtime/lib/ui/vdom"
)


var UiFunctions = map[string] interface{} {
	"ui-bind": func(view qt.Widget, root rx.Observable, assets Value, h InteropContext) rx.Observable {
		var debug = h.GetDebugOptions().DebugUI
		var sched = h.Scheduler()
		var asset_index = ui.AssetIndex(func(path string) (Resource, bool) {
			return h.GetResource("web_asset", path)
		})
		var l = container.ListFrom(assets)
		var assets_used = make([] ui.Asset, l.Length())
		l.ForEach(func(i uint, v Value) {
			assets_used[i] = v.(ui.Asset)
		})
		return rx.NewSubscription(func(_ func(rx.Object)) func() {
			return ui.Bind(view, root, ui.BindOptions {
				Debug:      debug,
				Sched:      sched,
				AssetIndex: asset_index,
				AssetsUsed: assets_used,
			})
		})
	},
	"ui::TTF": func(info ProductValue, file stdlib.AssetFile) ui.TTF {
		var family = RuneSliceFromString(info.Elements[0].(String))
		var weight = RuneSliceFromString(info.Elements[1].(String))
		var style = RuneSliceFromString(info.Elements[2].(String))
		return ui.TTF {
			File: file,
			Font: ui.FontName {
				Family: family,
				Weight: weight,
				Style:  style,
			},
		}
	},
	"ui::CSS": func(file stdlib.AssetFile) ui.CSS {
		return ui.CSS { File: file }
	},
	"ui::JS": func(file stdlib.AssetFile) ui.JS {
		return ui.JS { File: file }
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
				Styles: ui.VdomMergeStyles(container.ListFrom([] *vdom.Styles {
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
				Attrs:  ui.VdomMergeAttrs(container.ListFrom([] *vdom.Attrs {
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
				Events: ui.VdomMergeEvents(container.ListFrom([] *vdom.Events {
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
		var list = container.ListFrom(children_)
		if list.Length() == 0 { return vdom.EmptyContent }
		var children = make([] *vdom.Node, list.Length())
		list.ForEach(func(i uint, item Value) {
			children[i] = item.(*vdom.Node)
		})
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
		return util.CheckReal(qt.WebViewEventPayloadGetFloat(ev, RuneSliceFromString(key)))
	},
	"ui-event-payload-get-number": func(ev *qt.WebViewEventPayload, key String) uint {
		var x = util.CheckReal(qt.WebViewEventPayloadGetFloat(ev, RuneSliceFromString(key)))
		return uint(x)
	},
	"ui-event-payload-get-bool": func(ev *qt.WebViewEventPayload, key String) SumValue {
		return ToBool(qt.WebViewEventPayloadGetBool(ev, RuneSliceFromString(key)))
	},
}

