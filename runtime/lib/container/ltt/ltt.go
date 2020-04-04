package ltt

import (
	. "kumachan/runtime/common"
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
	assert(ld >= rd, "violation of leftist property")
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

func (node *LTT) Merge(another *LTT, cmp Compare) *LTT {
	if node == nil { return another }
	if another == nil { return node }
	var smaller *LTT
	var bigger *LTT
	switch cmp(node.Value, another.Value) {
	case Smaller, Equal:
		smaller = node
		bigger = another
	case Bigger:
		bigger = node
		smaller = another
	default:
		panic("impossible branch")
	}
	var v = smaller.Value
	var a = smaller.Left
	var b = smaller.Right.Merge(bigger, cmp)
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

func (node *LTT) Popped(cmp Compare) (Value, *LTT, bool) {
	if node != nil {
		var rest = node.Left.Merge(node.Right, cmp)
		return node.Value, rest, true
	} else {
		return nil, nil, false
	}
}

func (node *LTT) Pushed(v Value, cmp Compare) *LTT {
	return node.Merge(Leaf(v), cmp)
}


func assert(ok bool, msg string) {
	if !ok { panic(msg) }
}
