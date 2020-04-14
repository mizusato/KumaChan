package rx


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
	return Effect{func(sched Scheduler, ob *observer) {
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
	var acc = init
	return Effect { func(sched Scheduler, ob *observer) {
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
	var acc = init
	return Effect { func(sched Scheduler, ob *observer) {
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
