package rx


/**
 *  FRP with wired components
 *    Sink :> Bus(aka Subject) >: Reactive(aka BehaviorSubject)
 *    transformations:
 *    Sink[A]     --> adapt[B->A]         --> Sink[B]
 *    Reactive[A] --> adapt[A->B->A]      --> Sink[B]
 *    Reactive[A] --> morph[A->B->A,A->B] --> Reactive[B]
 *    Reactive[(A,B)] --> project[A] --> Reactive[A]
 *    Reactive[(A,B)] --> project[B] --> Reactive[B]
 *    Reactive[A|B]   --> branch[A]  --> Reactive[A] (restricted update operation)
 *    Reactive[A|B]   --> branch[B]  --> Reactive[B] (restricted update operation)
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
	Snapshot() Effect
}

// ReactiveEntity is a Reactive that is NOT derived from another Reactive
type ReactiveEntity = *ReactiveImpl


// Projection Key Chain

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

func ReactiveBranch (
	r Reactive,
	in  (func(Object) Object),
	out (func(Object) (Object,bool)),
) Reactive {
	return &FilterMappedReactive {
		AdaptedReactive: &AdaptedReactive {
			Reactive: r,
			In: func(_ Object) func(Object) Object {
				return in
			},
		},
		Out: out,
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
func (m *MorphedReactive) Snapshot() Effect {
	return m.Reactive.Snapshot()
}

type ProjectedReactive struct {
	Reactive  Reactive
	In        (func(Object) (func(Object) Object))
	Out       (func(Object) Object)
	Key       *KeyChain
}
func (p *ProjectedReactive) ChainedKey(key *KeyChain) *KeyChain {
	if p.Key == nil && key == nil {
		return nil
	} else if p.Key != nil && key == nil {
		return p.Key
	} else {
		return &KeyChain {
			Parent: p.Key,
			Key:    key,
		}
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
func (p *ProjectedReactive) Snapshot() Effect {
	return p.Reactive.Snapshot()
}

type FilterMappedReactive struct {
	*AdaptedReactive
	Out  func(Object) (Object, bool)
}
func (m *FilterMappedReactive) Watch() Effect {
	return m.Reactive.Watch().FilterMap(m.Out)
}
func (m *FilterMappedReactive) Update(f (func(Object) Object), key_chain *KeyChain) Effect {
	return m.Reactive.Update(func(current Object) Object {
		var current_out, ok = m.Out(current)
		if ok {
			return m.In(current)(f(current_out))
		} else {
			panic("FilterMappedReactive: invalid update operation")
		}
	}, key_chain)
}
func (m *FilterMappedReactive) Project(key_chain *KeyChain) Effect {
	return m.Reactive.Project(key_chain).FilterMap(m.Out)
}
func (m *FilterMappedReactive) Snapshot() Effect {
	return m.Reactive.Snapshot()
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
	bus          *BusImpl  // Bus<ReactiveStateChange|Pair<ReactiveSnapshots,Object>>
	last_change  ReactiveStateChange
	snapshots    ReactiveSnapshots
}
type ReactiveStateChange struct {
	Value     Object
	KeyChain  *KeyChain
}
type ReactiveSnapshots struct {
	Undo  *Stack // Stack<ReactiveStateChange>
	Redo  *Stack // Stack<ReactiveStateChange>
}
func CreateReactive(init Object) *ReactiveImpl {
	return &ReactiveImpl {
		bus:         CreateBus(),
		last_change: ReactiveStateChange {
			Value:    init,
			KeyChain: nil,
		},
	}
}
func (r *ReactiveImpl) Watch() Effect {
	return NewListener(func(next func(Object)) func() {
		next(r.last_change.Value)
		var l = r.bus.addListener(Listener {
			Notify: func(obj Object) {
				var change, is_change = obj.(ReactiveStateChange)
				if is_change {
					next(change.Value)
				}
			},
		})
		return func() {
			r.bus.removeListener(l)
		}
	})
}
func (r *ReactiveImpl) WatchDiff() Effect {
	return NewListener(func(next func(Object)) func() {
		next(Pair { r.snapshots, r.last_change.Value})
		var l = r.bus.addListener(Listener {
			Notify: func(obj Object) {
				var pair, is_pair = obj.(Pair)
				if is_pair {
					next(pair)
				}
			},
		})
		return func() {
			r.bus.removeListener(l)
		}
	})
}
func (r *ReactiveImpl) Project(k *KeyChain) Effect {
	return NewListener(func(next func(Object)) func() {
		next(r.last_change.Value)
		var l = r.bus.addListener(Listener {
			Notify: func(obj Object) {
				var change, is_change = obj.(ReactiveStateChange)
				if is_change && k.Includes(change.KeyChain) {
					next(change.Value)
				}
			},
		})
		return func() {
			r.bus.removeListener(l)
		}
	})
}
func (r *ReactiveImpl) commit(change ReactiveStateChange) {
	r.last_change = change
	for _, l := range r.bus.listeners {
		l.Notify(change)
	}
}
func (r *ReactiveImpl) notifyDiff() {
	for _, l := range r.bus.listeners {
		l.Notify(Pair { r.snapshots, r.last_change.Value })
	}
}
func (r *ReactiveImpl) Emit(new_state Object) Effect {
	return NewSync(func() (Object, bool) {
		var change = ReactiveStateChange {
			Value:    new_state,
			KeyChain: nil,
		}
		r.commit(change)
		if r.snapshots.Redo != nil {
			r.snapshots.Redo = nil
		}
		r.notifyDiff()
		return nil, true
	})
}
func (r *ReactiveImpl) Update(f (func(Object) Object), k *KeyChain) Effect {
	return NewSync(func() (Object, bool) {
		var old_state = r.last_change.Value
		var new_state = f(old_state)
		var change = ReactiveStateChange {
			Value:    new_state,
			KeyChain: k,
		}
		r.commit(change)
		if r.snapshots.Redo != nil {
			r.snapshots.Redo = nil
		}
		r.notifyDiff()
		return nil, true
	})
}
func (r *ReactiveImpl) Snapshot() Effect {
	return NewSync(func() (Object, bool) {
		r.snapshots.Redo = nil
		r.snapshots.Undo = r.snapshots.Undo.Pushed(r.last_change)
		r.notifyDiff()
		return nil, true
	})
}
func (r *ReactiveImpl) Undo() Effect {
	return NewSync(func() (Object, bool) {
		var top, rest, ok = r.snapshots.Undo.Popped()
		if ok {
			var current = r.last_change
			r.snapshots.Redo = r.snapshots.Redo.Pushed(current)
			r.snapshots.Undo = rest
			var change = top.(ReactiveStateChange)
			r.commit(change)
			r.notifyDiff()
			return true, true
		} else {
			return false, true
		}
	})
}
func (r *ReactiveImpl) Redo() Effect {
	return NewSync(func() (Object, bool) {
		var top, rest, ok = r.snapshots.Redo.Popped()
		if ok {
			var current = r.last_change
			r.snapshots.Undo = r.snapshots.Undo.Pushed(current)
			r.snapshots.Redo = rest
			var change = top.(ReactiveStateChange)
			r.commit(change)
			r.notifyDiff()
			return true, true
		} else {
			return false, true
		}
	})
}

