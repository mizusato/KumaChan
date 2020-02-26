package rx


func Concat(effects []Effect, concurrent uint) Effect {
	if concurrent == 0 { panic("invalid concurrent amount") }
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.CreateChild()
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

func (e Effect) ConcatMap(f func(Object)Effect, concurrent uint) Effect {
	if concurrent == 0 { panic("invalid concurrent amount") }
	return Effect { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.CreateChild()
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


type QueueScheduler struct {
	underlying   Scheduler
	queue        *queue
	running      uint
	max_running  uint
}

func QueueSchedulerFrom(sched Scheduler, concurrent uint) *QueueScheduler {
	if concurrent == 0 { panic("invalid concurrent amount") }
	return &QueueScheduler{
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

type queue struct {
	data     [] queue_item
	next_id  uint64
}

type queue_item struct {
	id     uint64
	value  Effect
}

func new_queue() *queue {
	return &queue {
		next_id: 0,
		data:    make([]queue_item, 0),
	}
}

func (q *queue) push(e Effect) {
	q.data = append(q.data, queue_item{ value: e, id: q.next_id})
	q.next_id += 1
}

func (q *queue) pop() (Effect, bool) {
	var L = len(q.data)
	if L == 0 {
		return Effect{}, false
	} else {
		var popped = q.data[0]
		var last_index = L - 1
		var last = q.data[last_index]
		q.data[0] = last
		q.data = q.data[:last_index]
		var node = 0
		for (node*2 + 1) < last_index {
			var left = node*2 + 1
			var right = node*2 + 2
			if right < last_index {
				var node_id = q.data[node].id
				var left_id = q.data[left].id
				var right_id = q.data[right].id
				if node_id < left_id && node_id < right_id {
					break
				} else if left_id < right_id {
					var left_data = q.data[left]
					q.data[left] = q.data[node]
					q.data[node] = left_data
					node = left
				} else {
					var right_data = q.data[right]
					q.data[right] = q.data[node]
					q.data[node] = right_data
					node = right
				}
			} else {
				var node_id = q.data[node].id
				var left_id = q.data[left].id
				if node_id < left_id {
					break
				} else {
					var left_data = q.data[left]
					q.data[left] = q.data[node]
					q.data[node] = left_data
					node = left
				}
			}
		}
		return popped.value, true
	}
}
