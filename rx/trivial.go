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

func (sched TrivialScheduler) RunTopLevel(e Effect, r Receiver) {
	sched.EventLoop.commit(func() {
		sched.run(e, &observer {
			context:  r.Context,
			next: func(x Object) {
				if r.Values != nil {
					r.Values <- x
				}
			},
			error: func(e Object) {
				if r.Error != nil {
					r.Error <- e
					close(r.Error)
				}
				if r.Terminate != nil {
					r.Terminate <- false
				}
			},
			complete: func() {
				if r.Values != nil {
					close(r.Values)
				}
				if r.Terminate != nil {
					r.Terminate <- true
				}
			},
		})
	})
}

