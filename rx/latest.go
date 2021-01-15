package rx


type KeyTrackedEffectVector struct {
	HasKey       func(key string) bool
	IterateKeys  func(func(string))
	CloneKeys    func() ([] string)
	GetEffect    func(key string) Effect  // effect won't change if key persists
}

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
					var values_clone = make([] Optional, len(values))
					copy(values_clone, values)
					c.pass(values_clone)
				},
				error: func(err Object) {
					c.throw(err)
				},
				complete: func() {
					c.delete_child()
				},
			})
		}
		c.parent_complete()
	} }
}

func CombineLatestWaitReady(effects ([] Effect)) Effect {
	return CombineLatest(effects).ConcatMap(func(values_ Object) Effect {
		var values = values_.([] Optional)
		var ready_values = make([] Object, len(values))
		var ok = true
		for i := 0; i < len(values); i += 1 {
			var opt = values[i]
			ready_values[i] = opt.Value
			if !(opt.HasValue) {
				ok = false
			}
		}
		if ok {
			return NewSyncSequence(func(next func(Object)) (bool, Object) {
				next(ready_values)
				return true, nil
			})
		} else {
			return NewSyncSequence(func(next func(Object)) (bool, Object) {
				return true, nil
			})
		}
	})
}

func KeyTrackedDynamicCombineLatestWaitReady(e Effect) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var running = make(map[string] disposeFunc)
		var values = make(map[string] Object)
		var keys = make([] string, 0)
		var first = true
		var emit_if_all_ready = func() {
			var ready_values = make([] Object, len(keys))
			var all_ready = true
			for i, k := range keys {
				var v, exists = values[k]
				if exists {
					ready_values[i] = v
				} else {
					all_ready = false
					break
				}
			}
			if all_ready {
				c.pass(ready_values)
			} else {
				// do nothing
			}
		}
		sched.run(e, &observer {
			context:  ctx,
			next: func(obj Object) {
				var vec = obj.(KeyTrackedEffectVector)
				var keys_changed = false
				for _, key := range keys {
					if !(vec.HasKey(key)) {
						keys_changed = true
						running[key](behaviour_cancel)
						delete(running, key)
						delete(values, key)  // no-op when entry not existing
					}
				}
				var run_queue = make([] func(), 0)
				var i = 0
				vec.IterateKeys(func(key string) {
					if i < len(keys) && keys[i] != key {
						keys_changed = true
					}
					i += 1
					var _, is_running = running[key]
					if !(is_running) {
						keys_changed = true
						var this_ctx, this_dispose = ctx.create_disposable_child()
						running[key] = this_dispose
						c.new_child()
						var this_effect = vec.GetEffect(key)
						var run = func() {
							sched.run(this_effect, &observer {
								context:  this_ctx,
								next: func(obj Object) {
									values[key] = obj
									emit_if_all_ready()
								},
								error: func(err Object) {
									c.throw(err)
								},
								complete: func() {
									c.delete_child()
								},
							})
						}
						run_queue = append(run_queue, run)
					}
				})
				if keys_changed {
					keys = vec.CloneKeys()
					if len(run_queue) == 0 {
						emit_if_all_ready()
					}
				} else {
					if first {
						c.pass(make([] Object, 0))
					}
				}
				first = false
				for _, run := range run_queue {
					// subscription should happen after `keys` updated
					run()
				}
			},
			error: func(err Object) {
				c.throw(err)
			},
			complete: func() {
				c.parent_complete()
			},
		})
	} }
}

