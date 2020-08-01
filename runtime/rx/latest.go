package rx


func (e Effect) WithLatestFrom(values Effect) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var current_value Optional
		c.new_child()
		sched.run(values, &observer {
			context: ctx,
			next: func(value Object) {
				current_value = Optional { true, value }
			},
			error: func(err Object) {
				c.throw(err)
			},
			complete: func() {
				c.delete_child()
			},
		})
		c.new_child()
		sched.run(e, &observer {
			context:  ctx,
			next: func(obj Object) {
				c.pass(Pair { obj, current_value })
			},
			error: func(err Object) {
				c.throw(err)
			},
			complete: func() {
				c.delete_child()
			},
		})
		c.parent_complete()
	} }
}

func CombineLatest(effects ([] Effect)) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var values = make([] Optional, len(effects))
		for i, e := range effects {
			c.new_child()
			sched.run(e, &observer {
				context: ctx,
				next: func(obj Object) {
					values[i].Value = obj
					values[i].HasValue = true
					c.pass(values)
				},
				error: func(err Object) {
					c.throw(err)
				},
				complete: func() {
					c.delete_child()
				},
			})
		}
	} }
}
