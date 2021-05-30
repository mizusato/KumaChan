package container

import (
	. "kumachan/interpreter/def"
)


type Queue struct {
	Heap       PriorityQueue
	NextNumber uint64
}
type QueueEntry struct {
	Value   Value
	Number  uint64
}
func NewQueue() Queue {
	return Queue {
		Heap: NewPriorityQueue(func(v1 Value, v2 Value) bool {
			var e1 = v1.(QueueEntry)
			var e2 = v2.(QueueEntry)
			return (e1.Number < e2.Number)
		}),
		NextNumber: 0,
	}
}
func (q Queue) Pushed(v Value) Queue {
	var n = q.NextNumber
	return Queue {
		Heap: q.Heap.Pushed(QueueEntry {
			Value:  v,
			Number: n,
		}),
		NextNumber: (n + 1),
	}
}
func (q Queue) Shifted() (Value, Queue, bool) {
	var e, rest, ok = q.Heap.Popped()
	if ok {
		return e.(QueueEntry).Value, Queue {
			Heap:       rest,
			NextNumber: q.NextNumber,
		}, true
	} else {
		return nil, Queue{}, false
	}
}
func (q Queue) Front() (Value, bool) {
	var e, ok = q.Heap.Top()
	if ok {
		return e.(QueueEntry).Value, true
	} else {
		return nil, false
	}
}

