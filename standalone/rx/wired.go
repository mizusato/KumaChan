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
	Emit(obj Object) Observable
}

// Bus accepts and provides values
type Bus interface {
	Sink
	Watch() Observable
}
func DistinctWatch(b Bus, eq func(Object,Object)(bool)) Observable {
	return b.Watch().DistinctUntilChanged(eq)
}

// Reactive accepts and provides values, while holding a current value
type Reactive interface {
	Bus
	Read() Observable
	Update(f func(old_state Object)(Object)) Observable
}

// ReactiveEntity is a Reactive that is NOT derived from another Reactive
type ReactiveEntity = *ReactiveImpl


// Operators

func Connect(source Observable, sink Sink) Observable {
	return source.ConcatMap(func(value Object) Observable {
		return sink.Emit(value)
	}).WaitComplete()
}

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


// Operators Implementations

type AdaptedSink struct {
	base     Sink
	adapter  func(Object) Object
}
func (a *AdaptedSink) Emit(obj Object) Observable {
	return a.base.Emit(a.adapter(obj))
}

type AdaptedReactive struct {
	base  Reactive
	in    func(Object) (func(Object) Object)
}
func (a *AdaptedReactive) Emit(obj Object) Observable {
	return a.base.Update(func(old_state Object) Object {
		return a.in(old_state)(obj)
	})
}

type MorphedReactive struct {
	*AdaptedReactive
	out  func(Object) Object
}
func (m *MorphedReactive) Watch() Observable {
	return m.base.Watch().Map(m.out)
}
func (m *MorphedReactive) Read() Observable {
	return m.base.Read().Map(m.out)
}
func (m *MorphedReactive) Update(f (func(Object) Object)) Observable {
	return m.base.Update(func(obj Object) Object {
		return m.in(obj)(f(m.out(obj)))
	})
}

type FilterMappedReactive struct {
	*AdaptedReactive
	out  func(Object) (Object, bool)
}
func (m *FilterMappedReactive) Watch() Observable {
	return m.base.Watch().FilterMap(m.out)
}
func (m *FilterMappedReactive) Read() Observable {
	return m.base.Read().Map(func(current Object) Object {
		var current_out, ok = m.out(current)
		if ok {
			return current_out
		} else {
			panic("FilterMappedReactive: invalid read operation")
		}
	})
}
func (m *FilterMappedReactive) Update(f (func(Object) Object)) Observable {
	return m.base.Update(func(current Object) Object {
		var current_out, ok = m.out(current)
		if ok {
			return m.in(current)(f(current_out))
		} else {
			panic("FilterMappedReactive: invalid update operation")
		}
	})
}


// Trivial Sink: BlackHole and Callback

type BlackHole struct{}
func (_ BlackHole) Emit(_ Object) Observable {
	return NewConstant(nil)
}

type Callback  func(Object) Observable
func (cb Callback) Emit(obj Object) Observable {
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
func (b *BusImpl) Watch() Observable {
	return NewSubscription(func(next func(Object)) func() {
		var id = b.addWatcher(Watcher {
			Notify: next,
		})
		return func() {
			b.removeWatcher(id)
		}
	})
}
func (b *BusImpl) Emit(obj Object) Observable {
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
	bus    *BusImpl
	value  Object
}
func CreateReactive(init Object) ReactiveEntity {
	return &ReactiveImpl {
		bus:   CreateBus(),
		value: init,
	}
}
func (r *ReactiveImpl) commit(new_value Object) {
	r.value = new_value
	r.bus.notify(new_value)
}
func (r *ReactiveImpl) Watch() Observable {
	return NewSubscription(func(next func(Object)) func() {
		next(r.value)
		var w = r.bus.addWatcher(Watcher {
			Notify: func(obj Object) {
				next(obj)
			},
		})
		return func() {
			r.bus.removeWatcher(w)
		}
	})
}
func (r *ReactiveImpl) Read() Observable {
	return NewSync(func() (Object, bool) {
		return r.value, true
	})
}
func (r *ReactiveImpl) Emit(new_value Object) Observable {
	return NewSync(func() (Object, bool) {
		r.commit(new_value)
		return nil, true
	})
}
func (r *ReactiveImpl) Update(f (func(Object) Object)) Observable {
	return NewSync(func() (Object, bool) {
		var old_value = r.value
		var new_value = f(old_value)
		r.commit(new_value)
		return nil, true
	})
}

