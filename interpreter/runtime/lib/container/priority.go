package container

import (
	. "kumachan/interpreter/def"
	. "kumachan/standalone/util/error"
	"kumachan/interpreter/runtime/lib/container/ltt"
)


type PriorityQueue struct {
	LTT  *ltt.LTT
	LtOp LessThanOperator
}
func NewPriorityQueue(lt LessThanOperator) PriorityQueue {
	return PriorityQueue { LTT: nil, LtOp: lt }
}
func (h PriorityQueue) From(t *ltt.LTT) PriorityQueue {
	return PriorityQueue { LTT: t, LtOp: h.LtOp }
}
func (h PriorityQueue) Pushed(v Value) PriorityQueue {
	return h.From(h.LTT.Pushed(v, h.LtOp))
}
func (h PriorityQueue) Popped() (Value, PriorityQueue, bool) {
	var popped, rest, exists = h.LTT.Popped(h.LtOp)
	return popped, h.From(rest), exists
}
func (h PriorityQueue) Top() (Value, bool) {
	return h.LTT.Top()
}
func (h PriorityQueue) Inspect(inspect func(Value)(ErrorMessage)) ErrorMessage {
	var items = make([] ErrorMessage, 0)
	h.LTT.Walk(func(v Value) {
		items = append(items, inspect(v))
	})
	return ListErrMsgItems(items, "PriorityQueue")
}

