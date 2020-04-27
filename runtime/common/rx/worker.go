package rx

import (
	"sync"
	"runtime"
)


type Worker struct {
	mutex     *sync.Mutex
	pending   [] func()
	incoming  chan func()
	notify    chan struct{}
	disposed  bool
}

func CreateWorker() *Worker {
	var mutex sync.Mutex
	var w = &Worker {
		mutex:    &mutex,
		pending:  make([]func(), 0),
		incoming: make(chan func()),
		notify:   make(chan struct{}, 1),
	}
	go (func() {
		for work := range w.incoming {
			w.mutex.Lock()
			w.pending = append(w.pending, work)
			select {
			case w.notify <- struct{}{}:
			default:
			}
			w.mutex.Unlock()
		}
		close(w.notify)
	})()
	go (func() {
		for range w.notify {
			w.mutex.Lock()
			if len(w.pending) > 0 {
				var current_works = w.pending
				w.pending = make([] func(), 0)
				w.mutex.Unlock()
				for _, work := range current_works {
					work()
				}
			} else {
				w.mutex.Unlock()
			}
		}
	})()
	runtime.SetFinalizer(w, func() {
		w.Dispose()
	})
	return w
}

func (w *Worker) Do(work func()) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if !(w.disposed) {
		w.incoming <- work
	}
}

func (w *Worker) Dispose() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.disposed = true
	close(w.incoming)
}
