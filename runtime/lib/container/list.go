package container

import (
	"fmt"
	"strconv"
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

type UpOrDown int
const (
	Up UpOrDown = iota
	Down
)

func ListFormatKey(key String) string {
	return strconv.Quote(GoStringFromString(key))
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
			Revision: 0,
		})
		if duplicate {
			panic(fmt.Sprintf("list: duplicate key: %s", ListFormatKey(key)))
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
		panic(fmt.Sprintf("list: key not found: %s", ListFormatKey(key)))
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
		panic(fmt.Sprintf("list: duplicate key: %s", ListFormatKey(key)))
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
	var new_index = l.mustGetIndexValueInserted(key, v)
	var old_keys = l.Keys
	var new_len = uint(1 + len(old_keys))
	var pos = uint(0)
	var new_keys = make([] String, new_len)
	copy(new_keys[1:], old_keys)
	new_keys[pos] = key
	return List {
		Keys:  new_keys,
		Index: new_index,
	}
}

func (l List) Appended(key String, v Value) List {
	var new_index = l.mustGetIndexValueInserted(key, v)
	var old_keys = l.Keys
	var new_len = uint(1 + len(old_keys))
	var pos = (new_len - 1)
	var new_keys = make([] String, new_len)
	copy(new_keys, old_keys)
	new_keys[pos] = key
	return List {
		Keys:  new_keys,
		Index: new_index,
	}
}

func (l List) Inserted(key String, v Value, pos BeforeOrAfter, pivot String) List {
	var new_index = l.mustGetIndexValueInserted(key, v)
	var old_keys = l.Keys
	var new_keys = make([] String, 0, (1 + len(old_keys)))
	var found = false
	for _, this := range old_keys {
		if StringCompare(this, pivot) == Equal {
			if found { panic("something went wrong") }
			found = true
			if pos == Before {
				new_keys = append(new_keys, key)
				new_keys = append(new_keys, this)
			} else if pos == After {
				new_keys = append(new_keys, this)
				new_keys = append(new_keys, key)
			}
		} else {
			new_keys = append(new_keys, this)
		}
	}
	if !(found) {
		panic(fmt.Sprintf("list: pivot key not found: %s", ListFormatKey(key)))
	}
	return List {
		Keys:  new_keys,
		Index: new_index,
	}
}

func (l List) Moved(target String, pos BeforeOrAfter, pivot String) List {
	if StringCompare(target, pivot) == Equal {
		return l
	} else {
		var old_keys = l.Keys
		var new_keys = make([]String, 0, len(old_keys))
		var target_found = false
		var pivot_found = false
		for _, this := range old_keys {
			if StringCompare(this, target) == Equal {
				if target_found { panic("something went wrong") }
				target_found = true
				// do nothing
			} else if StringCompare(this, pivot) == Equal {
				if pivot_found { panic("something went wrong") }
				pivot_found = true
				if pos == Before {
					new_keys = append(new_keys, target)
					new_keys = append(new_keys, this)
				} else if pos == After {
					new_keys = append(new_keys, this)
					new_keys = append(new_keys, target)
				}
			} else {
				new_keys = append(new_keys, this)
			}
		}
		if !(target_found) {
			panic(fmt.Sprintf("list: target key not found: %s", ListFormatKey(target)))
		}
		if !(pivot_found) {
			panic(fmt.Sprintf("list: pivot key not found: %s", ListFormatKey(target)))
		}
		return List {
			Keys:  new_keys,
			Index: l.Index,
		}
	}
}

func (l List) Adjust(target String, direction UpOrDown) (List, bool) {
	var old_keys = l.Keys
	var new_keys = make([]String, len(old_keys))
	var target_found = false
	var ok = false
	for i, this := range old_keys {
		if StringCompare(this, target) == Equal {
			if target_found { panic("something went wrong") }
			target_found = true
			if direction == Up {
				if (i - 1) < len(old_keys) {
					var prev = old_keys[i - 1]
					new_keys[i - 1] = this
					new_keys[i] = prev
					ok = true
				} else {
					new_keys[i] = this
					ok = false
				}
			} else if direction == Down {
				if (i + 1) < len(old_keys) {
					var next = old_keys[i + 1]
					new_keys[i] = next
					new_keys[i + 1] = this
					ok = true
				} else {
					new_keys[i] = this
					ok = false
				}
			} else {
				panic("impossible branch")
			}
		} else {
			new_keys[i] = this
		}
	}
	if !(target_found) {
		panic(fmt.Sprintf("list: target key not found: %s", ListFormatKey(target)))
	}
	return List {
		Keys:  new_keys,
		Index: l.Index,
	}, ok
}

func (l List) Swapped(a_key String, b_key String) List {
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

