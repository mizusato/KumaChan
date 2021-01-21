package rx


func (e Action) CompleteWhen(p func(Object)(bool)) Action {
	return Action { func(sched Scheduler, ob *observer) {
		var ctx, ctx_dispose = ob.context.create_disposable_child()
		sched.run(e, &observer {
			context:  ctx,
			next:     func(obj Object) {
				if p(obj) {
					ctx_dispose(behaviour_cancel)
					ob.complete()
				} else {
					ob.next(obj)
				}
			},
			error: func(err Object) {
				ctx_dispose(behaviour_terminate)
				ob.error(err)
			},
			complete: func() {
				ctx_dispose(behaviour_terminate)
				ob.complete()
			},
		})
	} }
}
