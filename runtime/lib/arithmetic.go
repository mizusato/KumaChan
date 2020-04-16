package lib

import (
	"kumachan/stdlib"
	"math"
	"math/big"
)


var ArithmeticFunctions = map[string] interface{} {
	"+Int": func(a *big.Int, b *big.Int) *big.Int {
		var c big.Int
		return c.Add(a, b)
	},
	"-Int": func(a *big.Int, b *big.Int) *big.Int {
		var c big.Int
		return c.Sub(a, b)
	},
	"*Int": func(a *big.Int, b *big.Int) *big.Int {
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
	"+Number": func(a uint, b uint) uint {
		var r = a + b
		if r < a || r < b {
			panic("Number Overflow")
		}
		return r
	},
	"-Number": func(a uint, b uint) uint {
		if a < b {
			panic("Number Underflow")
		}
		return a - b
	},
	"*Number": func(a uint, b uint) uint {
		var r = a * b
		if r < a || r < b {
			panic("Number Overflow")
		} else if (r == a && b != 1) || (r == b && a != 1) {
			panic("Number Overflow")
		}
		return r
	},
	"/Number": func(a uint, b uint) uint {
		return a / b
	},
	"%Number": func(a uint, b uint) uint {
		return a % b
	},
	"+Float": func(a float64, b float64) float64 {
		return stdlib.CheckFloat(a + b)
	},
	"-Float":  func(a float64, b float64) float64 {
		return stdlib.CheckFloat(a - b)
	},
	"*Float": func(a float64, b float64) float64 {
		return stdlib.CheckFloat(a * b)
	},
	"/Float": func(a float64, b float64) float64 {
		return stdlib.CheckFloat(a / b)
	},
	"%Float": func(a float64, b float64) float64 {
		return stdlib.CheckFloat(math.Mod(a, b))
	},
	"+Int8": func(a int8, b int8) int8 {
		return a + b
	},
	"-Int8": func(a int8, b int8) int8 {
		return a - b
	},
	"*Int8": func(a int8, b int8) int8 {
		return a * b
	},
	"/Int8": func(a int8, b int8) int8 {
		return a / b
	},
	"%Int8": func(a int8, b int8) int8 {
		return a % b
	},
	"+Uint8": func(a uint8, b uint8) uint8 {
		return a + b
	},
	"-Uint8": func(a uint8, b uint8) uint8 {
		return a - b
	},
	"*Uint8": func(a uint8, b uint8) uint8 {
		return a * b
	},
	"/Uint8": func(a uint8, b uint8) uint8 {
		return a / b
	},
	"%Uint8": func(a uint8, b uint8) uint8 {
		return a % b
	},
	"+Int16": func(a int16, b int16) int16 {
		return a + b
	},
	"-Int16": func(a int16, b int16) int16 {
		return a - b
	},
	"*Int16": func(a int16, b int16) int16 {
		return a * b
	},
	"/Int16": func(a int16, b int16) int16 {
		return a / b
	},
	"%Int16": func(a int16, b int16) int16 {
		return a % b
	},
	"+Uint16": func(a uint16, b uint16) uint16 {
		return a + b
	},
	"-Uint16": func(a uint16, b uint16) uint16 {
		return a - b
	},
	"*Uint16": func(a uint16, b uint16) uint16 {
		return a * b
	},
	"/Uint16": func(a uint16, b uint16) uint16 {
		return a / b
	},
	"%Uint16": func(a uint16, b uint16) uint16 {
		return a % b
	},
	"+Int32": func(a int32, b int32) int32 {
		return a + b
	},
	"-Int32": func(a int32, b int32) int32 {
		return a - b
	},
	"*Int32": func(a int32, b int32) int32 {
		return a * b
	},
	"/Int32": func(a int32, b int32) int32 {
		return a / b
	},
	"%Int32": func(a int32, b int32) int32 {
		return a % b
	},
	"+Uint32": func(a uint32, b uint32) uint32 {
		return a + b
	},
	"-Uint32": func(a uint32, b uint32) uint32 {
		return a - b
	},
	"*Uint32": func(a uint32, b uint32) uint32 {
		return a * b
	},
	"/Uint32": func(a uint32, b uint32) uint32 {
		return a / b
	},
	"%Uint32": func(a uint32, b uint32) uint32 {
		return a % b
	},
	"+Int64": func(a int64, b int64) int64 {
		return a + b
	},
	"-Int64": func(a int64, b int64) int64 {
		return a - b
	},
	"*Int64": func(a int64, b int64) int64 {
		return a * b
	},
	"/Int64": func(a int64, b int64) int64 {
		return a / b
	},
	"%Int64": func(a int64, b int64) int64 {
		return a % b
	},
	"+Uint64": func(a uint64, b uint64) uint64 {
		return a + b
	},
	"-Uint64": func(a uint64, b uint64) uint64 {
		return a - b
	},
	"*Uint64": func(a uint64, b uint64) uint64 {
		return a * b
	},
	"/Uint64": func(a uint64, b uint64) uint64 {
		return a / b
	},
	"%Uint64": func(a uint64, b uint64) uint64 {
		return a % b
	},
	"floor": func(x float64) float64 {
		return math.Floor(x)
	},
	"ceil": func(x float64) float64 {
		return math.Ceil(x)
	},
	"round": func(x float64) float64 {
		return math.Round(x)
	},
	"real-**": func(x float64, p float64) float64 {
		if x == 0.0 && p == 0.0 {
			panic("cannot evaluate 0.0 to the power of 0.0")
		}
		return math.Pow(x, p)
	},
	"real-sqrt": func(x float64) float64 {
		if x >= 0.0 {
			return math.Sqrt(x)
		} else {
			panic("cannot take real square root on negative number")
		}
	},
	"real-cbrt": func(x float64) float64 {
		return math.Cbrt(x)
	},
	"real-exp": func(x float64) float64 {
		return math.Exp(x)
	},
	"real-log": func(x float64) float64 {
		if x > 0 {
			return math.Log(x)
		} else {
			panic("cannot take real logarithm of non-positive number")
		}
	},
	"real-sin": func(x float64) float64 {
		return math.Sin(x)
	},
	"real-cos": func(x float64) float64 {
		return math.Cos(x)
	},
	"real-tan": func(x float64) float64 {
		var sine = math.Sin(x)
		var cosine = math.Cos(x)
		if cosine == 0.0 {
			panic("input out of the domain of tangent function")
		}
		return (sine / cosine)
	},
	"real-asin": func(x float64) float64 {
		if -1.0 <= x && x <= 1.0 {
			return math.Asin(x)
		} else {
			panic("input out of the domain of inverse sine function")
		}
	},
	"real-acos": func(x float64) float64 {
		if -1.0 <= x && x <= 1.0 {
			return math.Acos(x)
		} else {
			panic("input out of the domain of inverse cosine function")
		}
	},
	"real-atan": func(x float64) float64 {
		return math.Atan(x)
	},
	"atan2": func(y float64, x float64) float64 {
		return math.Atan2(y, x)
	},
	// complex arithmetic operations are implemented in `core.km` (non-native)
}
