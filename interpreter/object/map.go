package object

import ."kumachan/interpreter/assertion"

const MapShrinkFactor = 4
const __MapNoSuchEntry = "Map: no such entry"
const __MapInvalidObject = "Map: invalid value object"

type StrMap struct {
	__Data      map[string]Object
	__Added     int
}

type IntMap struct {
	__Data      map[int]Object
	__Added     int
}

func NewStrMap () *StrMap {
	return &StrMap {
		__Data: make(map[string]Object),
		__Added: 0,
	}
}

func NewIntMap () *IntMap {
	return &IntMap {
		__Data: make(map[int]Object),
		__Added: 0,
	}
}

func (m *StrMap) Has(key string) bool  {
	var _, exists = m.__Data[key]
	return exists
}

func (m *IntMap) Has(key int) bool {
	var _, exists = m.__Data[key]
	return exists
}

func (m *StrMap) Get(key string) Object {
	Assert(m.Has(key), __MapNoSuchEntry)
	return m.__Data[key]
}

func (m *IntMap) Get(key int) Object {
	Assert(m.Has(key), __MapNoSuchEntry)
	return m.__Data[key]
}

func (m *StrMap) Set(key string, value Object) {
	if !m.Has(key) { m.__Added += 1 }
	m.__Data[key] = value
}

func (m *IntMap) Set(key int, value Object) {
	if !m.Has(key) { m.__Added += 1 }
	m.__Data[key] = value
}

func (m *StrMap) Delete(key string) {
	Assert(m.Has(key), __MapNoSuchEntry)
	delete(m.__Data, key)
	if len(m.__Data) < m.__Added / MapShrinkFactor {
		var old_data = &(m.__Data)
		var new_data = make(map[string]Object)
		for k, v := range *old_data {
			new_data[k] = v
		}
		*old_data = new_data
		m.__Added = len(new_data)
	}
}

func (m *IntMap) Delete(key int) {
	Assert(m.Has(key), __MapNoSuchEntry)
	delete(m.__Data, key)
	if len(m.__Data) < m.__Added / MapShrinkFactor {
		var old_data = &(m.__Data)
		var new_data = make(map[int]Object)
		for k, v := range *old_data {
			new_data[k] = v
		}
		*old_data = new_data
		m.__Added = len(new_data)
	}
}

func (m *StrMap) Size() int {
	return len(m.__Data)
}

func (m *IntMap) Size() int {
	return len(m.__Data)
}
