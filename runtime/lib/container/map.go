package container

import (
	. "kumachan/lang"
	. "kumachan/misc/util/error"
	"kumachan/runtime/lib/container/avl"
)


type Map struct {
	AVL  *avl.AVL
	Cmp  Compare
}
type MapEntry struct {
	Key    Value
	Value  Value
}
func NewMap(cmp Compare) Map {
	return Map {
		AVL: nil,
		Cmp: func(e1 Value, e2 Value) Ordering {
			return cmp(e1.(MapEntry).Key, e2.(MapEntry).Key)
		},
	}
}
func NewMapOfStringKey() Map {
	return NewMap(func(k1 Value, k2 Value) Ordering {
		var s1 = k1.(String)
		var s2 = k2.(String)
		return StringFastCompare(s1, s2)
	})
}
func (m Map) IsEmpty() bool {
	return m.AVL == nil
}
func (m Map) From(a *avl.AVL) Map {
	return Map { AVL: a, Cmp: m.Cmp }
}
func (m Map) Size() uint {
	if m.AVL == nil { return 0 }
	return uint(m.AVL.Size)
}
func (m Map) Lookup(k Value) (Value, bool) {
	var kv, exists = m.AVL.Lookup(MapEntry { Key: k }, m.Cmp)
	if exists {
		return kv.(MapEntry).Value, true
	} else {
		return nil, false
	}
}
func (m Map) Inserted(k Value, v Value) (Map, bool) {
	var entry = MapEntry {
		Key:   k,
		Value: v,
	}
	var inserted, override = m.AVL.Inserted(entry, m.Cmp)
	return m.From(inserted), override
}
func (m Map) Deleted(k Value) (Value, Map, bool) {
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
func (m Map) ForEach(f func(k Value, v Value)) {
	m.AVL.Walk(func(kv Value) {
		var entry = kv.(MapEntry)
		f(entry.Key, entry.Value)
	})
}
func (m Map) Inspect(inspect func(Value)ErrorMessage) ErrorMessage {
	var items = make([] ErrorMessage, 0)
	m.ForEach(func(k Value, v Value) {
		var entry_msg = make(ErrorMessage, 0)
		entry_msg.WriteAll(inspect(k))
		entry_msg.WriteText(TS_NORMAL, ":")
		entry_msg.Write(T_SPACE)
		entry_msg.WriteAll(inspect(v))
		items = append(items, entry_msg)
	})
	return ListErrMsgItems(items, "Map")
}

