package lib

import (
	"fmt"
	"math/big"
	. "kumachan/runtime/common"
)


var StringFunctions = map[string] Value {
	"String from Int": func(n *big.Int) []rune {
		return []rune(n.String())
	},
	"String from Float": func(x float64) []rune {
		return []rune(fmt.Sprint(x))
	},
	"String from Int8": func(n int8) []rune {
		return []rune(fmt.Sprint(n))
	},
	"String from Int16": func(n int16) []rune {
		return []rune(fmt.Sprint(n))
	},
	"String from Int32": func(n int32) []rune {
		return []rune(fmt.Sprint(n))
	},
	"String from Int64": func(n int64) []rune {
		return []rune(fmt.Sprint(n))
	},
	"String from Uint8": func(n uint8) []rune {
		return []rune(fmt.Sprint(n))
	},
	"String from Uint16": func(n uint16) []rune {
		return []rune(fmt.Sprint(n))
	},
	"String from Uint32": func(n uint32) []rune {
		return []rune(fmt.Sprint(n))
	},
	"String from Uint64": func(n uint64) []rune {
		return []rune(fmt.Sprint(n))
	},
	// TODO: split, join, contains, find, trim, substr-view, substr-copy,
	//       +String
}
