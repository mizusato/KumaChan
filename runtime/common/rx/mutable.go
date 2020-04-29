package rx

import (
	"reflect"
	"sync"
)


type Source interface {
	Listen() Effect
}


type Sink struct {
	mutex      *sync.RWMutex
	nextId     uint64
	listeners  map[uint64] Listener
}
type Listener struct {
	Notify  func(Object)
}
func CreateSink() *Sink {
	var mutex sync.RWMutex
	return &Sink {
		mutex:     &mutex,
		nextId:    0,
		listeners: make(map[uint64] Listener, 0),
	}
}

func (s *Sink) Listen() Effect {
	return CreateEffect(func(sender Sender) {
		var l = s.addListener(Listener {
			Notify: func(value Object) {
				sender.Next(value)
			},
		})
		<- sender.Context().Done()
		s.removeListener(l)
	})
}

func (s *Sink) Emit(value Object) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for _, l := range s.listeners {
		l.Notify(value)
	}
}

func (s *Sink) addListener(l Listener) uint64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var id = s.nextId
	s.listeners[id] = l
	s.nextId = (id + 1)
	return id
}

func (s *Sink) removeListener(id uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var _, exists = s.listeners[id]
	if !exists { panic("cannot remove absent listener") }
	delete(s.listeners, id)
}


type Cell struct {
	value   Object
}
func CreateCell(init_val Object) *Cell {
	return &Cell { init_val }
}

func (c *Cell) Get() Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		return c.value, true
	})
}

func (c *Cell) Set(new_val Object) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		c.value = new_val
		return nil, true
	})
}

func (c *Cell) Map(f func(Object)Object) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		c.value = f(c.value)
		return nil, true
	})
}


type List struct {
	slice_rv reflect.Value
}
func CreateList(l interface{}) List {
	var rv = reflect.ValueOf(l)
	if rv.Kind() != reflect.Slice { panic("given value is not a slice") }
	return List { rv }
}

func (l List) ToRaw() Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		var copied_rv = reflect.MakeSlice(l.slice_rv.Type(), l.slice_rv.Len(), l.slice_rv.Len())
		reflect.Copy(copied_rv, l.slice_rv)
		return copied_rv.Interface(), true
	})
}

func (l List) Length() Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		return uint(l.slice_rv.Len()), true
	})
}

func (l List) Get(index uint) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		var val = l.slice_rv.Index(int(index)).Interface()
		return val, true
	})
}

func (l List) Set(index uint, val Object) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		l.slice_rv.Index(int(index)).Set(reflect.ValueOf(val))
		return nil, true
	})
}

func (l List) Shift() Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		var length = l.slice_rv.Len()
		if length > 0 {
			var head = l.slice_rv.Index(0)
			l.slice_rv = l.slice_rv.Slice(1, length)
			return struct { Object; bool } { head, true }, true
		} else {
			return struct { Object; bool } { nil, false }, true
		}
	})
}

func (l List) Pop() Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		var length = l.slice_rv.Len()
		if length > 0 {
			var tail = l.slice_rv.Index(length-1)
			l.slice_rv = l.slice_rv.Slice(0, length-1)
			return struct { Object; bool } { tail, true }, true
		} else {
			return struct { Object; bool } { nil, false }, true
		}
	})
}

func (l List) Push(new_tail Object) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		l.slice_rv = reflect.Append(l.slice_rv, reflect.ValueOf(new_tail))
		return nil, true
	})
}


type StringHashMap  map[string] Object
func CreateStringHashMap(m map[string] Object) StringHashMap {
	return StringHashMap(m)
}

func (m StringHashMap) Has(key string) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		var _, exists = m[key]
		return exists, true
	})
}

func (m StringHashMap) Get(key string) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		var val, exists = m[key]
		if exists {
			return struct { Object; bool } { val, true }, true
		} else {
			return struct { Object; bool } { nil, false }, true
		}
	})
}

func (m StringHashMap) Set(key string, val Object) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		m[key] = val
		return nil, true
	})
}

func (m StringHashMap) Delete(key string) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		var deleted, exists = m[key]
		if exists {
			delete(m, key)
			return struct { Object; bool } { deleted, true }, true
		} else {
			return struct { Object; bool } { nil, false }, true
		}
	})
}

type NumberHashMap  map[uint] Object
func CreateNumberHashMap(m map[uint] Object) NumberHashMap {
	return NumberHashMap(m)
}

func (m NumberHashMap) Has(key uint) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		var _, exists = m[key]
		return exists, true
	})
}

func (m NumberHashMap) Get(key uint) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		var val, exists = m[key]
		if exists {
			return struct { Object; bool } { val, true }, true
		} else {
			return struct { Object; bool } { nil, false }, true
		}
	})
}

func (m NumberHashMap) Set(key uint, val Object) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		m[key] = val
		return nil, true
	})
}

func (m NumberHashMap) Delete(key uint) Effect {
	return CreateBlockingEffect(func()(Object, bool) {
		var deleted, exists = m[key]
		if exists {
			delete(m, key)
			return struct { Object; bool } { deleted, true }, true
		} else {
			return struct { Object; bool } { nil, false }, true
		}
	})
}


type Buffer struct {
	data  *([] byte)
}
func CreateBuffer(capacity uint) Buffer {
	var data = make([] byte, 0, capacity)
	return Buffer { &data }
}

func (buf Buffer) Write(bytes ([] byte)) Effect {
	return CreateBlockingEffect(func() (Object, bool) {
		*buf.data = append(*buf.data, bytes...)
		return nil, true
	})
}

func (buf Buffer) Dump() Effect {
	return CreateEffect(func(sender Sender) {
		var dumped = make([] byte, len(*buf.data))
		copy(dumped, *buf.data)
		sender.Next(dumped)
		sender.Complete()
	})
}
