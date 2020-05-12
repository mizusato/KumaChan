package rx


func Merge(effects []Effect) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		for _, item := range effects {
			c.new_child()
			sched.run(item, &observer {
				context: ctx,
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
		}
		c.parent_complete()
	} }
}

func (e Effect) MergeMap(f func(Object)Effect) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, ctx_dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, ctx_dispose)
		sched.run(e, &observer {
			context: ctx,
			next: func(x Object) {
				var item = f(x)
				c.new_child()
				sched.run(item, &observer {
					context: ctx,
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


type collector struct {
	observer          *observer
	dispose           disposeFunc
	num_children      uint
	no_more_children  bool
}

func new_collector(ob *observer, dispose disposeFunc) *collector {
	return &collector {
		observer:         ob,
		dispose:          dispose,
		num_children:     0,
		no_more_children: false,
	}
}

func (c *collector) pass(x Object) {
	c.observer.next(x)
}

func (c *collector) throw(e Object) {
	c.observer.error(e)
	c.dispose(behaviour_cancel)
}

func (c *collector) new_child() {
	c.num_children += 1
}

func (c *collector) delete_child() {
	if c.num_children == 0 { panic("something went wrong") }
	c.num_children -= 1
	if c.num_children == 0 && c.no_more_children {
		c.observer.complete()
		c.dispose(behaviour_terminate)
	}
}

func (c *collector) parent_complete() {
	c.no_more_children = true
	if c.num_children == 0 {
		c.observer.complete()
		c.dispose(behaviour_terminate)
	}
}
