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

func (sched TrivialScheduler) run(effect Action, ob *observer) {
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

