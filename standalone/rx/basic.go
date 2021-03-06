package rx


func (e Observable) StartWith(obj Object) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
		ob.next(obj)
		sched.run(e, ob)
	} }
}

func (e Observable) Map(f func(Object)(Object)) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
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

func (e Observable) Filter(f func(Object)(bool)) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
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

func (e Observable) FilterMap(f func(Object)(Object,bool)) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
		sched.run(e, &observer {
			context:  ob.context,
			next:     func(val Object) {
				var mapped, ok = f(val)
				if ok {
					ob.next(mapped)
				}
			},
			error:    ob.error,
			complete: ob.complete,
		})
	} }
}

func (e Observable) Reduce(f func(Object,Object)Object, init Object) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
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

func (e Observable) Scan(f func(Object,Object)Object, init Object) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
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

func (e Observable) Take(amount uint) Observable {
	if amount == 0 { panic("take: invalid amount 0") }
	return Observable { func(sched Scheduler, ob *observer) {
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

func (e Observable) DiscardComplete() Observable {
	return Observable { func(sched Scheduler, ob *observer) {
		sched.run(e, &observer {
			context:  ob.context,
			next:     ob.next,
			error:    ob.error,
			complete: func() {
				// no-op
			},
		})
	} }
}

