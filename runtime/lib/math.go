package lib

import (
	"math"
	"math/big"
	"math/cmplx"
	"kumachan/util"
	"kumachan/runtime/common"
)


var MathFunctions = map[string] interface{} {
	"Float* from Float":     func(x float64)    float64    { return x },
	"Complex* from Complex": func(z complex128) complex128 { return z },
	"check-nan-inf-float": func(x float64) common.SumValue {
		if util.IsValidFloat(x) {
			return common.Just(x)
		} else {
			return common.Na()
		}
	},
	"check-nan-inf-complex": func(z complex128) common.SumValue {
		if util.IsValidComplex(z) {
			return common.Just(z)
		} else {
			return common.Na()
		}
	},
	// arithmetic
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
		return util.CheckFloat(a + b)
	},
	"-Float":  func(a float64, b float64) float64 {
		return util.CheckFloat(a - b)
	},
	"*Float": func(a float64, b float64) float64 {
		return util.CheckFloat(a * b)
	},
	"/Float": func(a float64, b float64) float64 {
		return util.CheckFloat(a / b)
	},
	"%Float": func(a float64, b float64) float64 {
		return util.CheckFloat(math.Mod(a, b))
	},
	"+Float!": func(a float64, b float64) float64 {
		return a + b
	},
	"-Float!":  func(a float64, b float64) float64 {
		return a - b
	},
	"*Float!": func(a float64, b float64) float64 {
		return a * b
	},
	"/Float!": func(a float64, b float64) float64 {
		return a / b
	},
	"%Float!": func(a float64, b float64) float64 {
		return math.Mod(a, b)
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
	// advanced
	"floor": func(x float64) float64 {
		return util.CheckFloat(math.Floor(x))
	},
	"ceil": func(x float64) float64 {
		return util.CheckFloat(math.Ceil(x))
	},
	"round": func(x float64) float64 {
		return util.CheckFloat(math.Round(x))
	},
	"real-**": func(x float64, p float64) float64 {
		if x == 0.0 && p == 0.0 {
			panic("cannot evaluate 0.0 to the power of 0.0")
		}
		return util.CheckFloat(math.Pow(x, p))
	},
	"real-sqrt": func(x float64) float64 {
		if x >= 0.0 {
			return util.CheckFloat(math.Sqrt(x))
		} else {
			panic("cannot take real square root on negative number")
		}
	},
	"real-cbrt": func(x float64) float64 {
		return util.CheckFloat(math.Cbrt(x))
	},
	"real-exp": func(x float64) float64 {
		return util.CheckFloat(math.Exp(x))
	},
	"real-log": func(x float64) float64 {
		if x > 0 {
			return util.CheckFloat(math.Log(x))
		} else {
			panic("cannot take real logarithm of non-positive number")
		}
	},
	"real-sin": func(x float64) float64 {
		return util.CheckFloat(math.Sin(x))
	},
	"real-cos": func(x float64) float64 {
		return util.CheckFloat(math.Cos(x))
	},
	"real-tan": func(x float64) float64 {
		return util.CheckFloat(math.Tan(x))
	},
	"real-asin": func(x float64) float64 {
		if -1.0 <= x && x <= 1.0 {
			return util.CheckFloat(math.Asin(x))
		} else {
			panic("input out of the domain of inverse sine function")
		}
	},
	"real-acos": func(x float64) float64 {
		if -1.0 <= x && x <= 1.0 {
			return util.CheckFloat(math.Acos(x))
		} else {
			panic("input out of the domain of inverse cosine function")
		}
	},
	"real-atan": func(x float64) float64 {
		return util.CheckFloat(math.Atan(x))
	},
	"atan2": func(y float64, x float64) float64 {
		return util.CheckFloat(math.Atan2(y, x))
	},
	"floor!":     math.Floor,
	"ceil!":      math.Ceil,
	"round!":     math.Round,
	"real-**!":   math.Pow,
	"real-sqrt!": math.Sqrt,
	"real-cbrt!": math.Cbrt,
	"real-exp!":  math.Exp,
	"real-log!":  math.Log,
	"real-sin!":  math.Sin,
	"real-cos!":  math.Cos,
	"real-tan!":  math.Tan,
	"real-asin!": math.Asin,
	"real-acos!": math.Acos,
	"real-atan!": math.Atan,
	"atan2!":     math.Atan2,
	// complex
	"complex": func(re float64, im float64) complex128 {
		return complex(re, im)
	},
	"polar": func(norm float64, arg float64) complex128 {
		if norm >= 0 {
			return complex((norm * math.Cos(arg)), (norm * math.Sin(arg)))
		} else {
			panic("negative norm")
		}
	},
	"<real>": func(z complex128) float64 {
		return real(z)
	},
	"<imag>": func(z complex128) float64 {
		return imag(z)
	},
	"<conj>": func(z complex128) complex128 {
		var re, im = real(z), imag(z)
		return complex(re, -im)
	},
	"<norm>": func(z complex128) float64 {
		var re, im = real(z), imag(z)
		return math.Sqrt((re * re) + (im * im))
	},
	"<arg>": func(z complex128) float64 {
		var re, im = real(z), imag(z)
		return math.Atan2(im, re)
	},
	"+complex": func(z1 complex128, z2 complex128) complex128 {
		return util.CheckComplex(z1 + z2)
	},
	"f+complex": func(f float64, z complex128) complex128 {
		return util.CheckComplex(complex(f, 0) + z)
	},
	"complex+f": func(z complex128, f float64) complex128 {
		return util.CheckComplex(z + complex(f, 0))
	},
	"-complex": func(z1 complex128, z2 complex128) complex128 {
		return util.CheckComplex(z1 - z2)
	},
	"f-complex": func(f float64, z complex128) complex128 {
		return util.CheckComplex(complex(f, 0) - z)
	},
	"complex-f": func(z complex128, f float64) complex128 {
		return util.CheckComplex(z - complex(f, 0))
	},
	"*complex": func(z1 complex128, z2 complex128) complex128 {
		return util.CheckComplex(z1 * z2)
	},
	"f*complex": func(f float64, z complex128) complex128 {
		return util.CheckComplex(complex(f, 0) * z)
	},
	"complex*f": func(z complex128, f float64) complex128 {
		return util.CheckComplex(z * complex(f, 0))
	},
	"/complex": func(z1 complex128, z2 complex128) complex128 {
		return util.CheckComplex(z1 / z2)
	},
	"f/complex": func(f float64, z complex128) complex128 {
		return util.CheckComplex(complex(f, 0) / z)
	},
	"complex/f": func(z complex128, f float64) complex128 {
		return util.CheckComplex(z / complex(f, 0))
	},
	"+complex!": func(z1 complex128, z2 complex128) complex128 {
		return (z1 + z2)
	},
	"f+complex!": func(f float64, z complex128) complex128 {
		return (complex(f, 0) + z)
	},
	"complex+f!": func(z complex128, f float64) complex128 {
		return (z + complex(f, 0))
	},
	"-complex!": func(z1 complex128, z2 complex128) complex128 {
		return (z1 - z2)
	},
	"f-complex!": func(f float64, z complex128) complex128 {
		return (complex(f, 0) - z)
	},
	"complex-f!": func(z complex128, f float64) complex128 {
		return (z - complex(f, 0))
	},
	"*complex!": func(z1 complex128, z2 complex128) complex128 {
		return (z1 * z2)
	},
	"f*complex!": func(f float64, z complex128) complex128 {
		return (complex(f, 0) * z)
	},
	"complex*f!": func(z complex128, f float64) complex128 {
		return (z * complex(f, 0))
	},
	"/complex!": func(z1 complex128, z2 complex128) complex128 {
		return (z1 / z2)
	},
	"f/complex!": func(f float64, z complex128) complex128 {
		return (complex(f, 0) / z)
	},
	"complex/f!": func(z complex128, f float64) complex128 {
		return (z / complex(f, 0))
	},
	"complex-exp": func(z complex128) complex128 {
		return util.CheckComplex(cmplx.Exp(z))
	},
	"complex-log": func(z complex128) complex128 {
		return util.CheckComplex(cmplx.Log(z))
	},
	"complex-sin": func(z complex128) complex128 {
		return util.CheckComplex(cmplx.Sin(z))
	},
	"complex-cos": func(z complex128) complex128 {
		return util.CheckComplex(cmplx.Cos(z))
	},
	"complex-tan": func(z complex128) complex128 {
		return util.CheckComplex(cmplx.Tan(z))
	},
	"complex-asin": func(z complex128) complex128 {
		return util.CheckComplex(cmplx.Asin(z))
	},
	"complex-acos": func(z complex128) complex128 {
		return util.CheckComplex(cmplx.Acos(z))
	},
	"complex-atan": func(z complex128) complex128 {
		return util.CheckComplex(cmplx.Atan(z))
	},
	"complex-sqrt": func(z complex128) complex128 {
		var norm, arg = cmplx.Polar(z)
		return util.CheckComplex(cmplx.Rect(math.Sqrt(norm), (arg / 2.0)))
	},
	"complex-cbrt": func(z complex128) complex128 {
		var norm, arg = cmplx.Polar(z)
		return util.CheckComplex(cmplx.Rect(math.Cbrt(norm), (arg / 3.0)))
	},
	"complex-exp!":  cmplx.Exp,
	"complex-log!":  cmplx.Log,
	"complex-sin!":  cmplx.Sin,
	"complex-cos!":  cmplx.Cos,
	"complex-tan!":  cmplx.Tan,
	"complex-asin!": cmplx.Asin,
	"complex-acos!": cmplx.Acos,
	"complex-atan!": cmplx.Atan,
	"complex-sqrt!": func(z complex128) complex128 {
		var norm, arg = cmplx.Polar(z)
		return cmplx.Rect(math.Sqrt(norm), (arg / 2.0))
	},
	"complex-cbrt!": func(z complex128) complex128 {
		var norm, arg = cmplx.Polar(z)
		return cmplx.Rect(math.Cbrt(norm), (arg / 3.0))
	},
}

