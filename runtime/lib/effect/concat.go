package effect

import (
	"context"
	. "kumachan/runtime/common"
)

func (e Effect) ConcatMap(f func(Value)Value, concurrent uint) Effect {
	return Effect { Action: func(r EffectRunner, ob *Observer) {
		var ctx, dispose = context.WithCancel(ob.Context)
		var c = CollectorFrom(ob, ctx, dispose)
		var q = QueueRunnerFrom(r, concurrent)
		r.Run(e, &Observer {
			Context:  ctx,
			Next: func(v Value) {
				var item = EffectFrom(f(v))
				c.NewChild()
				q.Run(item, &Observer {
					Context:  ctx,
					Next: func(v Value) {
						c.Pass(v)
					},
					Error: func(e Value) {
						c.Throw(e)
					},
					Complete: func() {
						c.DeleteChild()
					},
				})
			},
			Error: func(e Value) {
				c.Throw(e)
			},
			Complete: func() {
				c.ParentComplete()
			},
		})
	} }
}


type QueueRunner struct {
	Raw         EffectRunner
	Queue       *Queue
	Running     uint
	MaxRunning  uint
}

func QueueRunnerFrom(r EffectRunner, concurrent uint) *QueueRunner {
	if concurrent == 0 { panic("invalid concurrent amount") }
	return &QueueRunner {
		Raw:        r,
		Queue:      NewQueue(),
		Running:    0,
		MaxRunning: concurrent,
	}
}

func (qr *QueueRunner) Run(e Effect, ob *Observer) {
	if qr.Running < qr.MaxRunning {
		qr.Running += 1
		qr.Raw.Run(e, &Observer {
			Context:  ob.Context,
			Next:     ob.Next,
			Error:    ob.Error,
			Complete: func() {
				ob.Complete()
				qr.Running -= 1
				var next_item, exists = qr.Queue.Pop()
				if exists {
					qr.Run(next_item, ob)
				}
			},
		})
	} else {
		qr.Queue.Push(e)
	}
}

type Queue struct {
	Data    []QueueItem
	NextId  uint64
}

type QueueItem struct {
	Id     uint64
	Value  Effect
}

func NewQueue() *Queue {
	return &Queue {
		NextId: 0,
		Data:   make([]QueueItem, 0),
	}
}

func (q *Queue) Push(e Effect) {
	q.Data = append(q.Data, QueueItem { Value: e, Id: q.NextId })
	q.NextId += 1
}

func (q *Queue) Pop() (Effect, bool) {
	var L = len(q.Data)
	if L == 0 {
		return Effect {}, false
	} else {
		var popped = q.Data[0]
		var last_index = L - 1
		var last = q.Data[last_index]
		q.Data[0] = last
		q.Data = q.Data[:last_index]
		var node = 0
		for (node*2 + 1) < last_index {
			var left = node*2 + 1
			var right = node*2 + 2
			if right < last_index {
				var node_id = q.Data[node].Id
				var left_id = q.Data[left].Id
				var right_id = q.Data[right].Id
				if node_id < left_id && node_id < right_id {
					break
				} else if left_id < right_id {
					var left_data = q.Data[left]
					q.Data[left] = q.Data[node]
					q.Data[node] = left_data
					node = left
				} else {
					var right_data = q.Data[right]
					q.Data[right] = q.Data[node]
					q.Data[node] = right_data
					node = right
				}
			} else {
				var node_id = q.Data[node].Id
				var left_id = q.Data[left].Id
				if node_id < left_id {
					break
				} else {
					var left_data = q.Data[left]
					q.Data[left] = q.Data[node]
					q.Data[node] = left_data
					node = left
				}
			}
		}
		return popped.Value, true
	}
}
