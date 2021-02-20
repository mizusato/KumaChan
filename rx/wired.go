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
 *    Reactive[A|B]   --> branch[A]  --> Reactive[A] (restricted read/update operation)
 *    Reactive[A|B]   --> branch[B]  --> Reactive[B] (restricted read/update operation)
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
	Read() Action
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

func ReactiveDistinctView(r Reactive, eq func(Object,Object)(bool)) Reactive {
	return DistinctViewReactive {
		base:  r,
		equal: eq,
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
func (m *MorphedReactive) Read() Action {
	return m.base.Read().Map(m.out)
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
func (p *ProjectedReactive) Read() Action {
	return p.base.Read().Map(p.out)
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
func (m *FilterMappedReactive) Read() Action {
	return m.base.Read().Map(func(current Object) Object {
		var current_out, ok = m.out(current)
		if ok {
			return current_out
		} else {
			panic("FilterMappedReactive: invalid read operation")
		}
	})
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
func (a AutoSnapshotReactive) Read() Action {
	return a.Entity.Read()
}
func (a AutoSnapshotReactive) Emit(obj Object) Action {
	return a.Entity.EmitWithSnapshot(obj, true)
}
func (a AutoSnapshotReactive) Update(f func(Object)(Object), key_chain *KeyChain) Action {
	return a.Entity.UpdateWithSnapshot(f, key_chain, true)
}
func (a AutoSnapshotReactive) Project(key_chain *KeyChain) Action {
	return a.Entity.Project(key_chain)
}
func (_ AutoSnapshotReactive) Snapshot() Action {
	panic("suspicious snapshot operation on a auto-snapshot reactive")
}

type DistinctViewReactive struct {
	base   Reactive
	equal  func(Object,Object) bool
}
func (d DistinctViewReactive) Watch() Action {
	return d.base.Watch().DistinctUntilChanged(d.equal)
}
func (d DistinctViewReactive) Project(key_chain *KeyChain) Action {
	return d.base.Project(key_chain).DistinctUntilChanged(d.equal)
}
func (d DistinctViewReactive) Read() Action {
	return d.base.Read()
}
func (d DistinctViewReactive) Emit(obj Object) Action {
	return d.base.Emit(obj)
}
func (d DistinctViewReactive) Update(f func(Object)(Object), key_chain *KeyChain) Action {
	return d.base.Update(f, key_chain)
}
func (d DistinctViewReactive) Snapshot() Action {
	return d.base.Snapshot()
}


// Trivial Sink: Callback

type Callback  func(Object) Action
func (cb Callback) Emit(obj Object) Action {
	return cb(obj)
}


// Basic Implementations of Bus[T] and Reactive[T]

type BusImpl struct {
	next_id   uint64
	watchers  [] Watcher       // first in, first notified
	index     map[uint64] uint // id --> position in watchers
}
type Watcher struct {
	Notify  func(Object)
}
func CreateBus() *BusImpl {
	return &BusImpl {
		next_id:  0,
		watchers: make([] Watcher, 0),
		index:    make(map[uint64] uint),
	}
}
func (b *BusImpl) Watch() Action {
	return NewSubscription(func(next func(Object)) func() {
		var id = b.addWatcher(Watcher {
			Notify: next,
		})
		return func() {
			b.removeWatcher(id)
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
	for _, w := range b.copyWatcher() {
		w.Notify(obj)
	}
}
func (b *BusImpl) copyWatcher() ([] Watcher) {
	var the_copy = make([] Watcher, len(b.watchers))
	copy(the_copy, b.watchers)
	return the_copy
}
func (b *BusImpl) addWatcher(w Watcher) uint64 {
	var id = b.next_id
	var pos = uint(len(b.watchers))
	b.watchers = append(b.watchers, w)
	b.index[id] = pos
	b.next_id = (id + 1)
	return id
}
func (b *BusImpl) removeWatcher(id uint64) {
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
	b.watchers[pos] = Watcher {}
	var L = uint(len(b.watchers))
	if !(L >= 1) { panic("something went wrong") }
	for i := pos; i < (L-1); i += 1 {
		b.watchers[i] = b.watchers[i + 1]
	}
	b.watchers[L-1] = Watcher {}
	b.watchers = b.watchers[:L-1]
}

type ReactiveImpl struct {
	bus          *BusImpl  // Bus<ReactiveStateChange|Pair<ReactiveSnapshots,Object>>
	last_change  ReactiveStateChange
	snapshots    ReactiveSnapshots
	equal        func(Object,Object) bool
}
type ReactiveStateChange struct {
	Value     Object
	KeyChain  *KeyChain
}
type ReactiveSnapshots struct {
	Undo  *Stack  // Stack<ReactiveStateChange>
	Redo  *Stack  // Stack<ReactiveStateChange>
}
func CreateReactive(init Object, equal (func(Object,Object)(bool))) *ReactiveImpl {
	return &ReactiveImpl {
		bus: CreateBus(),
		last_change: ReactiveStateChange {
			Value:    init,
			KeyChain: nil,
		},
		equal: equal,
	}
}
func (r *ReactiveImpl) Watch() Action {
	return NewSubscription(func(next func(Object)) func() {
		next(r.last_change.Value)
		var w = r.bus.addWatcher(Watcher{
			Notify: func(obj Object) {
				var change, is_change = obj.(ReactiveStateChange)
				if is_change {
					next(change.Value)
				}
			},
		})
		return func() {
			r.bus.removeWatcher(w)
		}
	})
}
func (r *ReactiveImpl) Read() Action {
	return NewSync(func() (Object, bool) {
		return r.last_change.Value, true
	})
}
func (r *ReactiveImpl) WatchDiff() Action {
	return NewSubscription(func(next func(Object)) func() {
		next(Pair { r.snapshots, r.last_change.Value })
		var w = r.bus.addWatcher(Watcher{
			Notify: func(obj Object) {
				var pair, is_pair = obj.(Pair)
				if is_pair {
					next(pair)
				}
			},
		})
		return func() {
			r.bus.removeWatcher(w)
		}
	})
}
func (r *ReactiveImpl) Project(k *KeyChain) Action {
	return NewSubscription(func(next func(Object)) func() {
		next(r.last_change.Value)
		var w = r.bus.addWatcher(Watcher {
			Notify: func(obj Object) {
				var change, is_change = obj.(ReactiveStateChange)
				if is_change && k.Includes(change.KeyChain) {
					next(change.Value)
				}
			},
		})
		return func() {
			r.bus.removeWatcher(w)
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
	return r.EmitWithSnapshot(new_state, false)
}
func (r *ReactiveImpl) EmitWithSnapshot(new_state Object, snapshot bool) Action {
	return NewSync(func() (Object, bool) {
		var old_state = r.last_change.Value
		if r.equal(new_state, old_state) {
			return nil, true
		}
		var change = ReactiveStateChange {
			Value:    new_state,
			KeyChain: nil,
		}
		if snapshot {
			r.doSnapshot()
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
	return r.UpdateWithSnapshot(f, k, false)
}
func (r *ReactiveImpl) UpdateWithSnapshot(f (func(Object) Object), k *KeyChain, snapshot bool) Action {
	return NewSync(func() (Object, bool) {
		var old_state = r.last_change.Value
		var new_state = f(old_state)  // TODO: consider optional new state
		if r.equal(old_state, new_state) {
			return nil, true
		}
		var change = ReactiveStateChange {
			Value:    new_state,
			KeyChain: k,
		}
		if snapshot {
			r.doSnapshot()
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
		r.doSnapshot()
		r.notifyDiff()
		return nil, true
	})
}
func (r *ReactiveImpl) doSnapshot() {
	var current = r.last_change.Value
	r.snapshots.Redo = nil
	r.snapshots.Undo = r.snapshots.Undo.Pushed(current)
}
func (r *ReactiveImpl) Undo() Action {
	return NewSync(func() (Object, bool) {
		var top, rest, ok = r.snapshots.Undo.Popped()
		if ok {
			var current = r.last_change.Value
			r.snapshots.Redo = r.snapshots.Redo.Pushed(current)
			r.snapshots.Undo = rest
			r.commit(ReactiveStateChange {
				Value:    top,
				KeyChain: nil,
			})
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
			var current = r.last_change.Value
			r.snapshots.Undo = r.snapshots.Undo.Pushed(current)
			r.snapshots.Redo = rest
			r.commit(ReactiveStateChange {
				Value:    top,
				KeyChain: nil,
			})
			r.notifyDiff()
			return nil, true
		} else {
			return nil, true
		}
	})
}

