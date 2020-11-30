package rx


/**
 *  FRP with wired components
 *    Sink <: Bus(aka Subject) <: Reactive(aka BehaviorSubject)
 *    Sink[A]     --> adapt[B->A]         --> Sink[B]
 *    Reactive[A] --> adapt[A->B->A]      --> Sink[B]
 *    Reactive[A] --> morph[A->B->A,A->B] --> Reactive[B]
 */

// Sink accepts values
type Sink interface {
	Emit(obj Object) Effect
}

// Bus accepts and provides values
type Bus interface {
	Sink
	Watch() Effect
}

// Reactive accepts and provides values, while holding a current value
type Reactive interface {
	Bus
	Update(f func(old_state Object) Object) Effect
}

type AdaptedSink struct {
	Sink     Sink
	Adapter  func(Object) Object
}
func SinkAdapt(sink Sink, adapter (func(Object) Object)) Sink {
	return &AdaptedSink {
		Sink:    sink,
		Adapter: adapter,
	}
}
func (a *AdaptedSink) Emit(obj Object) Effect {
	return a.Sink.Emit(a.Adapter(obj))
}

type AdaptedReactive struct {
	Reactive  Reactive
	In        func(Object) (func(Object) Object)
}
func ReactiveAdapt(r Reactive, in (func(Object) (func(Object) Object))) Sink {
	return &AdaptedReactive {
		Reactive: r,
		In:       in,
	}
}
func (a *AdaptedReactive) Emit(obj Object) Effect {
	return a.Reactive.Update(func(old_state Object) Object {
		return a.In(old_state)(obj)
	})
}

type MorphedReactive struct {
	*AdaptedReactive
	Out  func(Object) Object
}
func ReactiveMorph(r Reactive, in (func(Object) (func(Object) Object)), out (func(Object) Object)) Reactive {
	return &MorphedReactive {
		AdaptedReactive: &AdaptedReactive {
			Reactive: r,
			In:       in,
		},
		Out: out,
	}
}
func (m *MorphedReactive) Watch() Effect {
	return m.Reactive.Watch().Map(m.Out)
}
func (m *MorphedReactive) Update(f (func(Object) Object)) Effect {
	return m.Reactive.Update(func(obj Object) Object {
		return m.In(f(m.Out(obj)))
	})
}


type BusImpl struct {
	nextId     uint64
	listeners  map[uint64] Listener
}
type Listener struct {
	Notify  func(Object)
}
func CreateBus() *BusImpl {
	return &BusImpl{
		nextId:    0,
		listeners: make(map[uint64] Listener, 0),
	}
}
func (b *BusImpl) Watch() Effect {
	return NewListener(func(next func(Object)) func() {
		var l = b.addListener(Listener {
			Notify: next,
		})
		return func() {
			b.removeListener(l)
		}
	})
}
func (b *BusImpl) Emit(obj Object) Effect {
	return NewSync(func() (Object, bool) {
		for _, l := range b.listeners {
			l.Notify(obj)
		}
		return nil, true
	})
}
func (b *BusImpl) addListener(l Listener) uint64 {
	var id = b.nextId
	b.listeners[id] = l
	b.nextId = (id + 1)
	return id
}
func (b *BusImpl) removeListener(id uint64) {
	var _, exists = b.listeners[id]
	if !exists { panic("cannot remove absent listener") }
	delete(b.listeners, id)
}

type ReactiveImpl struct {
	bus    *BusImpl
	state  Object
}
func CreateReactive(init Object) *ReactiveImpl {
	return &ReactiveImpl {
		bus:   CreateBus(),
		state: init,
	}
}
func (r *ReactiveImpl) Watch() Effect {
	return NewListener(func(next func(Object)) func() {
		next(r.state)
		var l = r.bus.addListener(Listener {
			Notify: next,
		})
		return func() {
			r.bus.removeListener(l)
		}
	})
}
func (r *ReactiveImpl) Emit(obj Object) Effect {
	return NewSync(func() (Object, bool) {
		r.state = obj
		for _, l := range r.bus.listeners {
			l.Notify(obj)
		}
		return nil, true
	})
}
func (r *ReactiveImpl) Update(f (func(Object) Object)) Effect {
	return NewSync(func() (Object, bool) {
		var old_state = r.state
		var new_state = f(old_state)
		r.state = new_state
		for _, l := range r.bus.listeners {
			l.Notify(new_state)
		}
		return nil, true
	})
}

