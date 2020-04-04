package lib

import (
	"math"
	"math/big"
)


func CheckFloat(x float64) float64 {
	if math.IsNaN(x) {
		panic("Float Overflow: NaN")
	}
	if math.IsInf(x, 0) {
		panic("Float Overflow: Infinity")
	}
	return x
}


var ArithmeticFunctions = map[string] interface{} {
	"+integer": func(a *big.Int, b *big.Int) *big.Int {
		var c big.Int
		return c.Add(a, b)
	},
	"-integer": func(a *big.Int, b *big.Int) *big.Int {
		var c big.Int
		return c.Sub(a, b)
	},
	"*integer": func(a *big.Int, b *big.Int) *big.Int {
		var c big.Int
		return c.Mul(a, b)
	},
	"quorem": func(a *big.Int, b *big.Int) (*big.Int, *big.Int) {
		var q, m big.Int
		q.QuoRem(a, b, &m)
		return &q, &m
	},
	"divmod": func(a *big.Int, b *big.Int) (*big.Int, *big.Int) {
		var q, m big.Int
		q.DivMod(a, b, &m)
		return &q, &m
	},
	"+float": func(a float64, b float64) float64 {
		return CheckFloat(a + b)
	},
	"-float":  func(a float64, b float64) float64 {
		return CheckFloat(a - b)
	},
	"*float": func(a float64, b float64) float64 {
		return CheckFloat(a * b)
	},
	"/float": func(a float64, b float64) float64 {
		return CheckFloat(a / b)
	},
	"%float": func(a float64, b float64) float64 {
		return CheckFloat(math.Mod(a, b))
	},
	"+int8": func(a int8, b int8) int8 {
		return a + b
	},
	"-int8": func(a int8, b int8) int8 {
		return a - b
	},
	"*int8": func(a int8, b int8) int8 {
		return a * b
	},
	"/int8": func(a int8, b int8) int8 {
		return a / b
	},
	"%int8": func(a int8, b int8) int8 {
		return a % b
	},
	"+uint8": func(a uint8, b uint8) uint8 {
		return a + b
	},
	"-uint8": func(a uint8, b uint8) uint8 {
		return a - b
	},
	"*uint8": func(a uint8, b uint8) uint8 {
		return a * b
	},
	"/uint8": func(a uint8, b uint8) uint8 {
		return a / b
	},
	"%uint8": func(a uint8, b uint8) uint8 {
		return a % b
	},
	"+int16": func(a int16, b int16) int16 {
		return a + b
	},
	"-int16": func(a int16, b int16) int16 {
		return a - b
	},
	"*int16": func(a int16, b int16) int16 {
		return a * b
	},
	"/int16": func(a int16, b int16) int16 {
		return a / b
	},
	"%int16": func(a int16, b int16) int16 {
		return a % b
	},
	"+uint16": func(a uint16, b uint16) uint16 {
		return a + b
	},
	"-uint16": func(a uint16, b uint16) uint16 {
		return a - b
	},
	"*uint16": func(a uint16, b uint16) uint16 {
		return a * b
	},
	"/uint16": func(a uint16, b uint16) uint16 {
		return a / b
	},
	"%uint16": func(a uint16, b uint16) uint16 {
		return a % b
	},
	"+int32": func(a int32, b int32) int32 {
		return a + b
	},
	"-int32": func(a int32, b int32) int32 {
		return a - b
	},
	"*int32": func(a int32, b int32) int32 {
		return a * b
	},
	"/int32": func(a int32, b int32) int32 {
		return a / b
	},
	"%int32": func(a int32, b int32) int32 {
		return a % b
	},
	"+uint32": func(a uint32, b uint32) uint32 {
		return a + b
	},
	"-uint32": func(a uint32, b uint32) uint32 {
		return a - b
	},
	"*uint32": func(a uint32, b uint32) uint32 {
		return a * b
	},
	"/uint32": func(a uint32, b uint32) uint32 {
		return a / b
	},
	"%uint32": func(a uint32, b uint32) uint32 {
		return a % b
	},
	"+int64": func(a int64, b int64) int64 {
		return a + b
	},
	"-int64": func(a int64, b int64) int64 {
		return a - b
	},
	"*int64": func(a int64, b int64) int64 {
		return a * b
	},
	"/int64": func(a int64, b int64) int64 {
		return a / b
	},
	"%int64": func(a int64, b int64) int64 {
		return a % b
	},
	"+uint64": func(a uint64, b uint64) uint64 {
		return a + b
	},
	"-uint64": func(a uint64, b uint64) uint64 {
		return a - b
	},
	"*uint64": func(a uint64, b uint64) uint64 {
		return a * b
	},
	"/uint64": func(a uint64, b uint64) uint64 {
		return a / b
	},
	"%uint64": func(a uint64, b uint64) uint64 {
		return a % b
	},
}
