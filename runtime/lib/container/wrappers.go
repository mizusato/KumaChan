package container

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/lib/container/avl"
	"kumachan/runtime/lib/container/order"
	"kumachan/runtime/lib/container/ltt"
)


type Heap struct {
	LTT  *ltt.LTT
	Cmp  order.Compare
}

func NewHeap (cmp order.Compare) Heap {
	return Heap { LTT: nil, Cmp: cmp }
}

func (h Heap) From(t *ltt.LTT) Heap {
	return Heap { LTT: t, Cmp: h.Cmp }
}

func (h Heap) Push(v Value) Heap {
	return h.From(h.LTT.Pushed(v, h.Cmp))
}

func (h Heap) Pop() (Value, Heap, bool) {
	var popped, rest, exists = h.LTT.Popped(h.Cmp)
	return popped, h.From(rest), exists
}


type Set struct {
	AVL  *avl.AVL
	Cmp  order.Compare
}

func NewSet (cmp order.Compare) Set {
	return Set { AVL: nil, Cmp: cmp }
}

func (s Set) From(a *avl.AVL) Set {
	return Set { AVL: a, Cmp: s.Cmp }
}

func (s Set) Lookup(v Value) (Value, bool) {
	return s.AVL.Lookup(v, s.Cmp)
}

func (s Set) Insert(v Value) Set {
	return s.From(s.AVL.Inserted(v, s.Cmp))
}

func (s Set) Delete(v Value) (Value, Set, bool) {
	var deleted, rest, exists = s.AVL.Deleted(v, s.Cmp)
	return deleted, s.From(rest), exists
}


type Map struct {
	AVL  *avl.AVL
	Cmp  order.Compare
}

func NewMap (cmp order.Compare) Map {
	return Map {
		AVL: nil,
		Cmp: func(kv1 Value, kv2 Value) order.Ordering {
			var k1, _ = FromTuple2(kv1)
			var k2, _ = FromTuple2(kv2)
			return cmp(k1, k2)
		},
	}
}

func (m Map) From(a *avl.AVL) Map {
	return Map { AVL: a, Cmp: m.Cmp }
}

func (m Map) Lookup(k Value) (Value, bool) {
	var kv, exists = m.AVL.Lookup(k, m.Cmp)
	if exists {
		var _, v = FromTuple2(kv)
		return v, true
	} else {
		return nil, false
	}
}

func (m Map) Insert(k Value, v Value) Map {
	return m.From(m.AVL.Inserted(Tuple2(k, v), m.Cmp))
}

func (m Map) Delete(k Value) (Value, Map, bool) {
	var deleted, rest, exists = m.AVL.Deleted(Tuple2(k, nil), m.Cmp)
	var v Value
	if exists {
		_, v = FromTuple2(deleted)
	} else {
		v = nil
	}
	return v, m.From(rest), exists
}
