package rx


const single_multiple_return = "An effect that assumed to be a single-valued effect emitted multiple values"
const single_zero_return = "An effect that assumed to be a single-valued effect completed with zero values emitted"

func (e Effect) Then(f func(Object)Effect) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var returned = false
		var returned_value Object
		sched.run(e, &observer {
			context: ob.context,
			next: func(x Object) {
				if returned {
					panic(single_multiple_return)
				}
				returned = true
				returned_value = x
			},
			error: func(e Object) {
				ob.error(e)
			},
			complete: func() {
				if !returned {
					panic(single_zero_return)
				}
				var next = f(returned_value)
				sched.run(next, ob)
			},
		})
	} }
}

func (e Effect) WaitComplete() Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		sched.run(e, &observer {
			context:  ob.context,
			next:     func(_ Object) {
				// do nothing
			},
			error:    ob.error,
			complete: func() {
				ob.next(nil)
				ob.complete()
			},
		})
	} }
}

func (e Effect) TakeOne() Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, ctx_dispose = ob.context.create_disposable_child()
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
			},
		})
	} }
}
