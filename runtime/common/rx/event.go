package rx

import "runtime"


const MinimumEventQueueBufferSize = 32768

type Event struct {
	Kind     EventKind
	Payload  Object
	Observer *Observer
}

type EventKind  int
const (
	EV_Next  EventKind  =  iota
	EV_Error
	EV_Complete
)

type EventLoop struct {
	queue  chan Event
}

func SpawnEventLoop() *EventLoop {
	return SpawnEventLoopWithBufferSize(MinimumEventQueueBufferSize)
}

func SpawnEventLoopWithBufferSize(buf_size uint) *EventLoop {
	if buf_size < MinimumEventQueueBufferSize {
		buf_size = MinimumEventQueueBufferSize
	}
	var el = &EventLoop {
		queue: make(chan Event, buf_size),
	}
	go (func() {
		runtime.LockOSThread()
		for {
			select {
			case event := <- el.queue:
				if event.Observer.Context.disposed {
					continue
				}
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
		effect.Action(el, &Observer{
			Context: ob.Context,
			Next: func(v Object) {
				el.queue <- Event {
					Kind:     EV_Next,
					Payload:  v,
					Observer: ob,
				}
			},
			Error: func(e Object) {
				el.queue <- Event {
					Kind:     EV_Error,
					Payload:  e,
					Observer: ob,
				}
			},
			Complete: func() {
				el.queue <- Event {
					Kind:     EV_Complete,
					Payload:  nil,
					Observer: ob,
				}
			},
		})
	})()
}