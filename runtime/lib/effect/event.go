package effect

import (
	. "kumachan/runtime/common"
	"runtime"
)

const EventQueueBufferSize = 4096

type Event struct {
	Kind      EventKind
	Payload   Value
	Observer  *Observer
}

type EventKind  int
const (
	EV_Next  EventKind  =  iota
	EV_Error
	EV_Complete
)

type EventLoop struct {
	EventQueue  chan Event
}

func SpawnEventLoop() *EventLoop {
	var el = &EventLoop {
		EventQueue: make(chan Event, EventQueueBufferSize),
	}
	go (func() {
		runtime.LockOSThread()
		for {
			select {
			case event := <- el.EventQueue:
				switch event.Kind {
				case EV_Next:
					event.Observer.Next(event.Payload)
				case EV_Error:
					event.Observer.Error(event.Payload)
				case EV_Complete:
					event.Observer.Complete()
				default:
					panic("unknown event kind")
				}
			default:
			}
		}
	})()
	return el
}

func (el *EventLoop) Run(effect Effect, ob *Observer) {
	go (func() {
		effect.Action(el, &Observer {
			Context: ob.Context,
			Next: func(v Value) {
				el.EventQueue <- Event {
					Kind:     EV_Next,
					Payload:  v,
					Observer: ob,
				}
			},
			Error: func(e Value) {
				el.EventQueue <- Event {
					Kind:     EV_Error,
					Payload:  e,
					Observer: ob,
				}
			},
			Complete: func() {
				el.EventQueue <- Event {
					Kind:     EV_Complete,
					Payload:  nil,
					Observer: ob,
				}
			},
		})
	})()
}