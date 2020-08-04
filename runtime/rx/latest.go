package rx

import "reflect"


func RefEqual(a interface{}, b interface{}) bool {
	var x = reflect.ValueOf(a)
	var y = reflect.ValueOf(b)
	if x.Kind() == reflect.Ptr && y.Kind() == reflect.Ptr {
		if x.Pointer() == y.Pointer() {
			return true
		}
	}
	return false
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
				if current.HasValue && RefEqual(obj, current.Value) {
					return
				} else {
					c.pass(Pair { obj, current })
				}
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
		for i_, e := range effects {
			var i = i_
			c.new_child()
			sched.run(e, &observer {
				context: ctx,
				next: func(obj Object) {
					var has_saved = &(values[i].HasValue)
					var saved_latest = &(values[i].Value)
					if *has_saved && RefEqual(obj, *saved_latest) {
						return
					} else {
						*saved_latest = obj
						*has_saved = true
						c.pass(values)
					}
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