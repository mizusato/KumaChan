package container

import (
	"fmt"
	. "kumachan/runtime/common"
)


type List struct {
	Keys   [] String
	Index  Map  // Map<string,ListEntry>
}

type ListEntry struct {
	Value     Value
	Position  uint
	Revision  uint64
}

func NewList(array Array, get_key func(Value)(String)) List {
	var keys = make([] String, array.Length)
	var index = NewStrMap()
	for i := uint(0); i < array.Length; i += 1 {
		var value = array.GetItem(i)
		var key = get_key(array.GetItem(i))
		keys[i] = key
		var result, duplicate = index.Inserted(key, ListEntry {
			Value:    value,
			Position: i,
			Revision: 0,
		})
		if duplicate {
			panic(fmt.Sprintf("duplicate key: %s", GoStringFromString(key)))
		}
		index = result
	}
	return List {
		Keys:  keys,
		Index: index,
	}
}

func (l List) mustHaveEntry(key String) ListEntry {
	var entry, exists = l.Index.Lookup(key)
	if !(exists) { panic(fmt.Sprintf("key not found: %s", GoStringFromString(key))) }
	return entry.(ListEntry)
}

func (l List) updatedIndex(key String, entry ListEntry) Map {
	var updated, override = l.Index.Inserted(key, entry)
	if !(override) { panic("something went wrong") }
	return updated
}

func (l List) Get(key String) Value {
	var entry = l.mustHaveEntry(key)
	return entry.Value
}

func (l List) Updated(key String, f func(Value)(Value)) List {
	var entry = l.mustHaveEntry(key)
	var new_index = l.updatedIndex(key, ListEntry {
		Value:    f(entry.Value),
		Position: entry.Position,
		Revision: (entry.Revision + 1),
	})
	return List {
		Keys:  l.Keys,
		Index: new_index,
	}
}

// TODO: prepend, append, swap, insertBefore, insertAfter, moveUp, moveDown

