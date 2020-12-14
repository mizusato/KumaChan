package rx


func (e Effect) StartWith(obj Object) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		ob.next(obj)
		sched.run(e, ob)
	} }
}

func (e Effect) Map(f func(Object)Object) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		sched.run(e, &observer {
			context:  ob.context,
			next:     func(val Object) {
				ob.next(f(val))
			},
			error:    ob.error,
			complete: ob.complete,
		})
	} }
}

func (e Effect) Filter(f func(Object)bool) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		sched.run(e, &observer {
			context:  ob.context,
			next:     func(val Object) {
				if (f(val)) {
					ob.next(val)
				}
			},
			error:    ob.error,
			complete: ob.complete,
		})
	}}
}

func (e Effect) Reduce(f func(Object,Object)Object, init Object) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var acc = init
		sched.run(e, &observer {
			context:  ob.context,
			next:     func(val Object) {
				acc = f(acc, val)
			},
			error:    ob.error,
			complete: func() {
				ob.next(acc)
				ob.complete()
			},
		})
	} }
}

func (e Effect) Scan(f func(Object,Object)Object, init Object) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var acc = init
		sched.run(e, &observer {
			context:  ob.context,
			next:     func(val Object) {
				acc = f(acc, val)
				ob.next(acc)
			},
			error:    ob.error,
			complete: ob.complete,
		})
	} }
}

func (e Effect) Take(amount uint) Effect {
	if amount == 0 { panic("take: invalid amount 0") }
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, ctx_dispose = ob.context.create_disposable_child()
		var taken = uint(0)
		sched.run(e, &observer {
			context:  ctx,
			next: func(val Object) {
				ob.next(val)
				taken += 1
				if taken == amount {
					ctx_dispose(behaviour_cancel)
					ob.complete()
				}
			},
			error: func(err Object) {
				ctx_dispose(behaviour_terminate)
				ob.error(err)
			},
			complete: func() {
				ctx_dispose(behaviour_terminate)
			},
		})
	} }
}
