package lib

import . "kumachan/runtime/common"

var LogicFunctions = map[string] interface{} {
	"and": func(p SumValue, q SumValue) SumValue {
		return ToBool(BoolFrom(p) && BoolFrom(q))
	},
	"or": func(p SumValue, q SumValue) SumValue {
		return ToBool(BoolFrom(p) || BoolFrom(q))
	},
	"not": func(p SumValue) SumValue {
		return ToBool(!(BoolFrom(p)))
	},
}
