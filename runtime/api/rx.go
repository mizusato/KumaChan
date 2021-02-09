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

func AdaptReactiveDiff(diff rx.Action) rx.Action {
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

var nextProcessLevelGlobalId = uint64(0)

var EffectFunctions = map[string] Value {
	"sink-write": func(s rx.Sink, v Value) rx.Action {
		return s.Emit(v)
	},
	"sink-adapt": func(s rx.Sink, f Value, h InteropContext) rx.Sink {
		var adapter = func(obj rx.Object) rx.Object {
			return h.Call(f, obj)
		}
		return rx.SinkAdapt(s, adapter)
	},
	"bus-watch": func(b rx.Bus) rx.Action {
		return b.Watch()
	},
	"reactive-update": func(r rx.Reactive, f Value, h InteropContext) rx.Action {
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
	"reactive-snapshot": func(r rx.Reactive) rx.Action {
		return r.Snapshot()
	},
	"reactive-read": func(r rx.Reactive) rx.Action {
		return r.Read()
	},
	"reactive-list-consume": func(r rx.Reactive, k Value, h InteropContext) rx.Action {
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
					GetAction: func(key_rx string, index_source rx.Action) rx.Action {
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
						var item_action = h.Call(k, arg).(rx.Action)
						return item_action
					},
				}
			}),
		)
	} ,
	"reactive-entity-undo": func(r rx.ReactiveEntity) rx.Action {
		return r.Undo()
	},
	"reactive-entity-redo": func(r rx.ReactiveEntity) rx.Action {
		return r.Redo()
	},
	"reactive-entity-watch-diff": func(r rx.ReactiveEntity) rx.Action {
		return AdaptReactiveDiff(r.WatchDiff())
	},
	"reactive-entity-auto-snapshot": func(r rx.ReactiveEntity) rx.Reactive {
		return rx.AutoSnapshotReactive { Entity: r }
	},
	"callback": func(f Value, h InteropContext) rx.Sink {
		return rx.Callback(func(obj rx.Object) rx.Action {
			return h.Call(f, obj).(rx.Action)
		})
	},
	"new-bus": func() rx.Action {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateBus(), true
		})
	},
	"with-bus": func(f Value, h InteropContext) rx.Action {
		// this func is not useful due to the limitation of type inference
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateBus(), true
		}).ConcatMap(func(obj rx.Object) rx.Action {
			return h.Call(f, obj).(rx.Action)
		})
	},
	"new-reactive": func(init Value) rx.Action {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateReactive(init, RefEqual), true
		})
	},
	"with-reactive": func(init Value, f Value, h InteropContext) rx.Action {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateReactive(init, RefEqual), true
		}).Then(func(obj rx.Object) rx.Action {
			return h.Call(f, obj).(rx.Action)
		})
	},
	"with-auto-snapshot": func(init Value, f Value, h InteropContext) rx.Action {
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
		}).Then(func(obj rx.Object) rx.Action {
			return h.Call(f, obj).(rx.Action)
		})
	},
	"new-mutable": func(init Value) rx.Action {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateCell(init), true
		})
	},
	"with-mutable": func(init Value, f Value, h InteropContext) rx.Action {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateCell(init), true
		}).Then(func(obj rx.Object) rx.Action {
			return h.Call(f, obj).(rx.Action)
		})
	},
	"mutable-get": func(cell rx.Cell) rx.Action {
		return cell.Get()
	},
	"mutable-set": func(cell rx.Cell, v Value) rx.Action {
		return cell.Set(v)
	},
	"mutable-swap": func(cell rx.Cell, f Value, h InteropContext) rx.Action {
		return cell.Swap(func(v rx.Object) rx.Object {
			return h.Call(f, v)
		})
	},
	"with": func(main rx.Action, side rx.Action) rx.Action {
		return rx.Merge([] rx.Action {main, side.DiscardValues() })
	},
	"gen-random": func() rx.Action {
		return rx.NewSync(func() (rx.Object, bool) {
			return rand.Float64(), true
		})
	},
	"gen-sequential-id": func() rx.Action {
		return rx.NewSync(func() (rx.Object, bool) {
			var id = nextProcessLevelGlobalId
			nextProcessLevelGlobalId += 1
			return StringFromGoString(strconv.FormatUint(id, 16)), true
		})
	},
	"no-op": func() rx.Action {
		return rx.Noop()
	},
	"crash": func(msg String, h InteropContext) rx.Action {
		const bold = "\033[1m"
		const red = "\033[31m"
		const reset = "\033[0m"
		var point = h.GetErrorPoint()
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
	"go": func(f Value, h InteropContext) rx.Action {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			return h.Call(f, nil), true
		})
	},
	"go*": func(seq container.Seq, h InteropContext) rx.Action {
		return rx.NewGoroutine(func(sender rx.Sender) {
			if sender.Context().AlreadyCancelled() { return }
			for item, rest, ok := seq.Next(); ok; item, rest, ok = rest.Next() {
				sender.Next(item)
				if sender.Context().AlreadyCancelled() { return }
			}
			sender.Complete()
		})
	},
	"yield": func(v Value) rx.Action {
		return rx.NewSync(func() (rx.Object, bool) {
			return v, true
		})
	},
	"yield*-interval": func(l uint, r uint) rx.Action {
		return rx.NewSyncSequence(func(next func(rx.Object))(bool,rx.Object) {
			for i := l; i < r; i += 1 {
				next(i)
			}
			return true, nil
		})
	},
	"yield*-seq": func(seq container.Seq) rx.Action {
		return rx.NewSyncSequence(func(next func(rx.Object))(bool,rx.Object) {
			for item, rest, ok := seq.Next(); ok; item, rest, ok = rest.Next() {
				next(item)
			}
			return true, nil
		})
	},
	"yield*-array": func(av Value) rx.Action {
		var arr = container.ArrayFrom(av)
		return rx.NewSyncSequence(func(next func(rx.Object))(bool,rx.Object) {
			for i := uint(0); i < arr.Length; i += 1 {
				next(arr.GetItem(i))
			}
			return true, nil
		})
	},
	"take-one": func(e rx.Action) rx.Action {
		return e.TakeOne().Map(func(val rx.Object) rx.Object {
			var opt = val.(rx.Optional)
			if opt.HasValue {
				return Just(opt.Value)
			} else {
				return Na()
			}
		})
	},
	"assume-except": func(v Value) Value {
		return v
	},
	"wait": func(bundle ProductValue) rx.Action {
		var timeout = SingleValueFromBundle(bundle).(uint)
		return rx.Timer(timeout)
	},
	"tick": func(bundle ProductValue) rx.Action {
		var interval = SingleValueFromBundle(bundle).(uint)
		return rx.Ticker(interval)
	},
	"wait-complete": func(e rx.Action) rx.Action {
		return e.WaitComplete()
	},
	"forever": func(e rx.Action) rx.Action {
		var repeat rx.Action
		repeat = e.WaitComplete().Then(func(_ rx.Object) rx.Action {
			return repeat
		})
		return repeat
	},
	"then": func(e rx.Action, f Value, h InteropContext) rx.Action {
		return e.Then(func(val rx.Object) rx.Action {
			return h.Call(f, val).(rx.Action)
		})
	},
	"then-shortcut": func(a rx.Action, b rx.Action) rx.Action {
		return a.Then(func(_ rx.Object) rx.Action {
			return b
		})
	},
	"catch": func(e rx.Action, f Value, h InteropContext) rx.Action {
		return e.Catch(func(err rx.Object) rx.Action {
			return h.Call(f, err).(rx.Action)
		})
	},
	"catch-retry": func(e rx.Action, f Value, h InteropContext) rx.Action {
		return e.CatchRetry(func(err rx.Object) rx.Action {
			return h.Call(f, err).(rx.Action).Map(func(retry rx.Object) rx.Object {
				return FromBool(retry.(SumValue))
			})
		})
	},
	"catch-throw": func(e rx.Action, f Value, h InteropContext) rx.Action {
		return e.CatchThrow(func(err rx.Object) rx.Object {
			return h.Call(f, err)
		})
	},
	"throw": func(err Value) rx.Action {
		return rx.Throw(err)
	},
	"action-map": func(e rx.Action, f Value, h InteropContext) rx.Action {
		return e.Map(func(val rx.Object) rx.Object {
			return h.Call(f, val)
		})
	},
	"action-filter-map": func(e rx.Action, f Value, h InteropContext) rx.Action {
		return e.FilterMap(func(val rx.Object) (rx.Object, bool) {
			var maybe_mapped = h.Call(f, val).(SumValue)
			return Unwrap(maybe_mapped)
		})
	},
	"action-filter": func(e rx.Action, f Value, h InteropContext) rx.Action {
		return e.Filter(func(val rx.Object) bool {
			return FromBool((h.Call(f, val)).(SumValue))
		})
	},
	"action-reduce": func(e rx.Action, init Value, f Value, h InteropContext) rx.Action {
		return e.Reduce(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
	},
	"action-scan": func(e rx.Action, init Value, f Value, h InteropContext) rx.Action {
		return e.Scan(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
	},
	"debounce-time": func(e rx.Action, dueTime uint) rx.Action {
		return e.DebounceTime(dueTime)
	},
	"switch-map": func(e rx.Action, f Value, h InteropContext) rx.Action {
		return e.SwitchMap(func(val rx.Object) rx.Action {
			return h.Call(f, val).(rx.Action)
		})
	},
	"merge-map": func(e rx.Action, f Value, h InteropContext) rx.Action {
		return e.MergeMap(func(val rx.Object) rx.Action {
			return h.Call(f, val).(rx.Action)
		})
	},
	"concat-map": func(e rx.Action, f Value, h InteropContext) rx.Action {
		return e.ConcatMap(func(val rx.Object) rx.Action {
			return h.Call(f, val).(rx.Action)
		})
	},
	"mix-map": func(e rx.Action, n uint, f Value, h InteropContext) rx.Action {
		return e.MixMap(func(val rx.Object) rx.Action {
			return h.Call(f, val).(rx.Action)
		}, n)
	},
	"action-merge": func(av Value) rx.Action {
		var arr = container.ArrayFrom(av)
		var actions = make([] rx.Action, arr.Length)
		for i := uint(0); i < arr.Length; i += 1 {
			actions[i] = arr.GetItem(i).(rx.Action)
		}
		return rx.Merge(actions)
	},
	"action-concat": func(av Value) rx.Action {
		var arr = container.ArrayFrom(av)
		var actions = make([] rx.Action, arr.Length)
		for i := uint(0); i < arr.Length; i += 1 {
			actions[i] = arr.GetItem(i).(rx.Action)
		}
		return rx.Concat(actions)
	},
	"distinct-until-changed": func(a rx.Action, eq Value, h InteropContext) rx.Action {
		return a.DistinctUntilChanged(func(obj1 rx.Object, obj2 rx.Object) bool {
			var pair = &ValProd { Elements: [] Value { obj1, obj2 } }
			return FromBool(h.Call(eq, pair).(SumValue))
		})
	},
	"with-latest-from": func(a rx.Action, values rx.Action) rx.Action {
		return a.WithLatestFrom(values).Map(func(p rx.Object) rx.Object {
			var pair = p.(rx.Pair)
			return &ValProd { Elements: [] Value {
				pair.First,
				Optional2Maybe(pair.Second),
			} }
		})
	},
	"with-latest-from-reactive": func(a rx.Action, r rx.Reactive) rx.Action {
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
	"combine-latest": func(tuple ProductValue) rx.Action {
		var actions = make([] rx.Action, len(tuple.Elements))
		for i, el := range tuple.Elements {
			actions[i] = el.(rx.Action)
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
	"combine-latest!": func(tuple ProductValue) rx.Action {
		var actions = make([] rx.Action, len(tuple.Elements))
		for i, el := range tuple.Elements {
			actions[i] = el.(rx.Action)
		}
		return rx.CombineLatestWaitReady(actions).Map(func(values_ rx.Object) rx.Object {
			var values = values_.([] Value)
			return &ValProd { Elements: values }
		})
	},
	"combine-latest!-array": func(v Value) rx.Action {
		var array = container.ArrayFrom(v)
		var actions = make([] rx.Action, array.Length)
		for i := uint(0); i < array.Length; i += 1 {
			actions[i] = array.GetItem(i).(rx.Action)
		}
		return rx.CombineLatestWaitReady(actions)
	},
	"computed": func(tuple ProductValue, f Value, h InteropContext) rx.Action {
		var actions = make([] rx.Action, len(tuple.Elements))
		for i, el := range tuple.Elements {
			actions[i] = el.(rx.Action)
		}
		return rx.CombineLatestWaitReady(actions).Map(func(values_ rx.Object) rx.Object {
			var values = values_.([] Value)
			return h.Call(f, &ValProd { Elements: values })
		})
	},
}

