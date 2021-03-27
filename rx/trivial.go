package rx


type TrivialScheduler struct {
	EventLoop  *EventLoop
}

func (sched TrivialScheduler) dispatch(ev event) {
	sched.EventLoop.dispatch(ev)
}

func (sched TrivialScheduler) commit(t task) {
	sched.EventLoop.commit(t)
}

func (sched TrivialScheduler) run(observable Observable, output *observer) {
	if output.context.disposed {
		panic("cannot run an observable within a disposed context")
	}
	var terminated = false
	observable.effect(sched, &observer {
		context: output.context,
		next: func(x Object) {
			if !terminated && !output.context.disposed {
				output.next(x)
			}
		},
		error: func(e Object) {
			if !terminated && !output.context.disposed {
				terminated = true
				output.error(e)
			}
		},
		complete: func() {
			if !terminated && !output.context.disposed {
				terminated = true
				output.complete()
			}
		},
	})
}

