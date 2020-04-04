package lib

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/lib/container"
	"math/big"
)


var ComparisonFunctions = map[string] interface{} {
	"=string": func(a []rune, b []rune) SumValue {
		return CoreBool(container.StringCompare(a, b) == Equal)
	},
	"<string": func(a []rune, b []rune) SumValue {
		return CoreBool(container.StringCompare(a, b) == Smaller)
	},
	"<>string": func(a []rune, b []rune) SumValue {
		return CoreOrdering(container.StringCompare(a, b))
	},
	"=integer": func(a *big.Int, b *big.Int) SumValue {
		return CoreBool(a.Cmp(b) == 0)
	},
	"<integer": func(a *big.Int, b *big.Int) SumValue {
		return CoreBool(a.Cmp(b) == -1)
	},
	"<>integer": func(a *big.Int, b *big.Int) SumValue {
		var result = a.Cmp(b)
		if result < 0 {
			return CoreOrdering(Smaller)
		} else if result > 0 {
			return CoreOrdering(Bigger)
		} else {
			return CoreOrdering(Equal)
		}
	},
	"=float": func(a float64, b float64) SumValue {
		return CoreBool(a == b)
	},
	"<float": func(a float64, b float64) SumValue {
		return CoreBool(a < b)
	},
	"<>float": func(a float64, b float64) SumValue {
		if a < b {
			return CoreOrdering(Smaller)
		} else if a > b {
			return CoreOrdering(Bigger)
		} else {
			return CoreOrdering(Equal)
		}
	},
	"=int8": func(a int8, b int8) SumValue {
		return CoreBool(a == b)
	},
	"<int8": func(a int8, b int8) SumValue {
		return CoreBool(a < b)
	},
	"<>int8": func(a int8, b int8) SumValue {
		if a < b {
			return CoreOrdering(Smaller)
		} else if a > b {
			return CoreOrdering(Bigger)
		} else {
			return CoreOrdering(Equal)
		}
	},
	"=uint8": func(a uint8, b uint8) SumValue {
		return CoreBool(a == b)
	},
	"<uint8": func(a uint8, b uint8) SumValue {
		return CoreBool(a < b)
	},
	"<>uint8": func(a uint8, b uint8) SumValue {
		if a < b {
			return CoreOrdering(Smaller)
		} else if a > b {
			return CoreOrdering(Bigger)
		} else {
			return CoreOrdering(Equal)
		}
	},
	"=int16": func(a int16, b int16) SumValue {
		return CoreBool(a == b)
	},
	"<int16": func(a int16, b int16) SumValue {
		return CoreBool(a < b)
	},
	"<>int16": func(a int16, b int16) SumValue {
		if a < b {
			return CoreOrdering(Smaller)
		} else if a > b {
			return CoreOrdering(Bigger)
		} else {
			return CoreOrdering(Equal)
		}
	},
	"=uint16": func(a uint16, b uint16) SumValue {
		return CoreBool(a == b)
	},
	"<uint16": func(a uint16, b uint16) SumValue {
		return CoreBool(a < b)
	},
	"<>uint16": func(a uint16, b uint16) SumValue {
		if a < b {
			return CoreOrdering(Smaller)
		} else if a > b {
			return CoreOrdering(Bigger)
		} else {
			return CoreOrdering(Equal)
		}
	},
	"=int32": func(a int32, b int32) SumValue {
		return CoreBool(a == b)
	},
	"<int32": func(a int32, b int32) SumValue {
		return CoreBool(a < b)
	},
	"<>int32": func(a int32, b int32) SumValue {
		if a < b {
			return CoreOrdering(Smaller)
		} else if a > b {
			return CoreOrdering(Bigger)
		} else {
			return CoreOrdering(Equal)
		}
	},
	"=uint32": func(a uint32, b uint32) SumValue {
		return CoreBool(a == b)
	},
	"<uint32": func(a uint32, b uint32) SumValue {
		return CoreBool(a < b)
	},
	"<>uint32": func(a uint32, b uint32) SumValue {
		if a < b {
			return CoreOrdering(Smaller)
		} else if a > b {
			return CoreOrdering(Bigger)
		} else {
			return CoreOrdering(Equal)
		}
	},
	"=int64": func(a int64, b int64) SumValue {
		return CoreBool(a == b)
	},
	"<int64": func(a int64, b int64) SumValue {
		return CoreBool(a < b)
	},
	"<>int64": func(a int64, b int64) SumValue {
		if a < b {
			return CoreOrdering(Smaller)
		} else if a > b {
			return CoreOrdering(Bigger)
		} else {
			return CoreOrdering(Equal)
		}
	},
	"=uint64": func(a uint64, b uint64) SumValue {
		return CoreBool(a == b)
	},
	"<uint64": func(a uint64, b uint64) SumValue {
		return CoreBool(a < b)
	},
	"<>uint64": func(a uint64, b uint64) SumValue {
		if a < b {
			return CoreOrdering(Smaller)
		} else if a > b {
			return CoreOrdering(Bigger)
		} else {
			return CoreOrdering(Equal)
		}
	},
}