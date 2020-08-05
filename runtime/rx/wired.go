package rx


type Source interface {
	Receive() Effect
}

type Sink interface {
	Send(Object) Effect
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
	return CreateBlockingListenerEffect(func(next func(Object)) func() {
		var l = bus.addListener(Listener {
			Notify: next,
		})
		return func() {
			bus.removeListener(l)
		}
	})
}
func (bus *Bus) Send(obj Object) Effect {
	return CreateBlockingEffect(func() (Object, bool) {
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
	return CreateBlockingListenerEffect(func(next func(Object)) func() {
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
	return CreateBlockingEffect(func() (Object, bool) {
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