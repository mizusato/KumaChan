package api

import . "kumachan/lang"


var AssertionFunctions = map[string] interface{} {
	"panic": func(msg string) struct{} {
		panic("programmed panic: " + msg)
	},
	"assert": func(ok_ SumValue, k Value, h InteropContext) Value {
		var ok = FromBool(ok_)
		if ok {
			return h.Call(k, nil)
		} else {
			panic("assertion failed")
		}
	},
	"assert-some": func(maybe_v SumValue, k Value, h InteropContext) Value {
		var v, ok = Unwrap(maybe_v)
		if ok {
			return h.Call(k, v)
		} else {
			panic("assertion failed")
		}
	},
}

