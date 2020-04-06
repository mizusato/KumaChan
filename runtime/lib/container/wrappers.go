package container

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/lib/container/avl"
	"kumachan/runtime/lib/container/ltt"
)


type Heap struct {
	LTT   *ltt.LTT
	LtOp  LessThanOperator
}

func NewHeap (lt LessThanOperator) Heap {
	return Heap { LTT: nil, LtOp: lt }
}

func (h Heap) From(t *ltt.LTT) Heap {
	return Heap { LTT: t, LtOp: h.LtOp }
}

func (h Heap) Push(v Value) Heap {
	return h.From(h.LTT.Pushed(v, h.LtOp))
}

func (h Heap) Pop() (Value, Heap, bool) {
	var popped, rest, exists = h.LTT.Popped(h.LtOp)
	return popped, h.From(rest), exists
}


type Set struct {
	AVL *avl.AVL
	Cmp Compare
}

func NewSet (cmp Compare) Set {
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
	Cmp  Compare
}
type MapEntry struct {
	Key    Value
	Value  Value
}

func NewMap (cmp Compare) Map {
	return Map {
		AVL: nil,
		Cmp: func(e1 Value, e2 Value) Ordering {
			return cmp(e1.(MapEntry).Key, e2.(MapEntry).Key)
		},
	}
}

func (m Map) From(a *avl.AVL) Map {
	return Map { AVL: a, Cmp: m.Cmp }
}

func (m Map) Lookup(k Value) (Value, bool) {
	var kv, exists = m.AVL.Lookup(k, m.Cmp)
	if exists {
		return kv.(MapEntry).Value, true
	} else {
		return nil, false
	}
}

func (m Map) Insert(k Value, v Value) Map {
	var entry = MapEntry {
		Key:   k,
		Value: v,
	}
	return m.From(m.AVL.Inserted(entry, m.Cmp))
}

func (m Map) Delete(k Value) (Value, Map, bool) {
	var entry = MapEntry {
		Key:   k,
		Value: nil,
	}
	var deleted, rest, exists = m.AVL.Deleted(entry, m.Cmp)
	var v Value
	if exists {
		v = deleted.(MapEntry).Value
	} else {
		v = nil
	}
	return v, m.From(rest), exists
}
