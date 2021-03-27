package api

import (
	"os"
	"fmt"
	"reflect"
	"math/rand"
	"kumachan/rx"
	. "kumachan/lang"
	"kumachan/runtime/lib/container"
	"strconv"
)


func Optional2Maybe(obj rx.Object) Value {
	var opt = obj.(rx.Optional)
	if opt.HasValue {
		return Just(opt.Value)
	} else {
		return Na()
	}
}

type RxStackIterator struct {
	Stack  *rx.Stack
}
func (it RxStackIterator) Next() (Value, container.Seq, bool) {
	var val, rest, ok = it.Stack.Popped()
	if ok {
		return val, RxStackIterator { rest }, true
	} else {
		return nil, nil, false
	}
}
func (it RxStackIterator) GetItemType() reflect.Type {
	return ValueReflectType()
}

func AdaptReactiveDiff(diff rx.Observable) rx.Observable {
	var stack2seq = func(stack *rx.Stack) container.Seq {
		return RxStackIterator { stack }
	}
	return diff.Map(func(obj rx.Object) rx.Object {
		var pair = obj.(rx.Pair)
		var snapshots = pair.First.(rx.ReactiveSnapshots)
		var value = Value(pair.Second)
		return &ValProd { Elements: [] Value {
			&ValProd { Elements: [] Value {
				stack2seq(snapshots.Undo),
				stack2seq(snapshots.Redo),
			} },
			value,
		} }
	})
}

func recoverFromSyncCancellationPanic() {
	var err = recover()
	var _, is_sync_cancel = err.(SyncCancellationError)
	if err == nil || is_sync_cancel {
		// do nothing
	} else {
		panic(err)
	}
}

var nextProcessLevelGlobalId = uint64(0)

var EffectFunctions = map[string] Value {
	"connect": func(source rx.Observable, sink rx.Sink) rx.Observable {
		return rx.Connect(source, sink)
	},
	"sink-write": func(s rx.Sink, v Value) rx.Observable {
		return s.Emit(v)
	},
	"sink-adapt": func(s rx.Sink, f Value, h InteropContext) rx.Sink {
		var adapter = func(obj rx.Object) rx.Object {
			return h.Call(f, obj)
		}
		return rx.SinkAdapt(s, adapter)
	},
	"bus-watch": func(b rx.Bus) rx.Observable {
		return b.Watch()
	},
	"computed-read": func(computed rx.Observable) rx.Observable {
		return computed.TakeOneAsSingleAssumeSync().Map(func(opt_ rx.Object) rx.Object {
			var opt = opt_.(rx.Optional)
			if opt.HasValue {
				return opt.Value
			} else {
				panic("something went wrong")
			}
		})
	},
	"reactive-read": func(r rx.Reactive) rx.Observable {
		return r.Read()
	},
	"reactive-update": func(r rx.Reactive, f Value, h InteropContext) rx.Observable {
		return r.Update(func(old_state rx.Object) rx.Object {
			var new_state = h.Call(f, old_state)
			return new_state
		}, nil)
	},
	"reactive-adapt": func(r rx.Reactive, f Value, h InteropContext) rx.Sink {
		var in = func(old_state rx.Object) (func(rx.Object) rx.Object) {
			var adapter = h.Call(f, old_state)
			return func(obj rx.Object) rx.Object {
				var new_state = h.Call(adapter, obj)
				return new_state
			}
		}
		return rx.ReactiveAdapt(r, in)
	},
	"reactive-morph": func(r rx.Reactive, f Value, g Value, h InteropContext) rx.Reactive {
		var in = func(old_state rx.Object) (func(rx.Object) rx.Object) {
			var adapter = h.Call(f, old_state)
			return func(obj rx.Object) rx.Object {
				var new_state = h.Call(adapter, obj)
				return new_state
			}
		}
		var out = func(obj rx.Object) rx.Object {
			return h.Call(g, obj)
		}
		return rx.ReactiveMorph(r, in, out)
	},
	"reactive-snapshot": func(r rx.Reactive) rx.Observable {
		return r.Snapshot()
	},
	"reactive-list-consume": func(r rx.Reactive, k Value, h InteropContext) rx.Observable {
		return rx.KeyTrackedDynamicCombineLatestWaitReady (
			r.Watch().Map(func(list_ rx.Object) rx.Object {
				var list = list_.(container.List)
				return rx.KeyTrackedActionVector {
					HasKey: func(key_rx string) bool {
						var key = StringFromGoString(key_rx)
						return list.Has(key)
					},
					IterateKeys: func(f func(string)) {
						list.IterateKeySequence(func(key String) {
							var key_rx = GoStringFromString(key)
							f(key_rx)
						})
					},
					CloneKeys: func() []string {
						var keys = make([] string, 0, list.Length())
						list.IterateKeySequence(func(key String) {
							var key_rx = GoStringFromString(key)
							keys = append(keys, key_rx)
						})
						return keys
					},
					GetAction: func(key_rx string, index_source rx.Observable) rx.Observable {
						var key = StringFromGoString(key_rx)
						var in = func(old_state rx.Object) func(rx.Object) rx.Object {
							return func(new_item_value rx.Object) rx.Object {
								var old_list = old_state.(container.List)
								var new_list = old_list.Updated(key, func(_ Value) Value {
									return new_item_value
								})
								return new_list
							}
						}
						var out = func(state rx.Object) rx.Object {
							var list = state.(container.List)
							return list.Get(key)
						}
						var proj_key = &rx.KeyChain { Key: key_rx }
						var proj = rx.ReactiveProject(r, in, out, proj_key)
						var view = rx.ReactiveDistinctView(proj, RefEqual)
						var arg = &ValProd { Elements: [] Value {
							key, index_source, view,
						} }
						var item_action = h.Call(k, arg).(rx.Observable)
						return item_action
					},
				}
			}),
		)
	} ,
	"reactive-entity-undo": func(r rx.ReactiveEntity) rx.Observable {
		return r.Undo()
	},
	"reactive-entity-redo": func(r rx.ReactiveEntity) rx.Observable {
		return r.Redo()
	},
	"reactive-entity-watch-diff": func(r rx.ReactiveEntity) rx.Observable {
		return AdaptReactiveDiff(r.WatchDiff())
	},
	"reactive-entity-auto-snapshot": func(r rx.ReactiveEntity) rx.Reactive {
		return rx.AutoSnapshotReactive { Entity: r }
	},
	"blackhole": func() rx.Sink {
		return rx.BlackHole{}
	},
	"callback": func(f Value, h InteropContext) rx.Sink {
		return rx.Callback(func(obj rx.Object) rx.Observable {
			return h.Call(f, obj).(rx.Observable)
		})
	},
	"bus": func(_ Value, f Value, h InteropContext) rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateBus(), true
		}).Then(func(obj rx.Object) rx.Observable {
			return h.Call(f, obj).(rx.Observable)
		})
	},
	"reactive": func(init Value, k Value, h InteropContext) rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateReactive(init, RefEqual), true
		}).Then(func(r rx.Object) rx.Observable {
			return h.Call(k, r).(rx.Observable)
		})
	},
	"reactive+snapshot": func(init Value, k Value, h InteropContext) rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			var entity = rx.CreateReactive(init, RefEqual)
			var r = rx.AutoSnapshotReactive { Entity: entity }
			var undo = entity.Undo()
			var redo = entity.Redo()
			var diff = AdaptReactiveDiff(entity.WatchDiff())
			return &ValProd { Elements: [] Value {
				r, &ValProd { Elements: [] Value {
					undo, redo, diff,
				} },
			} }, true
		}).Then(func(r rx.Object) rx.Observable {
			return h.Call(k, r).(rx.Observable)
		})
	},
	"mutex": func(res Value, k Value, h InteropContext) rx.Observable {
		return rx.NewMutex(res).Then(func(mu rx.Object) rx.Observable {
			return h.Call(k, mu).(rx.Observable)
		})
	},
	"mutex-lock": func(mu *rx.Mutex, k Value, h InteropContext) rx.Observable {
		return mu.Lock(func(res rx.Object) rx.Observable {
			return h.Call(k, res).(rx.Observable)
		})
	},
	"new-mutable": func(init Value) rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateCell(init), true
		})
	},
	"with-mutable": func(init Value, f Value, h InteropContext) rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateCell(init), true
		}).Then(func(obj rx.Object) rx.Observable {
			return h.Call(f, obj).(rx.Observable)
		})
	},
	"mutable-get": func(cell rx.Cell) rx.Observable {
		return cell.Get()
	},
	"mutable-set": func(cell rx.Cell, v Value) rx.Observable {
		return cell.Set(v)
	},
	"mutable-swap": func(cell rx.Cell, f Value, h InteropContext) rx.Observable {
		return cell.Swap(func(v rx.Object) rx.Object {
			return h.Call(f, v)
		})
	},
	"as-source": func(action rx.Observable) rx.Observable {
		return action.DiscardComplete()
	},
	"with": func(main rx.Observable, side rx.Observable) rx.Observable {
		return main.With(side)
	},
	"gen-random": func() rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			return rand.Float64(), true
		})
	},
	"gen-monotonic-id-string": func() rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			var id = nextProcessLevelGlobalId
			nextProcessLevelGlobalId += 1
			return StringFromGoString(strconv.FormatUint(id, 16)), true
		})
	},
	"gen-monotonic-id": func() rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			var id = nextProcessLevelGlobalId
			nextProcessLevelGlobalId += 1
			return id, true
		})
	},
	"crash": func(msg String, h InteropContext) rx.Observable {
		const bold = "\033[1m"
		const red = "\033[31m"
		const reset = "\033[0m"
		var point = h.ErrorPoint()
		var source_point = point.Node.Point
		return rx.NewSync(func() (rx.Object, bool) {
			fmt.Fprintf (
				os.Stderr, "%v*** Crash: (%d, %d) at %s%v\n",
				bold+red,
				source_point.Row, source_point.Col, point.Node.CST.Name,
				reset,
			)
			fmt.Fprintf (
				os.Stderr, "%v%s%v\n",
				bold+red, GoStringFromString(msg), reset,
			)
			os.Exit(255)
			// noinspection GoUnreachableCode
			panic("program should have crashed")
		})
	},
	"go-thunk": func(f Value, h InteropContext) rx.Observable {
		return rx.NewGoroutineSingle(func(ctx *rx.Context) (rx.Object, bool) {
			defer recoverFromSyncCancellationPanic()
			return h.CallWithSyncContext(f, nil, ctx), true
		})
	},
	"go-seq": func(seq container.Seq, h InteropContext) rx.Observable {
		// TODO: should use CallWithSyncContext(next) when seq is a custom seq
		return rx.NewGoroutine(func(sender rx.Sender) {
			// defer recoverFromSyncCancellationPanic()
			if sender.Context().AlreadyCancelled() { return }
			for item, rest, ok := seq.Next(); ok; item, rest, ok = rest.Next() {
				sender.Next(item)
				if sender.Context().AlreadyCancelled() { return }
			}
			sender.Complete()
		})
	},
	"yield": func(v Value) rx.Observable {
		return rx.NewConstant(v)
	},
	"yield*-seq": func(seq container.Seq) rx.Observable {
		return rx.NewSyncSequence(func(next func(rx.Object))(bool,rx.Object) {
			for item, rest, ok := seq.Next(); ok; item, rest, ok = rest.Next() {
				next(item)
			}
			return true, nil
		})
	},
	"yield*-array": func(av Value) rx.Observable {
		var arr = container.ArrayFrom(av)
		return rx.NewSyncSequence(func(next func(rx.Object))(bool,rx.Object) {
			for i := uint(0); i < arr.Length; i += 1 {
				next(arr.GetItem(i))
			}
			return true, nil
		})
	},
	"take-one-as-single": func(e rx.Observable) rx.Observable {
		return e.TakeOneAsSingle().Map(func(val rx.Object) rx.Object {
			var opt = val.(rx.Optional)
			if opt.HasValue {
				return Just(opt.Value)
			} else {
				return Na()
			}
		})
	},
	"start-with": func(following rx.Observable, head_ Value) rx.Observable {
		var head = container.ArrayFrom(head_)
		return rx.Concat([] rx.Observable {
			rx.NewSyncSequence(func(next func(rx.Object))(bool,rx.Object) {
				for i := uint(0); i < head.Length; i += 1 {
					next(head.GetItem(i))
				}
				return true, nil
			}),
			following,
		})
	},
	"wait": func(bundle ProductValue) rx.Observable {
		var timeout = SingleValueFromBundle(bundle).(uint)
		return rx.Timer(timeout)
	},
	"tick": func(bundle ProductValue) rx.Observable {
		var interval = SingleValueFromBundle(bundle).(uint)
		return rx.Ticker(interval)
	},
	"wait-complete": func(e rx.Observable) rx.Observable {
		return e.WaitComplete()
	},
	"forever": func(e rx.Observable) rx.Observable {
		var repeat rx.Observable
		repeat = e.WaitComplete().Then(func(_ rx.Object) rx.Observable {
			return repeat
		})
		return repeat
	},
	"then": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.Then(func(val rx.Object) rx.Observable {
			return h.Call(f, val).(rx.Observable)
		})
	},
	"then-shortcut": func(a rx.Observable, b rx.Observable) rx.Observable {
		return a.Then(func(_ rx.Object) rx.Observable {
			return b
		})
	},
	"do": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.ChainSync(func(val rx.Object) rx.Observable {
			return h.Call(f, val).(rx.Observable)
		})
	},
	"do-shortcut": func(a rx.Observable, b rx.Observable) rx.Observable {
		return a.ChainSync(func(_ rx.Object) rx.Observable {
			return b
		})
	},
	"do-source": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.SyncThen(func(val rx.Object) rx.Observable {
			return h.Call(f, val).(rx.Observable)
		})
	},
	"do-source-shortcut": func(a rx.Observable, b rx.Observable) rx.Observable {
		return a.SyncThen(func(_ rx.Object) rx.Observable {
			return b
		})
	},
	"catch": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.Catch(func(err rx.Object) rx.Observable {
			return h.Call(f, err).(rx.Observable)
		})
	},
	"catch-retry": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.CatchRetry(func(err rx.Object) rx.Observable {
			return h.Call(f, err).(rx.Observable).Map(func(retry rx.Object) rx.Object {
				return FromBool(retry.(SumValue))
			})
		})
	},
	"catch-throw": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.CatchThrow(func(err rx.Object) rx.Object {
			return h.Call(f, err)
		})
	},
	"throw": func(err Value) rx.Observable {
		return rx.Throw(err)
	},
	"action-map": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.Map(func(val rx.Object) rx.Object {
			return h.Call(f, val)
		})
	},
	"action-filter-map": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.FilterMap(func(val rx.Object) (rx.Object, bool) {
			var maybe_mapped = h.Call(f, val).(SumValue)
			return Unwrap(maybe_mapped)
		})
	},
	"action-filter": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.Filter(func(val rx.Object) bool {
			return FromBool((h.Call(f, val)).(SumValue))
		})
	},
	"action-reduce": func(e rx.Observable, init Value, f Value, h InteropContext) rx.Observable {
		return e.Reduce(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
	},
	"action-scan": func(e rx.Observable, init Value, f Value, h InteropContext) rx.Observable {
		return e.Scan(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
	},
	"debounce-time": func(e rx.Observable, dueTime uint) rx.Observable {
		return e.DebounceTime(dueTime)
	},
	"switch-map": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.SwitchMap(func(val rx.Object) rx.Observable {
			return h.Call(f, val).(rx.Observable)
		})
	},
	"merge-map": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.MergeMap(func(val rx.Object) rx.Observable {
			return h.Call(f, val).(rx.Observable)
		})
	},
	"concat-map": func(e rx.Observable, f Value, h InteropContext) rx.Observable {
		return e.ConcatMap(func(val rx.Object) rx.Observable {
			return h.Call(f, val).(rx.Observable)
		})
	},
	"mix-map": func(e rx.Observable, n uint, f Value, h InteropContext) rx.Observable {
		return e.MixMap(func(val rx.Object) rx.Observable {
			return h.Call(f, val).(rx.Observable)
		}, n)
	},
	"action-merge": func(av Value) rx.Observable {
		var arr = container.ArrayFrom(av)
		var actions = make([] rx.Observable, arr.Length)
		for i := uint(0); i < arr.Length; i += 1 {
			actions[i] = arr.GetItem(i).(rx.Observable)
		}
		return rx.Merge(actions)
	},
	"action-concat": func(av Value) rx.Observable {
		var arr = container.ArrayFrom(av)
		var actions = make([] rx.Observable, arr.Length)
		for i := uint(0); i < arr.Length; i += 1 {
			actions[i] = arr.GetItem(i).(rx.Observable)
		}
		return rx.Concat(actions)
	},
	"distinct-until-changed": func(a rx.Observable, eq Value, h InteropContext) rx.Observable {
		return a.DistinctUntilChanged(func(obj1 rx.Object, obj2 rx.Object) bool {
			var pair = &ValProd { Elements: [] Value { obj1, obj2 } }
			return FromBool(h.Call(eq, pair).(SumValue))
		})
	},
	"with-latest-from": func(a rx.Observable, values rx.Observable) rx.Observable {
		return a.WithLatestFrom(values).Map(func(p rx.Object) rx.Object {
			var pair = p.(rx.Pair)
			return &ValProd { Elements: [] Value {
				pair.First,
				Optional2Maybe(pair.Second),
			} }
		})
	},
	"with-latest-from-reactive": func(a rx.Observable, r rx.Reactive) rx.Observable {
		return a.WithLatestFrom(r.Watch()).Map(func(p rx.Object) rx.Object {
			var pair = p.(rx.Pair)
			var r_opt = pair.Second.(rx.Optional)
			if !(r_opt.HasValue) { panic("something went wrong") }
			var r_value = r_opt.Value
			return &ValProd { Elements: [] Value {
				pair.First,
				r_value,
			} }
		})
	},
	"combine-latest": func(tuple ProductValue) rx.Observable {
		var actions = make([] rx.Observable, len(tuple.Elements))
		for i, el := range tuple.Elements {
			actions[i] = el.(rx.Observable)
		}
		return rx.CombineLatest(actions).Map(func(raw rx.Object) rx.Object {
			var raw_values = raw.([] rx.Optional)
			var values = make([] Value, len(actions))
			for i := 0; i < len(values); i += 1 {
				values[i] = Optional2Maybe(raw_values[i])
			}
			return &ValProd { Elements: values }
		})
	},
	"combine-latest*": func(tuple ProductValue) rx.Observable {
		var actions = make([] rx.Observable, len(tuple.Elements))
		for i, el := range tuple.Elements {
			actions[i] = el.(rx.Observable)
		}
		return rx.CombineLatestWaitReady(actions).Map(func(values_ rx.Object) rx.Object {
			var values = values_.([] Value)
			return &ValProd { Elements: values }
		})
	},
	"combine-latest*-array": func(v Value) rx.Observable {
		var array = container.ArrayFrom(v)
		var actions = make([] rx.Observable, array.Length)
		for i := uint(0); i < array.Length; i += 1 {
			actions[i] = array.GetItem(i).(rx.Observable)
		}
		return rx.CombineLatestWaitReady(actions)
	},
	"computed": func(tuple ProductValue, f Value, h InteropContext) rx.Observable {
		var actions = make([] rx.Observable, len(tuple.Elements))
		for i, el := range tuple.Elements {
			actions[i] = el.(rx.Observable)
		}
		return rx.CombineLatestWaitReady(actions).Map(func(values_ rx.Object) rx.Object {
			var values = values_.([] Value)
			return h.Call(f, &ValProd { Elements: values })
		})
	},
}

