package ltt

import (
	. "kumachan/lang"
)

/* Functional Leftist Tree: Implementation of Queue and Priority Queue */
type LTT struct {
	Value  Value
	Left   *LTT
	Right  *LTT
	Dist   uint64
}

func Node(v Value, left *LTT, right *LTT) *LTT {
	var ld = left.GetDist()
	var rd = right.GetDist()
	if !(ld >= rd) { panic("violation of leftist property") }
	return &LTT {
		Value: v,
		Left:  left,
		Right: right,
		Dist:  (1 + rd),
	}
}
func Leaf(v Value) *LTT {
	return &LTT {
		Value: v,
		Left:  nil,
		Right: nil,
		Dist:  1,
	}
}

func (node *LTT) GetDist() uint64 {
	if node != nil {
		return node.Dist
	} else {
		return 0
	}
}

func (node *LTT) Merge(another *LTT, lt LessThanOperator) *LTT {
	if node == nil { return another }
	if another == nil { return node }
	var smaller *LTT
	var bigger *LTT
	if lt(node.Value, another.Value) {
		smaller = node
		bigger = another
	} else {
		bigger = node
		smaller = another
	}
	var v = smaller.Value
	var a = smaller.Left
	var b = smaller.Right.Merge(bigger, lt)
	if a.GetDist() >= b.GetDist() {
		return Node(v, a, b)
	} else {
		return Node(v, b, a)
	}
}

func (node *LTT) Top() (Value, bool) {
	if node != nil {
		return node.Value, true
	} else {
		return nil, false
	}
}

func (node *LTT) Popped(lt LessThanOperator) (Value, *LTT, bool) {
	if node != nil {
		var rest = node.Left.Merge(node.Right, lt)
		return node.Value, rest, true
	} else {
		return nil, nil, false
	}
}

func (node *LTT) Pushed(v Value, lt LessThanOperator) *LTT {
	return node.Merge(Leaf(v), lt)
}

func (node *LTT) Walk(f func(Value)) {
	var q = []*LTT { node }
	for len(q) > 0 {
		var current = q[0]
		q = q[1:]
		if current != nil {
			f(current.Value)
			q = append(q, current.Right)
			q = append(q, current.Left)
		}
	}
}


