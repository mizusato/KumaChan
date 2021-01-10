package rx


func (e Effect) SwitchMap(f func(Object)Effect) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var cur_ctx, cur_dispose = ctx.create_disposable_child()
		var cur_terminated = false
		sched.run(e, &observer {
			context: ctx,
			next: func(x Object) {
				var item = f(x)
				c.new_child()
				if cur_terminated {
					cur_dispose(behaviour_terminate)
				} else {
					cur_dispose(behaviour_cancel)
				}
				cur_ctx, cur_dispose = ctx.create_disposable_child()
				sched.run(item, &observer {
					context: cur_ctx,
					next: func(x Object) {
						c.pass(x)
					},
					error: func(e Object) {
						c.throw(e)
						cur_terminated = true
					},
					complete: func() {
						c.delete_child()
						cur_terminated = true
					},
				})
			},
			error: func(e Object) {
				c.throw(e)
			},
			complete: func() {
				c.parent_complete()
			},
		})
	} }
}
