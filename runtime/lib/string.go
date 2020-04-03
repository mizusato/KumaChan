package lib

import (
	"math/big"
	. "kumachan/runtime/common"
)


var StringFunctions = map[string] Value {
	"str-integer": func(n *big.Int) []rune {
		return []rune(n.String())
	},
}
