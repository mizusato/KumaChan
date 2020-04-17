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
	return CreateBlockingEffect(func(next func(Object)) error {
		next(c.value)
		return nil
	})
}

func (c *Cell) Set(new_val Object) Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		c.value = new_val
		next(nil)
		return nil
	})
}

func (c *Cell) Map(f func(Object)Object) Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		c.value = f(c.value)
		next(nil)
		return nil
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
	return CreateBlockingEffect(func(next func(Object)) error {
		var copied_rv = reflect.MakeSlice(l.slice_rv.Type(), l.slice_rv.Len(), l.slice_rv.Len())
		reflect.Copy(copied_rv, l.slice_rv)
		next(copied_rv.Interface())
		return nil
	})
}

func (l List) Length() Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		next(uint(l.slice_rv.Len()))
		return nil
	})
}

func (l List) Get(index uint) Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		var val = l.slice_rv.Index(int(index)).Interface()
		next(val)
		return nil
	})
}

func (l List) Set(index uint, val Object) Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		l.slice_rv.Index(int(index)).Set(reflect.ValueOf(val))
		next(nil)
		return nil
	})
}

func (l List) Shift() Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		var length = l.slice_rv.Len()
		if length > 0 {
			var head = l.slice_rv.Index(0)
			l.slice_rv = l.slice_rv.Slice(1, length)
			next(struct { Object; bool } { head, true })
		} else {
			next(struct { Object; bool } { nil, false })
		}
		return nil
	})
}

func (l List) Pop() Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		var length = l.slice_rv.Len()
		if length > 0 {
			var tail = l.slice_rv.Index(length-1)
			l.slice_rv = l.slice_rv.Slice(0, length-1)
			next(struct { Object; bool } { tail, true })
		} else {
			next(struct { Object; bool } { nil, false })
		}
		return nil
	})
}

func (l List) Push(new_tail Object) Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		l.slice_rv = reflect.Append(l.slice_rv, reflect.ValueOf(new_tail))
		next(nil)
		return nil
	})
}


type Map  map[string] Object
func CreateMap(m map[string] Object) Map {
	return Map(m)
}

func (m Map) Has(key string) Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		var _, exists = m[key]
		next(exists)
		return nil
	})
}

func (m Map) Get(key string) Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		var val, exists = m[key]
		if exists {
			next(struct { Object; bool } { val, true })
		} else {
			next(struct { Object; bool } { nil, false })
		}
		return nil
	})
}

func (m Map) Set(key string, val Object) Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		m[key] = val
		next(nil)
		return nil
	})
}

func (m Map) Delete(key string) Effect {
	return CreateBlockingEffect(func(next func(Object)) error {
		var deleted, exists = m[key]
		if exists {
			delete(m, key)
			next(struct { Object; bool } { deleted, true })
		} else {
			next(struct { Object; bool } { nil, false })
		}
		return nil
	})
}

