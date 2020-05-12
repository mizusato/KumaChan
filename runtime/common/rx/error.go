package rx


const invalid_no_except = "An effect that assumed to be a no-exception effect has thrown an error"

func Throw(e Object) Effect {
	return Effect { func(_ Scheduler, ob *observer) {
		ob.error(e)
	} }
}

func (e Effect) NoExcept() Effect {
	return Effect{ func(sched Scheduler, ob *observer) {
		sched.run(e, &observer {
			context:  ob.context,
			next:     ob.next,
			error:    func(_ Object) {
				panic(invalid_no_except)
			},
			complete: ob.complete,
		})
	} }
}

func (e Effect) Catch(f func(Object)Effect) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		sched.run(e, &observer {
			context: ob.context,
			next: func(x Object) {
				ob.next(x)
			},
			error: func(err Object) {
				var caught_effect = f(err)
				sched.run(caught_effect, ob)
			},
			complete: func() {
				ob.complete()
			},
		})
	} }
}

func (e Effect) CatchRetry(f func(Object)Effect) Effect {
	var try Effect
	try = Effect { func(sched Scheduler, ob *observer) {
		sched.run(e, &observer {
			context: ob.context,
			next: func(x Object) {
				ob.next(x)
			},
			error: func(err Object) {
				var caught_effect =
					f(err).NoExcept().Then(func(retry Object) Effect {
						if retry.(bool) {
							return try
						} else {
							return Throw(err)
						}
					})
				sched.run(caught_effect, ob)
			},
			complete: func() {
				ob.complete()
			},
		})
	} }
	return try
}

func (e Effect) CatchThrow(error_mapper func(Object)Object) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		sched.run(e, &observer {
			context:  ob.context,
			next:     ob.next,
			error:    func(err Object) {
				ob.error(error_mapper(err))
			},
			complete: ob.complete,
		})
	} }
}
