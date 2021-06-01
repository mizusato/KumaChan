package vm2

import "kumachan/standalone/rx"

type Machine struct {
	options    Options
	parallel   GoroutinePool
	scheduler  rx.Scheduler
}

type Options struct {
	ParallelEnabled  bool
	// TODO
}

