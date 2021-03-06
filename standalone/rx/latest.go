package rx


type KeyTrackedActionVector struct {
	HasKey       func(key string) bool
	IterateKeys  func(func(string))
	CloneKeys    func() ([] string)
	GetAction    func(key string, index_source Observable) Observable
}

func (e Observable) WithLatestFrom(source Observable) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var current Optional
		c.new_child()
		sched.run(source, &observer {
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

func CombineLatest(actions ([] Observable)) Observable {
	if len(actions) == 0 {
		return NewConstant(make([] Optional, 0))
	}
	return Observable { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var values = make([] Optional, len(actions))
		for i_, e := range actions {
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

func CombineLatestWaitReady(actions ([] Observable)) Observable {
	return CombineLatest(actions).ConcatMap(func(values_ Object) Observable {
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

func KeyTrackedDynamicCombineLatestWaitReady(e Observable) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var running = make(map[string] disposeFunc)
		var values = make(map[string] Object)
		var indexes = make(map[string] ReactiveEntity)
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
				var vec = obj.(KeyTrackedActionVector)
				var keys_changed = false
				for _, key := range keys {
					if !(vec.HasKey(key)) {
						keys_changed = true
						running[key](behaviour_cancel)
						delete(running, key)
						indexes[key].Emit(nil)
						delete(indexes, key)
						delete(values, key)  // no-op when entry not existing
					}
				}
				var new_subscriptions = make([] func(), 0)
				var index_updates = make([] func(), 0)
				var i = 0
				vec.IterateKeys(func(key string) {
					var key_index = uint(i)
					var key_added_or_index_changed = false
					if i >= len(keys) || keys[i] != key {
						keys_changed = true
						key_added_or_index_changed = true
					}
					i += 1
					var _, is_running = running[key]
					if is_running && key_added_or_index_changed {
						var update = func() {
							indexes[key].commit(key_index)
						}
						index_updates = append(index_updates, update)
					}
					if !(is_running) {
						keys_changed = true
						var this_ctx, this_dispose = ctx.create_disposable_child()
						running[key] = this_dispose
						var index = CreateReactive(key_index)
						indexes[key] = index
						var index_source =
							DistinctWatch(index, func(a Object, b Object) bool {
								return a == b
							}).CompleteWhen(func(obj Object) bool {
								return obj == nil
							})
						c.new_child()
						var this_effect = vec.GetAction(key, index_source)
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
						new_subscriptions = append(new_subscriptions, run)
					}
				})
				if keys_changed {
					keys = vec.CloneKeys()
					if len(new_subscriptions) == 0 {
						emit_if_all_ready()
					}
				} else {
					if first {
						c.pass(make([] Object, 0))
					}
				}
				first = false
				for _, subscribe := range new_subscriptions {
					// subscription should happen after `keys` updated
					subscribe()
				}
				for _, index_update := range index_updates {
					index_update()
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

