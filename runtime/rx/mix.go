package rx


func Mix(effects []Effect, concurrent uint) Effect {
	if concurrent == 0 { panic("invalid concurrent amount") }
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var c = new_collector(ob, dispose)
		var q_sched = QueueSchedulerFrom(sched, concurrent)
		for _, item := range effects {
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

func (e Effect) MixMap(f func(Object)Effect, concurrent uint) Effect {
	if concurrent == 0 { panic("invalid concurrent amount") }
	return Effect { func(sched Scheduler, ob *observer) {
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


func Concat(effects []Effect) Effect {
	return Mix(effects, 1)
}

func (e Effect) ConcatMap(f func(Object)Effect) Effect {
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

func (qs *QueueScheduler) run(e Effect, ob *observer) {
	if qs.running < qs.max_running {
		qs.running += 1
		qs.underlying.run(e, &observer{
			context: ob.context,
			next:    ob.next,
			error:   ob.error,
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

func (qs *QueueScheduler) RunTopLevel(e Effect, r Receiver) {
	qs.underlying.RunTopLevel(e, r)
}

type queue  [] Effect
func new_queue() *queue {
	var q = queue(make([] Effect, 0))
	return &q
}
func (q *queue) push(e Effect) {
	*q = append(*q, e)
}
func (q *queue) pop() (Effect, bool) {
	if len(*q) > 0 {
		var e = (*q)[0]
		*q = (*q)[1:]
		return e, true
	} else {
		return Effect{}, false
	}
}
