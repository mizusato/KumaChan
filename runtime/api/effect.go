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

func AdaptReactiveDiff(diff rx.Effect) rx.Effect {
	var stack2seq = func(stack *rx.Stack) container.Seq {
		return container.MappedSeq {
			Input:  RxStackIterator { stack },
			Mapper: func(v Value) Value {
				return v.(rx.ReactiveStateChange).Value
			},
		}
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
	"sink-emit": func(s rx.Sink, v Value) rx.Effect {
		return s.Emit(v)
	},
	"sink-adapt": func(s rx.Sink, f Value, h InteropContext) rx.Sink {
		var adapter = func(obj rx.Object) rx.Object {
			return h.Call(f, obj)
		}
		return rx.SinkAdapt(s, adapter)
	},
	"bus-watch": func(b rx.Bus) rx.Effect {
		return b.Watch()
	},
	"reactive-update": func(r rx.Reactive, f Value, h InteropContext) rx.Effect {
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
	"reactive-snapshot": func(r rx.Reactive) rx.Effect {
		return r.Snapshot()
	},
	"reactive-list-consume": func(r rx.Reactive, k Value, h InteropContext) rx.Effect {
		return rx.KeyTrackedDynamicCombineLatestWaitReady (
			r.Watch().Map(func(list_ rx.Object) rx.Object {
				var list = list_.(container.List)
				return rx.KeyTrackedEffectVector {
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
					GetEffect: func(key_rx string, index_source rx.Effect) rx.Effect {
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
						var proj_key = &rx.KeyChain { Key: key_rx}
						var proj = rx.ReactiveProject(r, in, out, proj_key)
						var arg = &ValProd { Elements: [] Value {
							key, index_source, proj,
						} }
						var item_effect = h.Call(k, arg).(rx.Effect)
						return item_effect
					},
				}
			}),
		)
	} ,
	"reactive-entity-undo": func(r rx.ReactiveEntity) rx.Effect {
		return r.Undo()
	},
	"reactive-entity-redo": func(r rx.ReactiveEntity) rx.Effect {
		return r.Redo()
	},
	"reactive-entity-watch-diff": func(r rx.ReactiveEntity) rx.Effect {
		return AdaptReactiveDiff(r.WatchDiff())
	},
	"reactive-entity-auto-snapshot": func(r rx.ReactiveEntity) rx.Reactive {
		return rx.AutoSnapshotReactive { Entity: r }
	},
	"callback": func(f Value, h InteropContext) rx.Sink {
		return rx.Callback(func(obj rx.Object) rx.Effect {
			return h.Call(f, obj).(rx.Effect)
		})
	},
	"new-bus": func() rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateBus(), true
		})
	},
	"with-bus": func(f Value, h InteropContext) rx.Effect {
		// this func is not useful due to the limitation of type inference
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateBus(), true
		}).ConcatMap(func(obj rx.Object) rx.Effect {
			return h.Call(f, obj).(rx.Effect)
		})
	},
	"new-reactive": func(init Value) rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateReactive(init), true
		})
	},
	"with-reactive": func(init Value, f Value, h InteropContext) rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateReactive(init), true
		}).Then(func(obj rx.Object) rx.Effect {
			return h.Call(f, obj).(rx.Effect)
		})
	},
	"with-auto-snapshot": func(init Value, f Value, h InteropContext) rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			var entity = rx.CreateReactive(init)
			var r = rx.AutoSnapshotReactive { Entity: entity }
			var undo = entity.Undo()
			var redo = entity.Redo()
			var diff = AdaptReactiveDiff(entity.WatchDiff())
			return &ValProd { Elements: [] Value {
				r, &ValProd { Elements: [] Value {
					undo, redo, diff,
				} },
			} }, true
		}).Then(func(obj rx.Object) rx.Effect {
			return h.Call(f, obj).(rx.Effect)
		})
	},
	"new-mutable": func(init Value) rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateCell(init), true
		})
	},
	"with-mutable": func(init Value, f Value, h InteropContext) rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateCell(init), true
		}).Then(func(obj rx.Object) rx.Effect {
			return h.Call(f, obj).(rx.Effect)
		})
	},
	"mutable-get": func(cell rx.Cell) rx.Effect {
		return cell.Get()
	},
	"mutable-set": func(cell rx.Cell, v Value) rx.Effect {
		return cell.Set(v)
	},
	"mutable-swap": func(cell rx.Cell, f Value, h InteropContext) rx.Effect {
		return cell.Swap(func(v rx.Object) rx.Object {
			return h.Call(f, v)
		})
	},
	"with": func(main rx.Effect, side rx.Effect) rx.Effect {
		return rx.Merge([] rx.Effect { main, side.DiscardValues() })
	},
	"random": func() rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return rand.Float64(), true
		})
	},
	"proc-gid": func() rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			var id = nextProcessLevelGlobalId
			nextProcessLevelGlobalId += 1
			return StringFromGoString(strconv.FormatUint(id, 16)), true
		})
	},
	"crash": func(msg String, h InteropContext) rx.Effect {
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
	"go": func(f Value, h InteropContext) rx.Effect {
		return rx.NewGoroutineSingle(func() (rx.Object, bool) {
			return h.Call(f, nil), true
		})
	},
	"go*": func(seq container.Seq, h InteropContext) rx.Effect {
		return rx.NewGoroutine(func(sender rx.Sender) {
			if sender.Context().AlreadyCancelled() { return }
			for item, rest, ok := seq.Next(); ok; item, rest, ok = rest.Next() {
				sender.Next(item)
				if sender.Context().AlreadyCancelled() { return }
			}
			sender.Complete()
		})
	},
	"yield": func(v Value) rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return v, true
		})
	},
	"yield*-range": func(l uint, r uint) rx.Effect {
		return rx.NewSyncSequence(func(next func(rx.Object))(bool,rx.Object) {
			for i := l; i < r; i += 1 {
				next(i)
			}
			return true, nil
		})
	},
	"yield*-seq": func(seq container.Seq) rx.Effect {
		return rx.NewSyncSequence(func(next func(rx.Object))(bool,rx.Object) {
			for item, rest, ok := seq.Next(); ok; item, rest, ok = rest.Next() {
				next(item)
			}
			return true, nil
		})
	},
	"yield*-array": func(av Value) rx.Effect {
		var arr = container.ArrayFrom(av)
		return rx.NewSyncSequence(func(next func(rx.Object))(bool,rx.Object) {
			for i := uint(0); i < arr.Length; i += 1 {
				next(arr.GetItem(i))
			}
			return true, nil
		})
	},
	"take-one": func(e rx.Effect) rx.Effect {
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
	"wait": func(bundle ProductValue) rx.Effect {
		var timeout = SingleValueFromBundle(bundle).(uint)
		return rx.Timer(timeout)
	},
	"tick": func(bundle ProductValue) rx.Effect {
		var interval = SingleValueFromBundle(bundle).(uint)
		return rx.Ticker(interval)
	},
	"wait-complete": func(e rx.Effect) rx.Effect {
		return e.WaitComplete()
	},
	"forever": func(e rx.Effect) rx.Effect {
		var repeat rx.Effect
		repeat = e.WaitComplete().Then(func(_ rx.Object) rx.Effect {
			return repeat
		})
		return repeat
	},
	"then": func(e rx.Effect, f Value, h InteropContext) rx.Effect {
		return e.Then(func(val rx.Object) rx.Effect {
			return h.Call(f, val).(rx.Effect)
		})
	},
	"then-shortcut": func(a rx.Effect, b rx.Effect) rx.Effect {
		return a.Then(func(_ rx.Object) rx.Effect {
			return b
		})
	},
	"catch": func(e rx.Effect, f Value, h InteropContext) rx.Effect {
		return e.Catch(func(err rx.Object) rx.Effect {
			return h.Call(f, err).(rx.Effect)
		})
	},
	"catch-retry": func(e rx.Effect, f Value, h InteropContext) rx.Effect {
		return e.CatchRetry(func(err rx.Object) rx.Effect {
			return h.Call(f, err).(rx.Effect).Map(func(retry rx.Object) rx.Object {
				return FromBool(retry.(SumValue))
			})
		})
	},
	"catch-throw": func(e rx.Effect, f Value, h InteropContext) rx.Effect {
		return e.CatchThrow(func(err rx.Object) rx.Object {
			return h.Call(f, err)
		})
	},
	"throw": func(err Value) rx.Effect {
		return rx.Throw(err)
	},
	"effect-map": func(e rx.Effect, f Value, h InteropContext) rx.Effect {
		return e.Map(func(val rx.Object) rx.Object {
			return h.Call(f, val)
		})
	},
	"effect-map?": func(e rx.Effect, f Value, h InteropContext) rx.Effect {
		return e.FilterMap(func(val rx.Object) (rx.Object, bool) {
			var maybe_mapped = h.Call(f, val).(SumValue)
			return Unwrap(maybe_mapped)
		})
	},
	"effect-filter": func(e rx.Effect, f Value, h InteropContext) rx.Effect {
		return e.Filter(func(val rx.Object) bool {
			return FromBool((h.Call(f, val)).(SumValue))
		})
	},
	"effect-reduce": func(e rx.Effect, init Value, f Value, h InteropContext) rx.Effect {
		return e.Reduce(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
	},
	"effect-scan": func(e rx.Effect, init Value, f Value, h InteropContext) rx.Effect {
		return e.Scan(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
	},
	"debounce-time": func(e rx.Effect, dueTime uint) rx.Effect {
		return e.DebounceTime(dueTime)
	},
	"switch-map": func(e rx.Effect, f Value, h InteropContext) rx.Effect {
		return e.SwitchMap(func(val rx.Object) rx.Effect {
			return h.Call(f, val).(rx.Effect)
		})
	},
	"merge-map": func(e rx.Effect, f Value, h InteropContext) rx.Effect {
		return e.MergeMap(func(val rx.Object) rx.Effect {
			return h.Call(f, val).(rx.Effect)
		})
	},
	"concat-map": func(e rx.Effect, f Value, h InteropContext) rx.Effect {
		return e.ConcatMap(func(val rx.Object) rx.Effect {
			return h.Call(f, val).(rx.Effect)
		})
	},
	"mix-map": func(e rx.Effect, n uint, f Value, h InteropContext) rx.Effect {
		return e.MixMap(func(val rx.Object) rx.Effect {
			return h.Call(f, val).(rx.Effect)
		}, n)
	},
	"effect-merge": func(av Value) rx.Effect {
		var arr = container.ArrayFrom(av)
		var effects = make([] rx.Effect, arr.Length)
		for i := uint(0); i < arr.Length; i += 1 {
			effects[i] = arr.GetItem(i).(rx.Effect)
		}
		return rx.Merge(effects)
	},
	"effect-concat": func(av Value) rx.Effect {
		var arr = container.ArrayFrom(av)
		var effects = make([] rx.Effect, arr.Length)
		for i := uint(0); i < arr.Length; i += 1 {
			effects[i] = arr.GetItem(i).(rx.Effect)
		}
		return rx.Concat(effects)
	},
	"with-latest-from": func(signal rx.Effect, values rx.Effect) rx.Effect {
		return signal.WithLatestFrom(values).Map(func(p rx.Object) rx.Object {
			var pair = p.(rx.Pair)
			return &ValProd { Elements: [] Value {
				pair.First,
				Optional2Maybe(pair.Second),
			} }
		})
	},
	"with-latest-from-reactive": func(signal rx.Effect, r rx.Reactive) rx.Effect {
		return signal.WithLatestFrom(r.Watch()).Map(func(p rx.Object) rx.Object {
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
	"combine-latest": func(tuple ProductValue) rx.Effect {
		var effects = make([] rx.Effect, len(tuple.Elements))
		for i, el := range tuple.Elements {
			effects[i] = el.(rx.Effect)
		}
		return rx.CombineLatest(effects).Map(func(raw rx.Object) rx.Object {
			var raw_values = raw.([] rx.Optional)
			var values = make([] Value, len(effects))
			for i := 0; i < len(values); i += 1 {
				values[i] = Optional2Maybe(raw_values[i])
			}
			return &ValProd { Elements: values }
		})
	},
	"combine-latest!": func(tuple ProductValue) rx.Effect {
		var effects = make([] rx.Effect, len(tuple.Elements))
		for i, el := range tuple.Elements {
			effects[i] = el.(rx.Effect)
		}
		return rx.CombineLatestWaitReady(effects).Map(func(values_ rx.Object) rx.Object {
			var values = values_.([] Value)
			return &ValProd { Elements: values }
		})
	},
	"combine-latest!-array": func(v Value) rx.Effect {
		var array = container.ArrayFrom(v)
		var effects = make([] rx.Effect, array.Length)
		for i := uint(0); i < array.Length; i += 1 {
			effects[i] = array.GetItem(i).(rx.Effect)
		}
		return rx.CombineLatestWaitReady(effects)
	},
}

