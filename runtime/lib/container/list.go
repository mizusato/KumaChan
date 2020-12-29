package container

import (
	. "kumachan/runtime/common"
	"fmt"
	"reflect"
)


type List struct {
	Data   interface {}  // slice<T>
	Index  Map           // Map<string,ListIndexEntry>
}

type ListIndexEntry struct {
	Position  uint
	Revision  uint64
}

func NewList(array Array, get_key func(Value)(String)) List {
	var data = array.CopyAsSlice(array.ItemType)
	var index = NewStrMap()
	for i := uint(0); i < array.Length; i += 1 {
		var key = get_key(array.GetItem(i))
		var result, duplicate = index.Inserted(key, ListIndexEntry {
			Position: i,
			Revision: 0,
		})
		if duplicate {
			panic(fmt.Sprintf("duplicate key: %s", GoStringFromString(key)))
		}
		index = result
	}
	return List {
		Data:  data,
		Index: index,
	}
}

func (l List) indexEntry(key String) ListIndexEntry {
	var entry, exists = l.Index.Lookup(key)
	if !(exists) { panic(fmt.Sprintf("key not found: %s", GoStringFromString(key))) }
	return entry.(ListIndexEntry)
}

func (l List) updatedIndex(key String, entry ListIndexEntry) Map {
	var updated, override = l.Index.Inserted(key, entry)
	if !(override) { panic("something went wrong") }
	return updated
}

func (l List) Get(key String) Value {
	var entry = l.indexEntry(key)
	return reflect.ValueOf(l.Data).Index(int(entry.Position)).Interface()
}

func (l List) Updated(key String, f func(Value)(Value)) List {
	var entry = l.indexEntry(key)
	var old_rv = reflect.ValueOf(l.Data)
	var new_rv = reflect.MakeSlice(old_rv.Type(), old_rv.Len(), old_rv.Cap())
	reflect.Copy(new_rv, old_rv)
	var item_rv = old_rv.Index(int(entry.Position))
	item_rv.Set(reflect.ValueOf(f(item_rv.Interface())))
	var new_data = new_rv.Interface()
	var new_index = l.updatedIndex(key, ListIndexEntry {
		Position: entry.Position,
		Revision: (entry.Revision + 1),
	})
	return List {
		Data:  new_data,
		Index: new_index,
	}
}

// TODO: prepend, append, swap, insertBefore, insertAfter, moveUp, moveDown

