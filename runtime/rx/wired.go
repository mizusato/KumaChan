package rx


type Source interface {
	Receive() Effect
}

type Sink interface {
	Send(Object) Effect
}

type MappedSource struct {
	Source  Source
	Mapper  func(Object) Object
}
func (m *MappedSource) Receive() Effect {
	return m.Source.Receive().Map(m.Mapper)
}

type AdaptedSink struct {
	Sink     Sink
	Adapter  func(Object) Object
}
func (a *AdaptedSink) Send(obj Object) Effect {
	return a.Sink.Send(a.Adapter(obj))
}

type Bus struct {
	nextId     uint64
	listeners  map[uint64] Listener
}
type Listener struct {
	Notify  func(Object)
}
func CreateBus() *Bus {
	return &Bus {
		nextId:    0,
		listeners: make(map[uint64] Listener, 0),
	}
}
func (bus *Bus) Receive() Effect {
	return NewListener(func(next func(Object)) func() {
		var l = bus.addListener(Listener {
			Notify: next,
		})
		return func() {
			bus.removeListener(l)
		}
	})
}
func (bus *Bus) Send(obj Object) Effect {
	return NewSync(func() (Object, bool) {
		for _, l := range bus.listeners {
			l.Notify(obj)
		}
		return nil, true
	})
}
func (bus *Bus) addListener(l Listener) uint64 {
	var id = bus.nextId
	bus.listeners[id] = l
	bus.nextId = (id + 1)
	return id
}
func (bus *Bus) removeListener(id uint64) {
	var _, exists = bus.listeners[id]
	if !exists { panic("cannot remove absent listener") }
	delete(bus.listeners, id)
}

type Latch struct {
	bus    *Bus
	state  Object
	init   Object
}
func CreateLatch(init Object) *Latch {
	return &Latch {
		bus:   CreateBus(),
		state: init,
		init:  init,
	}
}
func (latch *Latch) Receive() Effect {
	return NewListener(func(next func(Object)) func() {
		next(latch.state)
		var l = latch.bus.addListener(Listener {
			Notify: next,
		})
		return func() {
			latch.bus.removeListener(l)
		}
	})
}
func (latch *Latch) Send(obj Object) Effect {
	return NewSync(func() (Object, bool) {
		latch.state = obj
		for _, l := range latch.bus.listeners {
			l.Notify(obj)
		}
		return nil, true
	})
}
func (latch *Latch) Reset() Effect {
	return latch.Send(latch.init)
}

type AdaptedLatch struct {
	Latch       *Latch
	GetAdapter  func(Object) (func(Object) Object)
}
func (a *AdaptedLatch) Send(obj Object) Effect {
	return NewSync(func() (Object, bool) {
		var old_state = a.Latch.state
		var adapter = a.GetAdapter(old_state)
		var new_state = adapter(obj)
		a.Latch.state = new_state
		for _, l := range a.Latch.bus.listeners {
			l.Notify(new_state)
		}
		return nil, true
	})
}

type CombinedLatch struct {
	Elements  [] *Latch
}
func (c *CombinedLatch) Receive() Effect {
	return NewListener(func(next func(Object)) func() {
		var L = len(c.Elements)
		var values = make([] Object, L)
		for i, el := range c.Elements {
			values[i] = el.state
		}
		next(values)
		var listeners = make([] uint64, L)
		for loop_var, el := range c.Elements {
			var index = loop_var
			listeners[index] = el.bus.addListener(Listener {
				Notify: func(object Object) {
					values[index] = object
					next(values)
				},
			})
		}
		return func() {
			for i, l := range listeners {
				c.Elements[i].bus.removeListener(l)
			}
		}
	})
}

