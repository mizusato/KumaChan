package container

import (
	"fmt"
	"strconv"
	"reflect"
	. "kumachan/interpreter/runtime/def"
	. "kumachan/standalone/util/error"
)


type FlexList struct {
	keys   [] string
	index  Map // Map<string,FlexListEntry>
}
type FlexListEntry struct {
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
func FlexListFormatKey(key string) string {
	return strconv.Quote(key)
}

type FlexListIterator struct {
	List      FlexList
	NextIndex uint
}
func (it FlexListIterator) Next() (Value, Seq, bool) {
	var l = it.List
	var i = it.NextIndex
	if i < l.Length() {
		var key = l.keys[i]
		var entry = l.mustHaveEntry(key)
		var value = entry.Value
		var pair = Tuple(key, value)
		return pair, FlexListIterator {
			List:      l,
			NextIndex: (i + 1),
		}, true
	} else {
		return nil, nil, false
	}
}
func (it FlexListIterator) GetItemType() reflect.Type {
	return ValueReflectType()
}

func NewFlexList(l List, get_key func(Value)(string)) FlexList {
	var L = l.Length()
	var keys = make([] string, L)
	var index = NewMapOfStringKey()
	for i := uint(0); i < L; i += 1 {
		var value = l.at(i)
		var key = get_key(value)
		keys[i] = key
		var result, duplicate = index.Inserted(key, FlexListEntry{
			Value: value,
		})
		if duplicate {
			panic(fmt.Sprintf("FlexList: duplicate key: %s", FlexListFormatKey(key)))
		}
		index = result
	}
	return FlexList {
		keys:  keys,
		index: index,
	}
}

func (l FlexList) Inspect(inspect func(Value)ErrorMessage) ErrorMessage {
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
	return ListErrMsgItems(items, "FlexList")
}

func (l FlexList) mustHaveEntry(key string) FlexListEntry {
	var entry, exists = l.index.Lookup(key)
	if !(exists) {
		panic(fmt.Sprintf("FlexList: key not found: %s", FlexListFormatKey(key)))
	}
	return entry.(FlexListEntry)
}

func (l FlexList) mustGetIndexValueInserted(key string, v Value) Map {
	var index, override = l.index.Inserted(key, FlexListEntry{
		Value: v,
	})
	if override {
		panic(fmt.Sprintf("FlexList: duplicate key: %s", FlexListFormatKey(key)))
	}
	return index
}

func (l FlexList) mustGetIndexValueUpdated(key string, v Value) Map {
	var updated, override = l.index.Inserted(key, FlexListEntry {
		Value:    v,
	})
	if !(override) { panic("something went wrong") }
	return updated
}

func (l FlexList) Has(key string) bool {
	var _, exists = l.index.Lookup(key)
	return exists
}

func (l FlexList) Get(key string) Value {
	var entry = l.mustHaveEntry(key)
	return entry.Value
}

func (l FlexList) Length() uint {
	return uint(len(l.keys))
}

func (l FlexList) IterateKeySequence(f func(string)) {
	for _, key := range l.keys {
		f(key)
	}
}

func (l FlexList) Updated(target string, f func(Value)(Value)) FlexList {
	var entry = l.mustHaveEntry(target)
	var new_index = l.mustGetIndexValueUpdated(target, f(entry.Value))
	return FlexList {
		keys:  l.keys,
		index: new_index,
	}
}

func (l FlexList) Deleted(target string) FlexList {
	var _, new_index, ok = l.index.Deleted(target)
	if !(ok) {
		panic(fmt.Sprintf("FlexList: key not found: %s", FlexListFormatKey(target)))
	} else {
		var old_keys = l.keys
		var new_keys = make([] string, 0, (len(old_keys) - 1))
		var found = false
		for _, this := range old_keys {
			if this == target {
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
		return FlexList {
			keys:  new_keys,
			index: new_index,
		}
	}
}

func (l FlexList) Prepended(key string, v Value) FlexList {
	var new_index = l.mustGetIndexValueInserted(key, v)
	var old_keys = l.keys
	var new_len = uint(1 + len(old_keys))
	var pos = uint(0)
	var new_keys = make([] string, new_len)
	copy(new_keys[1:], old_keys)
	new_keys[pos] = key
	return FlexList {
		keys:  new_keys,
		index: new_index,
	}
}

func (l FlexList) Appended(key string, v Value) FlexList {
	var new_index = l.mustGetIndexValueInserted(key, v)
	var old_keys = l.keys
	var new_len = uint(1 + len(old_keys))
	var pos = (new_len - 1)
	var new_keys = make([] string, new_len)
	copy(new_keys, old_keys)
	new_keys[pos] = key
	return FlexList {
		keys:  new_keys,
		index: new_index,
	}
}

func (l FlexList) Inserted(key string, v Value, pos BeforeOrAfter, pivot string) FlexList {
	var new_index = l.mustGetIndexValueInserted(key, v)
	var old_keys = l.keys
	var new_keys = make([] string, 0, (1 + len(old_keys)))
	var found = false
	for _, this := range old_keys {
		if this == pivot {
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
		panic(fmt.Sprintf("FlexList: pivot key not found: %s", FlexListFormatKey(key)))
	}
	return FlexList {
		keys:  new_keys,
		index: new_index,
	}
}

func (l FlexList) Moved(target string, pos BeforeOrAfter, pivot string) FlexList {
	if target == pivot {
		return l
	} else {
		var old_keys = l.keys
		var new_keys = make([] string, 0, len(old_keys))
		var target_found = false
		var pivot_found = false
		for _, this := range old_keys {
			if this == target {
				if target_found { panic("something went wrong") }
				target_found = true
				// do nothing
			} else if this == pivot {
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
			panic(fmt.Sprintf("FlexList: target key not found: %s", FlexListFormatKey(target)))
		}
		if !(pivot_found) {
			panic(fmt.Sprintf("FlexList: pivot key not found: %s", FlexListFormatKey(target)))
		}
		return FlexList {
			keys:  new_keys,
			index: l.index,
		}
	}
}

func (l FlexList) Adjusted(target string, direction UpOrDown) (FlexList, bool) {
	var old_keys = l.keys
	var new_keys = make([] string, len(old_keys))
	var target_found = false
	var ok = false
	var skip = false
	for i, this := range old_keys {
		if this == target {
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
		panic(fmt.Sprintf("FlexList: target key not found: %s", FlexListFormatKey(target)))
	}
	return FlexList {
		keys:  new_keys,
		index: l.index,
	}, ok
}

func (l FlexList) Swapped(a_key string, b_key string) FlexList {
	if a_key == b_key {
		return l
	} else {
		var old_keys = l.keys
		var new_keys = make([] string, len(old_keys))
		copy(new_keys, old_keys)
		var a_found = false
		var b_found = false
		for i := 0; i < len(new_keys); i += 1 {
			var this = &new_keys[i]
			if *this == a_key {
				if a_found { panic("something went wrong") }
				a_found = true
				*this = b_key
			} else if *this == b_key {
				if b_found { panic("something went wrong") }
				b_found = true
				*this = a_key
			}
		}
		if !(a_found && b_found) { panic("something went wrong") }
		return FlexList {
			keys:  new_keys,
			index: l.index,
		}
	}
}

