package lib

import (
	"os"
	"fmt"
	"math/rand"
	. "kumachan/runtime/common"
	"kumachan/runtime/rx"
	"kumachan/runtime/lib/container"
)


func Optional2Maybe(obj rx.Object) Value {
	var opt = obj.(rx.Optional)
	if opt.HasValue {
		return Just(opt.Value)
	} else {
		return Na()
	}
}

var EffectFunctions = map[string] Value {
	"send": func(sink rx.Sink, v Value) rx.Effect {
		return sink.Send(v)
	},
	"receive": func(source rx.Source) rx.Effect {
		return source.Receive()
	},
	"source-map": func(source rx.Source, f Value, h MachineHandle) rx.Source {
		return &rx.MappedSource {
			Source: source,
			Mapper: func(obj rx.Object) rx.Object {
				return h.Call(f, obj)
			},
		}
	},
	"sink-adapt": func(sink rx.Sink, f Value, h MachineHandle) rx.Sink {
		return &rx.AdaptedSink {
			Sink:    sink,
			Adapter: func(obj rx.Object) rx.Object {
				return h.Call(f, obj)
			},
		}
	},
	"latch-adapt": func(latch *rx.Latch, f Value, h MachineHandle) rx.Sink {
		return &rx.AdaptedLatch {
			Latch:      latch,
			GetAdapter: func(old_state rx.Object) (func(rx.Object) rx.Object) {
				var adapter = h.Call(f, old_state)
				return func(obj rx.Object) rx.Object {
					var new_state = h.Call(adapter, obj)
					return new_state
				}
			},
		}
	},
	"latch-combine": func(tuple ProductValue) rx.Source {
		var latches = make([] *rx.Latch, len(tuple.Elements))
		for i, el := range tuple.Elements {
			latches[i] = el.(*rx.Latch)
		}
		return &rx.MappedSource {
			Source: &rx.CombinedLatch { Elements: latches },
			Mapper: func(obj rx.Object) rx.Object {
				var values = obj.([] rx.Object)
				return &ValProd { Elements: values }
			},
		}
	},
	"Source from *": func(source rx.Source) rx.Source {
		return source
	},
	"Sink from *": func(sink rx.Sink) rx.Sink {
		return sink
	},
	"new-bus": func() rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateBus(), true
		})
	},
	"new-latch": func(init Value) rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateLatch(init), true
		})
	},
	"latch-reset": func(l *rx.Latch) rx.Effect {
		return l.Reset()
	},
	"new-mutable": func(init Value) rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return rx.CreateCell(init), true
		})
	},
	"mutable-get": func(cell rx.Cell) rx.Effect {
		return cell.Get()
	},
	"mutable-set": func(cell rx.Cell, v Value) rx.Effect {
		return cell.Set(v)
	},
	"mutable-swap": func(cell rx.Cell, f Value, h MachineHandle) rx.Effect {
		return cell.Swap(func(v rx.Object) rx.Object {
			return h.Call(f, v)
		})
	},
	"random": func() rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return rand.Float64(), true
		})
	},
	"crash": func(msg String, h MachineHandle) rx.Effect {
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
	"emit": func(v Value) rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			return v, true
		})
	},
	"emit*-range": func(l uint, r uint) rx.Effect {
		return rx.NewSyncSequence(func(next func(rx.Object))(bool,rx.Object) {
			for i := l; i < r; i += 1 {
				next(i)
			}
			return true, nil
		})
	},
	"emit*-seq": func(seq container.Seq) rx.Effect {
		return rx.NewSyncSequence(func(next func(rx.Object))(bool,rx.Object) {
			for item, rest, ok := seq.Next(); ok; item, rest, ok = rest.Next() {
				next(item)
			}
			return true, nil
		})
	},
	"emit*-array": func(av Value) rx.Effect {
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
	"adapt-no-except": func(v Value) Value {
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
	"then": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.Then(func(val rx.Object) rx.Effect {
			return h.Call(f, val).(rx.Effect)
		})
	},
	"then-shortcut": func(a rx.Effect, b rx.Effect) rx.Effect {
		return a.Then(func(_ rx.Object) rx.Effect {
			return b
		})
	},
	"catch": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.Catch(func(err rx.Object) rx.Effect {
			return h.Call(f, err).(rx.Effect)
		})
	},
	"catch-retry": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.CatchRetry(func(err rx.Object) rx.Effect {
			return h.Call(f, err).(rx.Effect).Map(func(retry rx.Object) rx.Object {
				return BoolFrom(retry.(SumValue))
			})
		})
	},
	"catch-throw": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.CatchThrow(func(err rx.Object) rx.Object {
			return h.Call(f, err)
		})
	},
	"throw": func(err Value) rx.Effect {
		return rx.Throw(err)
	},
	"effect-map": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.Map(func(val rx.Object) rx.Object {
			return h.Call(f, val)
		})
	},
	"effect-filter": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.Filter(func(val rx.Object) bool {
			return BoolFrom((h.Call(f, val)).(SumValue))
		})
	},
	"effect-reduce": func(e rx.Effect, init Value, f Value, h MachineHandle) rx.Effect {
		return e.Reduce(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
	},
	"effect-scan": func(e rx.Effect, init Value, f Value, h MachineHandle) rx.Effect {
		return e.Scan(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
	},
	"debounce-time": func(e rx.Effect, dueTime uint) rx.Effect {
		return e.DebounceTime(dueTime)
	},
	"switch-map": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.SwitchMap(func(val rx.Object) rx.Effect {
			return h.Call(f, val).(rx.Effect)
		})
	},
	"merge-map": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.MergeMap(func(val rx.Object) rx.Effect {
			return h.Call(f, val).(rx.Effect)
		})
	},
	"concat-map": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.ConcatMap(func(val rx.Object) rx.Effect {
			return h.Call(f, val).(rx.Effect)
		})
	},
	"mix-map": func(e rx.Effect, n uint, f Value, h MachineHandle) rx.Effect {
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
}

