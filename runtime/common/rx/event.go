package rx

import "runtime"


const MinimumEventChannelBufferSize = 32768

type event struct {
	kind      event_kind
	payload   Object
	observer  *observer
}

type event_kind int
const (
	ev_next  event_kind  =  iota
	ev_error
	ev_complete
)

type EventLoop struct {
	channel  chan event
}

func SpawnEventLoop() *EventLoop {
	return SpawnEventLoopWithBufferSize(MinimumEventChannelBufferSize)
}

func SpawnEventLoopWithBufferSize(buf_size uint) *EventLoop {
	if buf_size < MinimumEventChannelBufferSize {
		buf_size = MinimumEventChannelBufferSize
	}
	var channel = make(chan event, buf_size)
	go (func() {
		runtime.LockOSThread()
		for {
			select {
			case ev := <- channel:
				switch ev.kind {
				case ev_next:
					ev.observer.next(ev.payload)
				case ev_error:
					ev.observer.error(ev.payload)
				case ev_complete:
					ev.observer.complete()
				}
			default:
				continue
			}
		}
	})()
	return &EventLoop { channel }
}

func (el *EventLoop) dispatch(ev event) {
	el.channel <- ev
}


type TrivialScheduler struct {
	EventLoop  *EventLoop
}

func (sched TrivialScheduler) dispatch(ev event) {
	sched.EventLoop.dispatch(ev)
}

func (sched TrivialScheduler) run(effect Effect, ob *observer) {
	var terminated = false
	effect.action(sched, &observer {
		context: ob.context,
		next: func(x Object) {
			if !terminated && !ob.context.disposed {
				ob.next(x)
			}
		},
		error: func(e Object) {
			if !terminated && !ob.context.disposed {
				terminated = true
				ob.error(e)
			}
		},
		complete: func() {
			if !terminated && !ob.context.disposed {
				terminated = true
				ob.complete()
			}
		},
	})
}
