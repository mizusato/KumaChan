package lib

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/lib/container"
	"math/big"
)


var ComparisonFunctions = map[string] interface{} {
	"=String": func(a []rune, b []rune) SumValue {
		return ToBool(container.StringCompare(a, b) == Equal)
	},
	"<String": func(a []rune, b []rune) SumValue {
		return ToBool(container.StringCompare(a, b) == Smaller)
	},
	"<>String": func(a []rune, b []rune) SumValue {
		return ToOrdering(container.StringCompare(a, b))
	},
	"=Int": func(a *big.Int, b *big.Int) SumValue {
		return ToBool(a.Cmp(b) == 0)
	},
	"<Int": func(a *big.Int, b *big.Int) SumValue {
		return ToBool(a.Cmp(b) == -1)
	},
	"<>Int": func(a *big.Int, b *big.Int) SumValue {
		var result = a.Cmp(b)
		if result < 0 {
			return ToOrdering(Smaller)
		} else if result > 0 {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
	"=Size": func(a uint, b uint) SumValue {
		return ToBool(a == b)
	},
	"<Size": func(a uint, b uint) SumValue {
		return ToBool(a < b)
	},
	"<>Size": func(a uint, b uint) SumValue {
		if a < b {
			return ToOrdering(Smaller)
		} else if a > b {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
	"=Float": func(a float64, b float64) SumValue {
		return ToBool(a == b)
	},
	"<Float": func(a float64, b float64) SumValue {
		return ToBool(a < b)
	},
	"<>Float": func(a float64, b float64) SumValue {
		if a < b {
			return ToOrdering(Smaller)
		} else if a > b {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
	"=Int8": func(a int8, b int8) SumValue {
		return ToBool(a == b)
	},
	"<Int8": func(a int8, b int8) SumValue {
		return ToBool(a < b)
	},
	"<>Int8": func(a int8, b int8) SumValue {
		if a < b {
			return ToOrdering(Smaller)
		} else if a > b {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
	"=Uint8": func(a uint8, b uint8) SumValue {
		return ToBool(a == b)
	},
	"<Uint8": func(a uint8, b uint8) SumValue {
		return ToBool(a < b)
	},
	"<>Uint8": func(a uint8, b uint8) SumValue {
		if a < b {
			return ToOrdering(Smaller)
		} else if a > b {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
	"=Int16": func(a int16, b int16) SumValue {
		return ToBool(a == b)
	},
	"<Int16": func(a int16, b int16) SumValue {
		return ToBool(a < b)
	},
	"<>Int16": func(a int16, b int16) SumValue {
		if a < b {
			return ToOrdering(Smaller)
		} else if a > b {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
	"=Uint16": func(a uint16, b uint16) SumValue {
		return ToBool(a == b)
	},
	"<Uint16": func(a uint16, b uint16) SumValue {
		return ToBool(a < b)
	},
	"<>Uint16": func(a uint16, b uint16) SumValue {
		if a < b {
			return ToOrdering(Smaller)
		} else if a > b {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
	"=Int32": func(a int32, b int32) SumValue {
		return ToBool(a == b)
	},
	"<Int32": func(a int32, b int32) SumValue {
		return ToBool(a < b)
	},
	"<>Int32": func(a int32, b int32) SumValue {
		if a < b {
			return ToOrdering(Smaller)
		} else if a > b {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
	"=Uint32": func(a uint32, b uint32) SumValue {
		return ToBool(a == b)
	},
	"<Uint32": func(a uint32, b uint32) SumValue {
		return ToBool(a < b)
	},
	"<>Uint32": func(a uint32, b uint32) SumValue {
		if a < b {
			return ToOrdering(Smaller)
		} else if a > b {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
	"=Int64": func(a int64, b int64) SumValue {
		return ToBool(a == b)
	},
	"<Int64": func(a int64, b int64) SumValue {
		return ToBool(a < b)
	},
	"<>Int64": func(a int64, b int64) SumValue {
		if a < b {
			return ToOrdering(Smaller)
		} else if a > b {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
	"=Uint64": func(a uint64, b uint64) SumValue {
		return ToBool(a == b)
	},
	"<Uint64": func(a uint64, b uint64) SumValue {
		return ToBool(a < b)
	},
	"<>Uint64": func(a uint64, b uint64) SumValue {
		if a < b {
			return ToOrdering(Smaller)
		} else if a > b {
			return ToOrdering(Bigger)
		} else {
			return ToOrdering(Equal)
		}
	},
}