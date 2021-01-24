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
	Emit(obj Object) Action
}

// Bus accepts and provides values
type Bus interface {
	Sink
	Watch() Action
}

// Reactive accepts and provides values, while holding a current value
type Reactive interface {
	Bus
	Update(f func(old_state Object) Object, k *KeyChain) Action
	Project(k *KeyChain) Action
	Snapshot() Action
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
		base:    sink,
		adapter: adapter,
	}
}

func ReactiveAdapt(r Reactive, in (func(Object) (func(Object) Object))) Sink {
	return &AdaptedReactive {
		base: r,
		in:   in,
	}
}

func ReactiveMorph (
	r    Reactive,
	in   (func(Object) (func(Object) Object)),
	out  (func(Object) Object),
) Reactive {
	return &MorphedReactive {
		AdaptedReactive: &AdaptedReactive {
			base: r,
			in:   in,
		},
		out: out,
	}
}

func ReactiveProject (
	r    Reactive,
	in   (func(Object) (func(Object) Object)),
	out  (func(Object) Object),
	key  *KeyChain,
) Reactive {
	return &ProjectedReactive {
		base: r,
		in:   in,
		out:  out,
		key:  key,
	}
}

func ReactiveBranch (
	r Reactive,
	in  (func(Object) Object),
	out (func(Object) (Object,bool)),
) Reactive {
	return &FilterMappedReactive {
		AdaptedReactive: &AdaptedReactive {
			base: r,
			in: func(_ Object) func(Object) Object {
				return in
			},
		},
		out: out,
	}
}


// Transformation API Implementations

type AdaptedSink struct {
	base     Sink
	adapter  func(Object) Object
}
func (a *AdaptedSink) Emit(obj Object) Action {
	return a.base.Emit(a.adapter(obj))
}

type AdaptedReactive struct {
	base  Reactive
	in    func(Object) (func(Object) Object)
}
func (a *AdaptedReactive) Emit(obj Object) Action {
	return a.base.Update(func(old_state Object) Object {
		return a.in(old_state)(obj)
	}, nil)
}

type MorphedReactive struct {
	*AdaptedReactive
	out  func(Object) Object
}
func (m *MorphedReactive) Watch() Action {
	return m.base.Watch().Map(m.out)
}
func (m *MorphedReactive) Update(f (func(Object) Object), key_chain *KeyChain) Action {
	return m.base.Update(func(obj Object) Object {
		return m.in(obj)(f(m.out(obj)))
	}, key_chain)
}
func (m *MorphedReactive) Project(key_chain *KeyChain) Action {
	return m.base.Project(key_chain).Map(m.out)
}
func (m *MorphedReactive) Snapshot() Action {
	return m.base.Snapshot()
}

type ProjectedReactive struct {
	base  Reactive
	in    (func(Object) (func(Object) Object))
	out   (func(Object) Object)
	key   *KeyChain
}
func (p *ProjectedReactive) ChainedKey(key *KeyChain) *KeyChain {
	if p.key == nil && key == nil {
		return nil
	} else if p.key != nil && key == nil {
		return p.key
	} else {
		return &KeyChain {
			Parent: p.key,
			Key:    key,
		}
	}
}
func (p *ProjectedReactive) Watch() Action {
	return p.base.Project(p.key).Map(p.out)
}
func (p *ProjectedReactive) Emit(obj Object) Action {
	return p.base.Update(func(old_state Object) Object {
		return p.in(old_state)(obj)
	}, p.key)
}
func (p *ProjectedReactive) Update(f (func(Object) Object), key *KeyChain) Action {
	return p.base.Update(func(obj Object) Object {
		return p.in(obj)(f(p.out(obj)))
	}, p.ChainedKey(key))
}
func (p *ProjectedReactive) Project(key *KeyChain) Action {
	return p.base.Project(p.ChainedKey(key)).Map(p.out)
}
func (p *ProjectedReactive) Snapshot() Action {
	return p.base.Snapshot()
}

type FilterMappedReactive struct {
	*AdaptedReactive
	out  func(Object) (Object, bool)
}
func (m *FilterMappedReactive) Watch() Action {
	return m.base.Watch().FilterMap(m.out)
}
func (m *FilterMappedReactive) Update(f (func(Object) Object), key_chain *KeyChain) Action {
	return m.base.Update(func(current Object) Object {
		var current_out, ok = m.out(current)
		if ok {
			return m.in(current)(f(current_out))
		} else {
			panic("FilterMappedReactive: invalid update operation")
		}
	}, key_chain)
}
func (m *FilterMappedReactive) Project(key_chain *KeyChain) Action {
	return m.base.Project(key_chain).FilterMap(m.out)
}
func (m *FilterMappedReactive) Snapshot() Action {
	return m.base.Snapshot()
}

type AutoSnapshotReactive struct {
	Entity  ReactiveEntity
}
func (a AutoSnapshotReactive) Watch() Action {
	return a.Entity.Watch()
}
func (a AutoSnapshotReactive) Emit(obj Object) Action {
	return a.Entity.Snapshot().Then(func(_ Object) Action {
		return a.Entity.Emit(obj)
	})
}
func (a AutoSnapshotReactive) Update(f func(Object)(Object), key_chain *KeyChain) Action {
	return a.Entity.Snapshot().Then(func(_ Object) Action {
		return a.Entity.Update(f, key_chain)
	})
}
func (a AutoSnapshotReactive) Project(key_chain *KeyChain) Action {
	return a.Entity.Project(key_chain)
}
func (_ AutoSnapshotReactive) Snapshot() Action {
	panic("suspicious snapshot operation on a auto-snapshot reactive")
}


// Trivial Sink: Callback

type Callback  func(Object) Action
func (cb Callback) Emit(obj Object) Action {
	return cb(obj)
}


// Basic Implementations of Bus[T] and Reactive[T]

type BusImpl struct {
	next_id    uint64
	listeners  [] Listener       // first in, first notified
	index      map[uint64] uint  // id --> position in listeners
}
type Listener struct {
	Notify  func(Object)
}
func CreateBus() *BusImpl {
	return &BusImpl {
		next_id:   0,
		listeners: make([] Listener, 0),
		index:     make(map[uint64] uint),
	}
}
func (b *BusImpl) Watch() Action {
	return NewListener(func(next func(Object)) func() {
		var id = b.addListener(Listener {
			Notify: next,
		})
		return func() {
			b.removeListener(id)
		}
	})
}
func (b *BusImpl) Emit(obj Object) Action {
	return NewSync(func() (Object, bool) {
		b.notify(obj)
		return nil, true
	})
}
func (b *BusImpl) notify(obj Object) {
	for _, l := range b.copyListeners() {
		l.Notify(obj)
	}
}
// TODO: rename listener -> watcher
func (b *BusImpl) copyListeners() ([] Listener) {
	var the_copy = make([] Listener, len(b.listeners))
	copy(the_copy, b.listeners)
	return the_copy
}
func (b *BusImpl) addListener(l Listener) uint64 {
	var id = b.next_id
	var pos = uint(len(b.listeners))
	b.listeners = append(b.listeners, l)
	b.index[id] = pos
	b.next_id = (id + 1)
	return id
}
func (b *BusImpl) removeListener(id uint64) {
	var pos, exists = b.index[id]
	if !exists { panic("cannot remove absent listener") }
	// update index
	delete(b.index, id)
	for current, _ := range b.index {
		if current > id {
			// position left shifted
			b.index[current] -= 1
		}
	}
	// update queue
	b.listeners[pos] = Listener {}
	var L = uint(len(b.listeners))
	if !(L >= 1) { panic("something went wrong") }
	for i := pos; i < (L-1); i += 1 {
		b.listeners[i] = b.listeners[i + 1]
	}
	b.listeners[L-1] = Listener {}
	b.listeners = b.listeners[:L-1]
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
	Undo  *Stack  // Stack<ReactiveStateChange>
	Redo  *Stack  // Stack<ReactiveStateChange>
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
func (r *ReactiveImpl) Watch() Action {
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
func (r *ReactiveImpl) WatchDiff() Action {
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
func (r *ReactiveImpl) Project(k *KeyChain) Action {
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
	r.bus.notify(change)
}
func (r *ReactiveImpl) notifyDiff() {
	r.bus.notify(Pair { r.snapshots, r.last_change.Value })
}
func (r *ReactiveImpl) Emit(new_state Object) Action {
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
func (r *ReactiveImpl) Update(f (func(Object) Object), k *KeyChain) Action {
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
func (r *ReactiveImpl) Snapshot() Action {
	return NewSync(func() (Object, bool) {
		r.snapshots.Redo = nil
		r.snapshots.Undo = r.snapshots.Undo.Pushed(r.last_change)
		r.notifyDiff()
		return nil, true
	})
}
func (r *ReactiveImpl) Undo() Action {
	return NewSync(func() (Object, bool) {
		var top, rest, ok = r.snapshots.Undo.Popped()
		if ok {
			var current = r.last_change
			r.snapshots.Redo = r.snapshots.Redo.Pushed(current)
			r.snapshots.Undo = rest
			var change = top.(ReactiveStateChange)
			r.commit(change)
			r.notifyDiff()
			return nil, true
		} else {
			return nil, true
		}
	})
}
func (r *ReactiveImpl) Redo() Action {
	return NewSync(func() (Object, bool) {
		var top, rest, ok = r.snapshots.Redo.Popped()
		if ok {
			var current = r.last_change
			r.snapshots.Undo = r.snapshots.Undo.Pushed(current)
			r.snapshots.Redo = rest
			var change = top.(ReactiveStateChange)
			r.commit(change)
			r.notifyDiff()
			return nil, true
		} else {
			return nil, true
		}
	})
}

