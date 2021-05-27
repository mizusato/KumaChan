package api

import (
	. "kumachan/lang"
	. "kumachan/runtime/lib/container"
	"math/big"
)


var ComparisonFunctions = map[string] interface{} {
	"sum-index-equal": func(a EnumValue, b EnumValue) EnumValue {
		return ToBool(a.Index == b.Index)
	},
	"=String": func(a string, b string) EnumValue {
		return ToBool(StringCompare(a, b) == Equal)
	},
	"<String": func(a string, b string) EnumValue {
		return ToBool(StringCompare(a, b) == Smaller)
	},
	"<>String": func(a string, b string) EnumValue {
		return ToOrdering(StringCompare(a, b))
	},
	"=Integer": func(a *big.Int, b *big.Int) EnumValue {
		return ToBool(a.Cmp(b) == 0)
	},
	"<Integer": func(a *big.Int, b *big.Int) EnumValue {
		return ToBool(a.Cmp(b) == -1)
	},
	"<>Integer": func(a *big.Int, b *big.Int) EnumValue {
		var result = a.Cmp(b)
		if result < 0 {
			return ToOrdering(Smaller)
		} else if result > 0 {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
	"<NormalFloat": func(a float64, b float64) EnumValue {
		return ToBool(a < b)
	},
}