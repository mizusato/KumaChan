package rx


const single_multiple_return = "An action that assumed to be a single-valued action emitted multiple values"
const single_zero_return = "An action that assumed to be a single-valued action completed with zero values emitted"
const single_unexpected_exception = "An action that assumed to be a single-valued action produced an unexpected exception"

func ScheduleSingle(e Action, sched Scheduler, ctx *Context) (Optional, bool) {
	var chan_ret = make(chan Object)
	var chan_err = make(chan Object)
	sched.commit(func() {
		var returned = false
		var returned_value Object
		sched.run(e, &observer {
			context: ctx,
			next: func(obj Object) {
				if returned {
					panic(single_multiple_return)
				}
				returned = true
				returned_value = obj
			},
			error: func(err Object) {
				if returned {
					panic(single_unexpected_exception)
				}
				chan_err <- err
			},
			complete: func() {
				if !returned {
					panic(single_zero_return)
				}
				chan_ret <- returned_value
			},
		})
	})
	select {
	case ret := <- chan_ret:
		return Optional { true, ret }, true
	case err := <- chan_err:
		return Optional { false, err }, true
	case <- ctx.CancelSignal():
		return Optional{}, false
	}
}

func (e Action) Then(f func(Object)(Action)) Action {
	return Action { func(sched Scheduler, ob *observer) {
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
				if returned {
					panic(single_unexpected_exception)
				}
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

func (e Action) WaitComplete() Action {
	return Action { func(sched Scheduler, ob *observer) {
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

func (e Action) TakeOneAsSingle() Action {
	return Action { func(sched Scheduler, ob *observer) {
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

