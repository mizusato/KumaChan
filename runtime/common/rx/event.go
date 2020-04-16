package rx

import "runtime"


const MinimumEventLoopBufferSize = 32768

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

type task func()

type EventLoop struct {
	event_channel  chan event
	task_channel   chan task
}

func SpawnEventLoop() *EventLoop {
	return SpawnEventLoopWithBufferSize(MinimumEventLoopBufferSize)
}

func SpawnEventLoopWithBufferSize(buf_size uint) *EventLoop {
	if buf_size < MinimumEventLoopBufferSize {
		buf_size = MinimumEventLoopBufferSize
	}
	var events = make(chan event, buf_size)
	var tasks = make(chan task, buf_size)
	go (func() {
		runtime.LockOSThread()
		for {
			select {
			case ev := <- events:
				process_event(ev)
			default:
				select {
				case t := <- tasks:
					t()
				case ev := <- events:
					process_event(ev)
				}
			}
		}
	})()
	return &EventLoop {
		event_channel: events,
		task_channel:  tasks,
	}
}

func (el *EventLoop) dispatch(ev event) {
	el.event_channel <- ev
}

func (el *EventLoop) commit(t task) {
	el.task_channel <- t
}

func process_event(ev event) {
	switch ev.kind {
	case ev_next:
		ev.observer.next(ev.payload)
	case ev_error:
		ev.observer.error(ev.payload)
	case ev_complete:
		ev.observer.complete()
	}
}