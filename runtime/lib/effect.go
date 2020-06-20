package lib

import (
	"os"
	"fmt"
	"math/rand"
	. "kumachan/runtime/common"
	"kumachan/runtime/rx"
	"kumachan/runtime/lib/container"
)


var EffectFunctions = map[string] Value {
	"listen": func(source rx.Source) rx.Effect {
		return source.Listen()
	},
	"random": func() rx.Effect {
		return rx.CreateBlockingEffect(func() (rx.Object, bool) {
			return rand.Float64(), true
		})
	},
	"crash": func(msg String, h MachineHandle) rx.Effect {
		const bold = "\033[1m"
		const red = "\033[31m"
		const reset = "\033[0m"
		var point = h.GetErrorPoint()
		var source_point = point.Node.Point
		return rx.CreateBlockingEffect(func() (rx.Object, bool) {
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
		return rx.CreateBlockingEffect(func() (rx.Object, bool) {
			return v, true
		})
	},
	"emit*-range": func(l uint, r uint) rx.Effect {
		return rx.CreateBlockingSequenceEffect(func(next func(rx.Object))(bool,rx.Object) {
			for i := l; i < r; i += 1 {
				next(i)
			}
			return true, nil
		})
	},
	"emit*-seq": func(seq container.Seq) rx.Effect {
		return rx.CreateBlockingSequenceEffect(func(next func(rx.Object))(bool,rx.Object) {
			for item, rest, ok := seq.Next(); ok; item, rest, ok = rest.Next() {
				next(item)
			}
			return true, nil
		})
	},
	"emit*-array": func(av Value) rx.Effect {
		var arr = container.ArrayFrom(av)
		return rx.CreateBlockingSequenceEffect(func(next func(rx.Object))(bool,rx.Object) {
			for i := uint(0); i < arr.Length; i += 1 {
				next(arr.GetItem(i))
			}
			return true, nil
		})
	},
	"take-one": func(e rx.Effect) rx.Effect {
		return e.TakeOne().Map(func(val rx.Object) rx.Object {
			var v = val.(struct { rx.Object; bool })
			if v.bool {
				return Just(v.Object)
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
}

