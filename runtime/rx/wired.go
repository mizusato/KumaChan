package rx


/**
 *  FRP with wired components
 *    Sink <: Bus(aka Subject) <: Reactive(aka BehaviorSubject)
 *    transformations:
 *    Sink[A]     --> adapt[B->A]         --> Sink[B]
 *    Reactive[A] --> adapt[A->B->A]      --> Sink[B]
 *    Reactive[A] --> morph[A->B->A,A->B] --> Reactive[B]
 *    Reactive[(A,B)] --> project[A] --> Reactive[A]
 *    Reactive[(A,B)] --> project[B] --> Reactive[B]
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
	Update(f func(old_state Object) Object, k *KeyChain) Effect
	Project(k *KeyChain) Effect
}
type KeyChain struct {
	Parent  *KeyChain
	Key     Object  // should be comparable by ==
}
func (k *KeyChain) Includes(another *KeyChain) bool {
	if KeyChainEqual(k, another) {
		return true
	} else {
		if k != nil {
			return k.Parent.Includes(another)
		} else {
			return false
		}
	}
}
func KeyChainEqual(a *KeyChain, b *KeyChain) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	} else {
		if a.Key == b.Key && a.Parent == b.Parent {
			return true
		} else {
			return false
		}
	}
}


// Transformation APIs

func SinkAdapt(sink Sink, adapter (func(Object) Object)) Sink {
	return &AdaptedSink {
		Sink:    sink,
		Adapter: adapter,
	}
}

func ReactiveAdapt(r Reactive, in (func(Object) (func(Object) Object))) Sink {
	return &AdaptedReactive {
		Reactive: r,
		In:       in,
	}
}

func ReactiveMorph (
	r    Reactive,
	in   (func(Object) (func(Object) Object)),
	out  (func(Object) Object),
) Reactive {
	return &MorphedReactive {
		AdaptedReactive: &AdaptedReactive {
			Reactive: r,
			In:       in,
		},
		Out: out,
	}
}

func ReactiveProject (
	r    Reactive,
	in   (func(Object) (func(Object) Object)),
	out  (func(Object) Object),
	key  *KeyChain,
) Reactive {
	return &ProjectedReactive {
		Reactive: r,
		In:       in,
		Out:      out,
		Key:      key,
	}
}


// Transformation API Implementations

type AdaptedSink struct {
	Sink     Sink
	Adapter  func(Object) Object
}
func (a *AdaptedSink) Emit(obj Object) Effect {
	return a.Sink.Emit(a.Adapter(obj))
}

type AdaptedReactive struct {
	Reactive  Reactive
	In        func(Object) (func(Object) Object)
}
func (a *AdaptedReactive) Emit(obj Object) Effect {
	return a.Reactive.Update(func(old_state Object) Object {
		return a.In(old_state)(obj)
	}, nil)
}

type MorphedReactive struct {
	*AdaptedReactive
	Out  func(Object) Object
}
func (m *MorphedReactive) Watch() Effect {
	return m.Reactive.Watch().Map(m.Out)
}
func (m *MorphedReactive) Update(f (func(Object) Object), key_chain *KeyChain) Effect {
	return m.Reactive.Update(func(obj Object) Object {
		return m.In(obj)(f(m.Out(obj)))
	}, key_chain)
}
func (m *MorphedReactive) Project(key_chain *KeyChain) Effect {
	return m.Reactive.Project(key_chain).Map(m.Out)
}

type ProjectedReactive struct {
	Reactive  Reactive
	In        (func(Object) (func(Object) Object))
	Out       (func(Object) Object)
	Key       *KeyChain
}
func (p *ProjectedReactive) ChainedKey(key *KeyChain) *KeyChain {
	return &KeyChain {
		Parent: p.Key,
		Key:    key,
	}
}
func (p *ProjectedReactive) Watch() Effect {
	return p.Reactive.Project(p.Key).Map(p.Out)
}
func (p *ProjectedReactive) Emit(obj Object) Effect {
	return p.Reactive.Update(func(old_state Object) Object {
		return p.In(old_state)(obj)
	}, p.Key)
}
func (p *ProjectedReactive) Update(f (func(Object) Object), key *KeyChain) Effect {
	return p.Reactive.Update(func(obj Object) Object {
		return p.In(obj)(f(p.Out(obj)))
	}, p.ChainedKey(key))
}
func (p *ProjectedReactive) Project(key *KeyChain) Effect {
	return p.Reactive.Project(p.ChainedKey(key)).Map(p.Out)
}


// Trivial Sink: Callback

type Callback  func(Object) Effect
func (cb Callback) Emit(obj Object) Effect {
	return cb(obj)
}


// Basic Implementations of Bus[T] and Reactive[T]

type BusImpl struct {
	nextId     uint64
	listeners  map[uint64] Listener
}
type Listener struct {
	Notify  func(Object)
}
func CreateBus() *BusImpl {
	return &BusImpl {
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
	state  ReactiveState
}
type ReactiveState struct {
	Value     Object
	KeyChain  *KeyChain
}
func CreateReactive(init Object) *ReactiveImpl {
	return &ReactiveImpl {
		bus:   CreateBus(),
		state: ReactiveState { Value: init },
	}
}
func (r *ReactiveImpl) Watch() Effect {
	return NewListener(func(next func(Object)) func() {
		next(r.state.Value)
		var l = r.bus.addListener(Listener {
			Notify: func(state Object) {
				next(state.(ReactiveState).Value)
			},
		})
		return func() {
			r.bus.removeListener(l)
		}
	})
}
func (r *ReactiveImpl) Emit(obj Object) Effect {
	return NewSync(func() (Object, bool) {
		var new_state = ReactiveState { Value: obj }
		r.state = new_state
		for _, l := range r.bus.listeners {
			l.Notify(new_state)
		}
		return nil, true
	})
}
func (r *ReactiveImpl) Update(f (func(Object) Object), k *KeyChain) Effect {
	return NewSync(func() (Object, bool) {
		var old_state_val = r.state.Value
		var new_state_val = f(old_state_val)
		var new_state = ReactiveState {
			Value:    new_state_val,
			KeyChain: k,
		}
		r.state = new_state
		for _, l := range r.bus.listeners {
			l.Notify(new_state)
		}
		return nil, true
	})
}
func (r *ReactiveImpl) Project(k *KeyChain) Effect {
	return NewListener(func(next func(Object)) func() {
		next(r.state.Value)
		var l = r.bus.addListener(Listener {
			Notify: func(state_ Object) {
				var state = state_.(ReactiveState)
				if k.Includes(state.KeyChain) {
					next(state.Value)
				}
			},
		})
		return func() {
			r.bus.removeListener(l)
		}
	})
}

