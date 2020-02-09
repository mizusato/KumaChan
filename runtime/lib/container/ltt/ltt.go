package ltt

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/lib/container/order"
)

/* Functional Leftist Tree: Implementation of Queue and Priority Queue */
type LTT struct {
	Value  Value
	Left   *LTT
	Right  *LTT
	Dist   int
}

func Node(v Value) *LTT {
	return &LTT {
		Value: v,
		Left:  nil,
		Right: nil,
		Dist:  1,
	}
}

func (node *LTT) Merge(another *LTT, cmp order.Compare) *LTT {
	if node == nil { return another }
	if another == nil { return node }
	var smaller *LTT
	var bigger *LTT
	switch cmp(node.Value, another.Value) {
	case order.Smaller, order.Equal:
		smaller = node
		bigger = another
	case order.Bigger:
		bigger = node
		smaller = another
	default:
		panic("impossible branch")
	}
	var root_value = smaller.Value
	var c1 = smaller.Left
	var c2 = smaller.Right.Merge(bigger, cmp)
	if c1.Dist >= c2.Dist {
		return &LTT {
			Value: root_value,
			Left: c1,
			Right: c2,
			Dist: (c2.Dist + 1),
		}
	} else {
		return &LTT {
			Value: root_value,
			Left: c2,
			Right: c1,
			Dist: (c1.Dist + 1),
		}
	}
}

func (node *LTT) Pop(cmp order.Compare) (Value, *LTT, bool) {
	var merged = node.Left.Merge(node.Right, cmp)
	return node.Value, merged, (merged != nil)
}

func (node *LTT) Push(v Value, cmp order.Compare) *LTT {
	return node.Merge(Node(v), cmp)
}
