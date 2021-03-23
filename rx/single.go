package rx


const single_multiple_return = "An action that assumed to be a single-valued effect emitted multiple values"
const single_zero_return = "An action that assumed to be a single-valued effect completed with zero values emitted"
const sync_did_not_complete = "An action that assumed synchronous did not complete synchronously"

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

func (e Action) ThenAssumeSync(f func(Object)(Action)) Action {
	return Action { func(sched Scheduler, ob *observer) {
		var returned = Optional {}
		var exception = Optional {}
		var completed = false
		sched.run(e, &observer {
			context: Background(), // chained sync action cannot be interrupted
			next: func(x Object) {
				if returned.HasValue {
					panic(single_multiple_return)
				}
				returned.HasValue = true
				returned.Value = x
			},
			error: func(err Object) {
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
			ob.error(exception.Value)
		} else if !(completed) {
			panic(sync_did_not_complete)
		} else if !(returned.HasValue) {
			panic("something went wrong")
		} else {
			var next = f(returned.Value)
			sched.run(next, ob)
		}
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

func (e Action) TakeOneAsSingleAssumeSync() Action {
	return Action { func(sched Scheduler, ob *observer) {
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

