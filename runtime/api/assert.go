package api

import . "kumachan/lang"


var AssertionFunctions = map[string] interface{} {
	"panic": func(msg String) struct{} {
		panic("programmed panic: " + GoStringFromString(msg))
	},
	"assert": func(ok_ SumValue, k Value, h InteropContext) Value {
		var ok = FromBool(ok_)
		if ok {
			return h.Call(k, nil)
		} else {
			panic("assertion failed")
		}
	},
}

