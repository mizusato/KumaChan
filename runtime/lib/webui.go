package lib

import (
	"sync"
	"kumachan/qt"
	"kumachan/runtime/rx"
	. "kumachan/runtime/common"
	"kumachan/runtime/lib/container"
)


var __WebUiLoaded = false
var __WebUiLoadedMutex sync.Mutex

var WebUiFunctions = map[string] interface{} {
	"webui-debug": func(dom rx.Effect, h MachineHandle) rx.Effect {
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
				qt.WebUiDebug(title)
				del_title()
				wait <- struct{}{}
			})
			__WebUiLoaded = true
			__WebUiLoadedMutex.Unlock()
			<- wait
			var window = qt.WebUiGetWindow()
			qt.Connect(window, "loadFinished()", func() {
				var sched = h.GetScheduler()
				var dom_ch = make(chan rx.Object, 16384)
				sched.RunTopLevel(dom, rx.Receiver {
					Context: rx.Background(),
					Values:  dom_ch,
				})
				go (func() {
					for update := range dom_ch {
						var root = update.(qt.DomNode)
						qt.CommitTask(func() {
							qt.WebUiUpdateVDOM(root)
						})
					}
				})()
			})
			qt.Connect(window, "handlerDetached()", func() {
				qt.UnregisterDetachedEventHandler()
			})
		})
	},
	"webui-dom-node": func(tag String, styles Value, events Value, content qt.DomContent) qt.DomNode {
		var style_arr = container.ArrayFrom(styles)
		var event_arr = container.ArrayFrom(events)
		var tag_ = RuneSliceFromString(tag)
		var styles_ = make([] qt.DomStyle, style_arr.Length)
		for i := uint(0); i < style_arr.Length; i += 1 {
			styles_[i] = style_arr.GetItem(i).(qt.DomStyle)
		}
		var events_ = make([] qt.DomEvent, event_arr.Length)
		for i := uint(0); i < event_arr.Length; i += 1 {
			events_[i] = event_arr.GetItem(i).(qt.DomEvent)
		}
		var ch = make(chan qt.DomNode)
		qt.MakeSureInitialized()
		qt.CommitTask(func() {
			ch <- qt.WebUiNewDomNode(tag_, qt.DomProps {
				Styles: styles_,
				Events: events_,
			}, content)
		})
		return <- ch
	},
	"webui-dom-style": func(key String, value String) qt.DomStyle {
		return qt.DomStyle {
			Key:   RuneSliceFromString(key),
			Value: RuneSliceFromString(value),
		}
	},
	"webui-dom-event": func(name String, prevent SumValue, stop SumValue, sink rx.Sink, h MachineHandle) qt.DomEvent {
		return qt.DomEvent {
			Name:    RuneSliceFromString(name),
			Prevent: BoolFrom(prevent),
			Stop:    BoolFrom(stop),
			Handler: func(payload qt.VariantMap) {
				h.GetScheduler().RunTopLevel(rx.CreateBlockingEffect(func() (rx.Object, bool) {
					sink.Send(payload)
					return nil, true
				}), rx.Receiver {})
			},
		}
	},
	"webui-dom-text": func(text String) qt.DomContent {
		return qt.DomText(RuneSliceFromString(text))
	},
	"webui-dom-children": func(children Value) qt.DomContent {
		var arr = container.ArrayFrom(children)
		var children_ = make([] qt.DomNode, arr.Length)
		for i := uint(0); i < arr.Length; i += 1 {
			children_[i] = arr.GetItem(i).(qt.DomNode)
		}
		return qt.DomChildren(children_)
	},
}
