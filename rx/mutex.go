package rx


type Mutex struct {
	resource  Object
	sched     *QueueScheduler
}

func CreateMutex(res Object, sched Scheduler) *Mutex {
	return &Mutex {
		resource: res,
		sched:    QueueSchedulerFrom(sched, 1),
	}
}

func NewMutex(res Object) Action {
	return Action { func(sched Scheduler, ob *observer) {
		ob.next(CreateMutex(res, sched))
		ob.complete()
	} }
}

func (mu *Mutex) Lock(mutation func(Object)(Action)) Action {
	return Action { func(sched Scheduler, ob *observer) {
		mu.sched.run(mutation(mu.resource), &observer {
			context: Background(),  // atomic mutation is NOT cancellable
			next: ob.next,
			error: func(_ Object) {
				panic("unexpected exception")
			},
			complete: ob.complete,
		})
	} }
}

