package container

import (
	. "kumachan/interpreter/runtime/def"
	. "kumachan/standalone/util/error"
	"kumachan/interpreter/runtime/lib/container/avl"
)


type Set struct {
	AVL *avl.AVL
	Cmp Compare
}
func NewSet(cmp Compare) Set {
	return Set { AVL: nil, Cmp: cmp }
}
func (s Set) IsEmpty() bool {
	return s.AVL == nil
}
func (s Set) From(a *avl.AVL) Set {
	return Set { AVL: a, Cmp: s.Cmp }
}
func (s Set) Size() uint {
	if s.AVL == nil { return 0 }
	return uint(s.AVL.Size)
}
func (s Set) Lookup(v Value) (Value, bool) {
	return s.AVL.Lookup(v, s.Cmp)
}
func (s Set) Inserted(v Value) (Set, bool) {
	var inserted, override = s.AVL.Inserted(v, s.Cmp)
	return s.From(inserted), override
}
func (s Set) Deleted(v Value) (Value, Set, bool) {
	var deleted, rest, exists = s.AVL.Deleted(v, s.Cmp)
	return deleted, s.From(rest), exists
}
func (s Set) ForEach(f func(Value)) {
	s.AVL.Walk(f)
}
func (s Set) Inspect(inspect func(Value)ErrorMessage) ErrorMessage {
	var items = make([] ErrorMessage, 0)
	s.ForEach(func(v Value) {
		items = append(items, inspect(v))
	})
	return ListErrMsgItems(items, "Set")
}

