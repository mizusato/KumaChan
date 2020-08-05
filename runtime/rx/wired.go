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
func (b *Bus) Receive() Effect {
	return CreateBlockingListenerEffect(func(next func(Object)) func() {
		var l = b.addListener(Listener {
			Notify: next,
		})
		return func() {
			b.removeListener(l)
		}
	})
}
func (b *Bus) Send(obj Object) Effect {
	return CreateBlockingEffect(func() (Object, bool) {
		for _, l := range b.listeners {
			l.Notify(obj)
		}
		return nil, true
	})
}
func (b *Bus) addListener(l Listener) uint64 {
	var id = b.nextId
	b.listeners[id] = l
	b.nextId = (id + 1)
	return id
}
func (b *Bus) removeListener(id uint64) {
	var _, exists = b.listeners[id]
	if !exists { panic("cannot remove absent listener") }
	delete(b.listeners, id)
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
func (l *Latch) Receive() Effect {
	return l.bus.Receive().StartWith(l.state)
}
func (l *Latch) Send(obj Object) Effect {
	return CreateBlockingEffect(func() (Object, bool) {
		l.state = obj
		for _, l := range l.bus.listeners {
			l.Notify(obj)
		}
		return nil, true
	})
}
func (l *Latch) Reset() Effect {
	return l.Send(l.init)
}
