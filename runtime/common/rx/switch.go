package rx


func (e Effect) SwitchMap(f func(Object)Effect) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.CreateChild()
		var c = new_collector(ob, dispose)
		var cur_ctx, cur_dispose = ctx.CreateChild()
		sched.run(e, &observer {
			context: ctx,
			next: func(x Object) {
				var item = f(x)
				c.new_child()
				cur_dispose()
				cur_ctx, cur_dispose = ctx.CreateChild()
				sched.run(item, &observer {
					context: cur_ctx,
					next: func(x Object) {
						c.pass(x)
					},
					error: func(e Object) {
						c.throw(e)
					},
					complete: func() {
						c.delete_child()
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
