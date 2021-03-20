package rx


func Mix(actions ([] Action), concurrent uint) Action {
	if concurrent == 0 { panic("invalid concurrent amount") }
	return Action { func(sched Scheduler, ob *observer) {
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

func (e Action) MixMap(f func(Object) Action, concurrent uint) Action {
	if concurrent == 0 {
		panic("invalid concurrent amount")
	}
	return Action { func(sched Scheduler, ob *observer) {
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


func Concat(actions ([] Action)) Action {
	return Mix(actions, 1)
}

func (e Action) ConcatMap(f func(Object) Action) Action {
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

func (qs *QueueScheduler) run(e Action, ob *observer) {
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

type queue  [] Action
func new_queue() *queue {
	var q = queue(make([] Action, 0))
	return &q
}
func (q *queue) push(e Action) {
	*q = append(*q, e)
}
func (q *queue) pop() (Action, bool) {
	if len(*q) > 0 {
		var e = (*q)[0]
		(*q)[0] = Action { nil }
		*q = (*q)[1:]
		return e, true
	} else {
		return Action {}, false
	}
}

