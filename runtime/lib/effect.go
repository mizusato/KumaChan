package lib

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/common/rx"
	"kumachan/runtime/lib/container"
)


var EffectFunctions = map[string] Value {
	"NoExcept* from Range": func(l uint, r uint) rx.Effect {
		return rx.CreateBlockingEffect(func(next func(rx.Object)) error {
			for i := l; i < r; i += 1 {
				next(i)
			}
			return nil
		})
	},
	"NoExcept* from Seq": func(seq container.Seq) rx.Effect {
		return rx.CreateBlockingEffect(func(next func(rx.Object)) error {
			for item, rest, ok := seq.Next(); ok; item, rest, ok = rest.Next() {
				next(item)
			}
			return nil
		})
	},
	"NoExcept* from Array": func(av Value) rx.Effect {
		var arr = container.ArrayFrom(av)
		return rx.CreateBlockingEffect(func(next func(rx.Object)) error {
			for i := uint(0); i < arr.Length; i += 1 {
				next(arr.GetItem(i))
			}
			return nil
		})
	},
	"oneshot": func(v Value) rx.Effect {
		return rx.CreateBlockingEffect(func(next func(rx.Object)) error {
			next(v)
			return nil
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
	"then": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.SingleThen(func(val rx.Object) rx.Effect {
			return h.Call(f, val).(rx.Effect)
		})
	},
	"then-shortcut": func(a rx.Effect, b rx.Effect) rx.Effect {
		return a.SingleThen(func(_ rx.Object) rx.Effect {
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
	"map-effect": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.Map(func(val rx.Object) rx.Object {
			return h.Call(f, val)
		})
	},
	"filter-effect": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.Filter(func(val rx.Object) bool {
			return BoolFrom((h.Call(f, val)).(SumValue))
		})
	},
	"reduce-effect": func(e rx.Effect, opts ProductValue, h MachineHandle) rx.Effect {
		var init, f = Tuple2From(opts)
		return e.Reduce(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
	},
	"scan-effect": func(e rx.Effect, opts ProductValue, h MachineHandle) rx.Effect {
		var init, f = Tuple2From(opts)
		return e.Scan(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
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
	"mix-map": func(e rx.Effect, opts ProductValue, h MachineHandle) rx.Effect {
		var c, f = Tuple2From(opts)
		return e.MixMap(func(val rx.Object) rx.Effect {
			return h.Call(f, val).(rx.Effect)
		}, c.(uint))
	},
}

