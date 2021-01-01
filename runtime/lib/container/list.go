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
	Revision  uint64
}

type BeforeOrAfter int
const (
	Before BeforeOrAfter = iota
	After
)

func NewList(array Array, get_key func(Value)(String)) List {
	var keys = make([] String, array.Length)
	var index = NewStrMap()
	for i := uint(0); i < array.Length; i += 1 {
		var value = array.GetItem(i)
		var key = get_key(array.GetItem(i))
		keys[i] = key
		var result, duplicate = index.Inserted(key, ListEntry {
			Value:    value,
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
	var entry_, exists = l.Index.Lookup(key)
	if !(exists) {
		panic(fmt.Sprintf("list: key not found: %s", GoStringFromString(key)))
	}
	var entry = entry_.(ListEntry)
	return entry
}

func (l List) mustGetIndexValueInserted(key String, v Value) Map {
	var index, override = l.Index.Inserted(key, ListEntry {
		Value:    v,
		Revision: 0,
	})
	if override {
		panic(fmt.Sprintf("list: duplicate key: %s", GoStringFromString(key)))
	}
	return index
}

func (l List) mustGetIndexValueUpdated(key String, entry ListEntry, v Value) Map {
	var updated, override = l.Index.Inserted(key, ListEntry {
		Value:    v,
		Revision: (entry.Revision + 1),
	})
	if !(override) { panic("something went wrong") }
	return updated
}

func (l List) Get(key String) Value {
	var entry = l.mustHaveEntry(key)
	return entry.Value
}

func (l List) Updated(key String, f func(Value)(Value)) List {
	var entry = l.mustHaveEntry(key)
	var new_index = l.mustGetIndexValueUpdated(key, entry, f(entry.Value))
	return List {
		Keys:  l.Keys,
		Index: new_index,
	}
}

func (l List) Prepended(key String, v Value) List {
	var old_keys = l.Keys
	var new_len = uint(len(old_keys) + 1)
	var pos = uint(0)
	var new_index = l.mustGetIndexValueInserted(key, v)
	var new_keys = make([] String, new_len)
	copy(new_keys[1:], old_keys)
	new_keys[pos] = key
	return List {
		Keys:  new_keys,
		Index: new_index,
	}
}

func (l List) Appended(key String, v Value) List {
	var old_keys = l.Keys
	var new_len = uint(len(old_keys) + 1)
	var pos = (new_len - 1)
	var new_index = l.mustGetIndexValueInserted(key, v)
	var new_keys = make([] String, new_len)
	copy(new_keys, old_keys)
	new_keys[pos] = key
	return List {
		Keys:  new_keys,
		Index: new_index,
	}
}

func (l List) Swap(a_key String, b_key String) List {
	l.mustHaveEntry(a_key)
	l.mustHaveEntry(b_key)
	if StringCompare(a_key, b_key) == Equal {
		return l
	} else {
		var old_keys = l.Keys
		var new_keys = make([] String, len(old_keys))
		copy(new_keys, old_keys)
		var a_found = false
		var b_found = false
		for i := 0; i < len(new_keys); i += 1 {
			var this = &new_keys[i]
			if StringCompare(*this, a_key) == Equal {
				if a_found { panic("something went wrong") }
				a_found = true
				*this = b_key
			} else if StringCompare(*this, b_key) == Equal {
				if b_found { panic("something went wrong") }
				b_found = true
				*this = a_key
			}
		}
		if !(a_found && b_found) { panic("something went wrong") }
		return List {
			Keys:  new_keys,
			Index: l.Index,
		}
	}
}

// TODO: insert(before,after), insertNew(before,after), moveUp, moveDown

