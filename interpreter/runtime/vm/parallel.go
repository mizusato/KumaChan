package vm

import (
	"runtime"
)


const GoroutinePoolRunQueueSize = 32768

type GoroutinePool struct {
	runQueue  chan func()
}
func CreateGoroutinePool() GoroutinePool {
	var p = GoroutinePool {
		runQueue: make(chan func(), GoroutinePoolRunQueueSize),
	}
	var n = runtime.NumCPU()
	for i := 0; i < n; i += 1 {
		go (func() {
			for task := range p.runQueue {
				task()
			}
		})()
	}
	return p
}
func (p GoroutinePool) Dispose() {
	close(p.runQueue)
}
func (p GoroutinePool) Execute(f func()) {
	select {
	case p.runQueue <- f:
	default:
		f()
	}
}

