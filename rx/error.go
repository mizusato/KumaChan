package rx


const invalid_no_except = "An effect that assumed to be a no-exception effect has thrown an error"

func Throw(e Object) Observable {
	return Observable { func(_ Scheduler, ob *observer) {
		ob.error(e)
	} }
}

func (e Observable) NoExcept() Observable {
	return Observable { func(sched Scheduler, ob *observer) {
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

func (e Observable) Catch(f func(Object) Observable) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
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

func (e Observable) CatchRetry(f func(Object) Observable) Observable {
	var try Observable
	try = Observable { func(sched Scheduler, ob *observer) {
		sched.run(e, &observer {
			context: ob.context,
			next: func(x Object) {
				ob.next(x)
			},
			error: func(err Object) {
				var caught_effect =
					f(err).NoExcept().Then(func(retry Object) Observable {
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

func (e Observable) CatchThrow(error_mapper func(Object)Object) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
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
