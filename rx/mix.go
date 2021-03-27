package rx


func Mix(actions ([] Observable), concurrent uint) Observable {
	if concurrent == 0 { panic("invalid concurrent amount") }
	return Observable { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var q_sched = QueueSchedulerFrom(sched, concurrent)
		for _, item := range actions {
			c.new_child()
			q_sched.run(item, &observer {
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

func (e Observable) MixMap(f func(Object) Observable, concurrent uint) Observable {
	if concurrent == 0 {
		panic("invalid concurrent amount")
	}
	return Observable { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var q_sched = QueueSchedulerFrom(sched, concurrent)
		sched.run(e, &observer {
			context: ctx,
			next: func(x Object) {
				var item = f(x)
				c.new_child()
				q_sched.run(item, &observer {
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


func Concat(actions ([] Observable)) Observable {
	return Mix(actions, 1)
}

func (e Observable) ConcatMap(f func(Object) Observable) Observable {
	return e.MixMap(f, 1)
}


type QueueScheduler struct {
	underlying   Scheduler
	queue        *queue
	running      uint
	max_running  uint
}

func QueueSchedulerFrom(sched Scheduler, concurrent uint) *QueueScheduler {
	if concurrent == 0 { panic("invalid concurrent amount") }
	return &QueueScheduler {
		underlying:  sched,
		queue:       new_queue(),
		running:     0,
		max_running: concurrent,
	}
}

func (qs *QueueScheduler) run(e Observable, ob *observer) {
	if qs.running < qs.max_running {
		qs.running += 1
		qs.underlying.run(e, &observer {
			context:  ob.context,
			next:     ob.next,
			error:    ob.error,
			complete: func() {
				ob.complete()
				qs.running -= 1
				var next_item, exists = qs.queue.pop()
				if exists {
					qs.run(next_item, ob)
				}
			},
		})
	} else {
		qs.queue.push(e)
	}
}

func (qs *QueueScheduler) dispatch(ev event) {
	qs.underlying.dispatch(ev)
}

func (qs *QueueScheduler) commit(t task) {
	qs.underlying.commit(t)
}

type queue  [] Observable
func new_queue() *queue {
	var q = queue(make([] Observable, 0))
	return &q
}
func (q *queue) push(e Observable) {
	*q = append(*q, e)
}
func (q *queue) pop() (Observable, bool) {
	if len(*q) > 0 {
		var e = (*q)[0]
		(*q)[0] = Observable { nil }
		*q = (*q)[1:]
		return e, true
	} else {
		return Observable {}, false
	}
}

