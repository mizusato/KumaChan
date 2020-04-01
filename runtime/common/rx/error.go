package rx


func (e Effect) Catch(f func(Object)Effect) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx = ob.context
		sched.run(e, &observer {
			context: ctx,
			next: func(x Object) {
				ob.next(x)
			},
			error: func(e Object) {
				var handler = f(e)
				sched.run(handler, &observer {
					context: ctx,
					next: func(x Object) {
						// do nothing
					},
					error: func(e Object) {
						// do nothing
					},
					complete: func() {
						ob.complete()
					},
				})
			},
			complete: func() {
				ob.complete()
			},
		})
	} }
}
