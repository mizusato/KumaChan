package lib

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/common/rx"
)


var EffectFunctions = map[string] Value {
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
	"reduce-effect": func(e rx.Effect, f Value, init Value, h MachineHandle) rx.Effect {
		return e.Reduce(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
	},
	"scan-effect": func(e rx.Effect, f Value, init Value, h MachineHandle) rx.Effect {
		return e.Scan(func(acc rx.Object, val rx.Object) rx.Object {
			return h.Call(f, ToTuple2(acc, val))
		}, init)
	},
}

