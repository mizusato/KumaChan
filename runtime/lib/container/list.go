package container

import (
	"fmt"
	"strconv"
	"reflect"
	. "kumachan/util/error"
	. "kumachan/lang"
)


type List struct {
	keys   [] String
	index  Map // Map<string,ListEntry>
}
type ListEntry struct {
	Value  Value
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

type ListIterator struct {
	List       List
	NextIndex  uint
}
func (it ListIterator) Next() (Value, Seq, bool) {
	var l = it.List
	var i = it.NextIndex
	if i < l.Length() {
		var key = l.keys[i]
		var entry = l.mustHaveEntry(key)
		var value = entry.Value
		var pair = &ValProd { Elements: [] Value { key, value } }
		return pair, ListIterator {
			List:      l,
			NextIndex: (i + 1),
		}, true
	} else {
		return nil, nil, false
	}
}
func (it ListIterator) GetItemType() reflect.Type {
	return ValueReflectType()
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
		})
		if duplicate {
			panic(fmt.Sprintf("list: duplicate key: %s", ListFormatKey(key)))
		}
		index = result
	}
	return List {
		keys:  keys,
		index: index,
	}
}

func (l List) Inspect(inspect func(Value)ErrorMessage) ErrorMessage {
	var items = make([] ErrorMessage, 0)
	for _, key := range l.keys {
		var entry = l.mustHaveEntry(key)
		var value = entry.Value
		var item = make(ErrorMessage, 0)
		item.WriteAll(inspect(key))
		item.WriteText(TS_NORMAL, ":")
		item.Write(T_SPACE)
		item.WriteAll(inspect(value))
		items = append(items, item)
	}
	return ListErrMsgItems(items, "List")
}

func (l List) mustHaveEntry(key String) ListEntry {
	var entry, exists = l.index.Lookup(key)
	if !(exists) {
		panic(fmt.Sprintf("list: key not found: %s", ListFormatKey(key)))
	}
	return entry.(ListEntry)
}

func (l List) mustGetIndexValueInserted(key String, v Value) Map {
	var index, override = l.index.Inserted(key, ListEntry {
		Value: v,
	})
	if override {
		panic(fmt.Sprintf("list: duplicate key: %s", ListFormatKey(key)))
	}
	return index
}

func (l List) mustGetIndexValueUpdated(key String, entry ListEntry, v Value) Map {
	var updated, override = l.index.Inserted(key, ListEntry {
		Value:    v,
	})
	if !(override) { panic("something went wrong") }
	return updated
}

func (l List) Has(key String) bool {
	var _, exists = l.index.Lookup(key)
	return exists
}

func (l List) Get(key String) Value {
	var entry = l.mustHaveEntry(key)
	return entry.Value
}

func (l List) Length() uint {
	return uint(len(l.keys))
}

func (l List) IterateKeySequence(f func(String)) {
	for _, key := range l.keys {
		f(key)
	}
}

func (l List) Updated(target String, f func(Value)(Value)) List {
	var entry = l.mustHaveEntry(target)
	var new_index = l.mustGetIndexValueUpdated(target, entry, f(entry.Value))
	return List {
		keys:  l.keys,
		index: new_index,
	}
}

func (l List) Deleted(target String) List {
	var _, new_index, ok = l.index.Deleted(target)
	if !(ok) {
		panic(fmt.Sprintf("list: key not found: %s", ListFormatKey(target)))
	} else {
		var old_keys = l.keys
		var new_keys = make([] String, 0, (len(old_keys) - 1))
		var found = false
		for _, this := range old_keys {
			if StringCompare(this, target) == Equal {
				if found {
					panic("something went wrong")
				}
				found = true
				// do nothing
			} else {
				new_keys = append(new_keys, this)
			}
		}
		if !(found) {
			panic("something went wrong")
		}
		return List {
			keys:  new_keys,
			index: new_index,
		}
	}
}

func (l List) Prepended(key String, v Value) List {
	var new_index = l.mustGetIndexValueInserted(key, v)
	var old_keys = l.keys
	var new_len = uint(1 + len(old_keys))
	var pos = uint(0)
	var new_keys = make([] String, new_len)
	copy(new_keys[1:], old_keys)
	new_keys[pos] = key
	return List {
		keys:  new_keys,
		index: new_index,
	}
}

func (l List) Appended(key String, v Value) List {
	var new_index = l.mustGetIndexValueInserted(key, v)
	var old_keys = l.keys
	var new_len = uint(1 + len(old_keys))
	var pos = (new_len - 1)
	var new_keys = make([] String, new_len)
	copy(new_keys, old_keys)
	new_keys[pos] = key
	return List {
		keys:  new_keys,
		index: new_index,
	}
}

func (l List) Inserted(key String, v Value, pos BeforeOrAfter, pivot String) List {
	var new_index = l.mustGetIndexValueInserted(key, v)
	var old_keys = l.keys
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
		keys:  new_keys,
		index: new_index,
	}
}

func (l List) Moved(target String, pos BeforeOrAfter, pivot String) List {
	if StringCompare(target, pivot) == Equal {
		return l
	} else {
		var old_keys = l.keys
		var new_keys = make([] String, 0, len(old_keys))
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
			keys:  new_keys,
			index: l.index,
		}
	}
}

func (l List) Adjusted(target String, direction UpOrDown) (List, bool) {
	var old_keys = l.keys
	var new_keys = make([] String, len(old_keys))
	var target_found = false
	var ok = false
	var skip = false
	for i, this := range old_keys {
		if StringCompare(this, target) == Equal {
			if target_found { panic("something went wrong") }
			target_found = true
			if direction == Up {
				if 0 <= (i - 1) && (i - 1) < len(old_keys) {
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
					skip = true
					ok = true
				} else {
					new_keys[i] = this
					ok = false
				}
			} else {
				panic("impossible branch")
			}
		} else {
			if skip {
				skip = false
				continue
			}
			new_keys[i] = this
		}
	}
	if !(target_found) {
		panic(fmt.Sprintf("list: target key not found: %s", ListFormatKey(target)))
	}
	return List {
		keys:  new_keys,
		index: l.index,
	}, ok
}

func (l List) Swapped(a_key String, b_key String) List {
	if StringCompare(a_key, b_key) == Equal {
		return l
	} else {
		var old_keys = l.keys
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
			keys:  new_keys,
			index: l.index,
		}
	}
}

