package rx


const sync_did_not_complete = "An action that assumed synchronous did not complete synchronously"

func runSync(action Observable, sched Scheduler, error func(Object)) (Object,bool) {
	var returned = Optional {}
	var exception = Optional {}
	var completed = false
	sched.run(action, &observer {
		context: Background(), // chained sync action cannot be interrupted
		next: func(x Object) {
			if returned.HasValue {
				panic(single_multiple_return)
			}
			returned.HasValue = true
			returned.Value = x
		},
		error: func(err Object) {
			if returned.HasValue {
				panic(single_unexpected_exception)
			}
			exception.HasValue = true
			exception.Value = err
		},
		complete: func() {
			if !(returned.HasValue) {
				panic(single_zero_return)
			}
			completed = true
		},
	})
	if exception.HasValue {
		error(exception.Value)
		return nil, false
	} else if !(completed) {
		panic(sync_did_not_complete)
	} else if !(returned.HasValue) {
		panic("something went wrong")
	} else {
		return returned.Value, true
	}
}

func (e Observable) SyncThen(f func(Object)(Observable)) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
		var x, ok = runSync(e, sched, ob.error)
		if ok {
			var next = f(x)
			sched.run(next, ob)
		}
	} }
}

func (e Observable) ChainSync(f func(Object)(Observable)) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
		var x, ok = runSync(e, sched, ob.error)
		if ok {
			var next = f(x)
			var y, ok = runSync(next, sched, ob.error)
			if ok {
				ob.next(y)
				ob.complete()
			}
		}
	} }
}

func (e Observable) TakeOneAsSingleAssumeSync() Observable {
	return Observable { func(sched Scheduler, ob *observer) {
		var ctx, ctx_dispose = ob.context.create_disposable_child()
		var completed = false
		sched.run(e, &observer {
			context:  ctx,
			next: func(val Object) {
				ctx_dispose(behaviour_cancel)
				ob.next(Optional { true, val })
				ob.complete()
			},
			error: func(err Object) {
				ctx_dispose(behaviour_terminate)
				ob.error(err)
			},
			complete: func() {
				ctx_dispose(behaviour_terminate)
				ob.next(Optional {} )
				ob.complete()
				completed = true
			},
		})
		if !(completed) {
			panic(sync_did_not_complete)
		}
	} }
}

