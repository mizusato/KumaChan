package lib

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/common/rx"
)


var EffectFunctions = map[string] Value {
	"effect-then-preset": func(a rx.Effect, b rx.Effect) rx.Effect {
		// TODO: implementation is not good, refactor this function
		return a.SwitchMap(func(_ rx.Object) rx.Effect {
			return b
		})
	},
	"effect-catch": func(e rx.Effect, f Value, h MachineHandle) rx.Effect {
		return e.Catch(func(rx.Object) rx.Effect {
			return h.Call(f, nil).(rx.Effect)
		})
	},
}

