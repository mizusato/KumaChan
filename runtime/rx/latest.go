package rx


func (e Effect) WithLatestFrom(values Effect) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var current Optional
		c.new_child()
		sched.run(values, &observer {
			context: ctx,
			next: func(value Object) {
				current = Optional { true, value }
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
				c.pass(Pair { obj, current })
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
	if len(effects) == 0 {
		return NewSync(func() (Object, bool) {
			return make([] Optional, 0), true
		})
	}
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var values = make([] Optional, len(effects))
		for i_, e := range effects {
			var i = i_
			c.new_child()
			sched.run(e, &observer {
				context: ctx,
				next: func(obj Object) {
					var has_saved = &(values[i].HasValue)
					var saved_latest = &(values[i].Value)
					*saved_latest = obj
					*has_saved = true
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

func CombineLatestWaitAll(effects ([] Effect)) Effect {
	return CombineLatest(effects).ConcatMap(func(opt_values_ Object) Effect {
		var opt_values = opt_values_.([] Optional)
		var values = make([] Object, len(opt_values))
		var ok = true
		for i := 0; i < len(opt_values); i += 1 {
			var opt = opt_values[i]
			values[i] = opt.Value
			if !(opt.HasValue) {
				ok = false
			}
		}
		if ok {
			return NewSyncSequence(func(next func(Object)) (bool, Object) {
				next(values)
				return true, nil
			})
		} else {
			return NewSyncSequence(func(next func(Object)) (bool, Object) {
				return true, nil
			})
		}
	})
}

