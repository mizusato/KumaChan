package api

import (
	"math"
	"math/big"
	"math/cmplx"
	"kumachan/misc/util"
	"kumachan/lang"
)


var MathFunctions = map[string] interface{} {
	"Float to Maybe[Real]": func(x float64) lang.SumValue {
		if util.IsValidReal(x) {
			return lang.Some(x)
		} else {
			return lang.None()
		}
	},
	"FloatComplex to Maybe[Complex]": func(z complex128) lang.SumValue {
		if util.IsValidComplex(z) {
			return lang.Some(z)
		} else {
			return lang.None()
		}
	},
	"Number from Uint32": func(n uint32) uint {
		return uint(n)
	},
	"Number from Uint16": func(n uint16) uint {
		return uint(n)
	},
	"Number from Uint8": func(n uint8) uint {
		return uint(n)
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
	"+Real": func(a float64, b float64) float64 {
		return util.CheckReal(a + b)
	},
	"-Real":  func(a float64, b float64) float64 {
		return util.CheckReal(a - b)
	},
	"*Real": func(a float64, b float64) float64 {
		return util.CheckReal(a * b)
	},
	"/Real": func(a float64, b float64) float64 {
		return util.CheckReal(a / b)
	},
	"%Real": func(a float64, b float64) float64 {
		return util.CheckReal(math.Mod(a, b))
	},
	"+Float": func(a float64, b float64) float64 {
		return a + b
	},
	"-Float":  func(a float64, b float64) float64 {
		return a - b
	},
	"*Float": func(a float64, b float64) float64 {
		return a * b
	},
	"/Float": func(a float64, b float64) float64 {
		return a / b
	},
	"%Float": func(a float64, b float64) float64 {
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
		return util.CheckReal(math.Floor(x))
	},
	"ceil": func(x float64) float64 {
		return util.CheckReal(math.Ceil(x))
	},
	"round": func(x float64) float64 {
		return util.CheckReal(math.Round(x))
	},
	"real-**": func(x float64, p float64) float64 {
		if x == 0.0 && p == 0.0 {
			panic("cannot evaluate 0.0 to the power of 0.0")
		}
		return util.CheckReal(math.Pow(x, p))
	},
	"real-sqrt": func(x float64) float64 {
		if x >= 0.0 {
			return util.CheckReal(math.Sqrt(x))
		} else {
			panic("cannot take real square root on negative number")
		}
	},
	"real-cbrt": func(x float64) float64 {
		return util.CheckReal(math.Cbrt(x))
	},
	"real-exp": func(x float64) float64 {
		return util.CheckReal(math.Exp(x))
	},
	"real-log": func(x float64) float64 {
		if x > 0 {
			return util.CheckReal(math.Log(x))
		} else {
			panic("cannot take real logarithm of non-positive number")
		}
	},
	"real-sin": func(x float64) float64 {
		return util.CheckReal(math.Sin(x))
	},
	"real-cos": func(x float64) float64 {
		return util.CheckReal(math.Cos(x))
	},
	"real-tan": func(x float64) float64 {
		return util.CheckReal(math.Tan(x))
	},
	"real-asin": func(x float64) float64 {
		if -1.0 <= x && x <= 1.0 {
			return util.CheckReal(math.Asin(x))
		} else {
			panic("input out of the domain of inverse sine function")
		}
	},
	"real-acos": func(x float64) float64 {
		if -1.0 <= x && x <= 1.0 {
			return util.CheckReal(math.Acos(x))
		} else {
			panic("input out of the domain of inverse cosine function")
		}
	},
	"real-atan": func(x float64) float64 {
		return util.CheckReal(math.Atan(x))
	},
	"atan2": func(y float64, x float64) float64 {
		return util.CheckReal(math.Atan2(y, x))
	},
	"[f] floor":     math.Floor,
	"[f] ceil":      math.Ceil,
	"[f] round":     math.Round,
	"[f] real-**":   math.Pow,
	"[f] real-sqrt": math.Sqrt,
	"[f] real-cbrt": math.Cbrt,
	"[f] real-exp":  math.Exp,
	"[f] real-log":  math.Log,
	"[f] real-sin":  math.Sin,
	"[f] real-cos":  math.Cos,
	"[f] real-tan":  math.Tan,
	"[f] real-asin": math.Asin,
	"[f] real-acos": math.Acos,
	"[f] real-atan": math.Atan,
	"atan2*":     math.Atan2,
	// complex
	"Complex": func(re float64, im float64) complex128 {
		return complex(re, im)
	},
	"FloatComplex": func(re float64, im float64) complex128 {
		return complex(re, im)
	},
	"ComplexPolar": func(norm float64, arg float64) complex128 {
		if norm >= 0 {
			return complex((norm * math.Cos(arg)), (norm * math.Sin(arg)))
		} else {
			panic("negative norm")
		}
	},
	"FloatComplexPolar": func(norm float64, arg float64) complex128 {
		if norm >= 0 {
			return complex((norm * math.Cos(arg)), (norm * math.Sin(arg)))
		} else {
			return cmplx.NaN()
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
	"c+c": func(z1 complex128, z2 complex128) complex128 {
		return util.CheckComplex(z1 + z2)
	},
	"r+c": func(f float64, z complex128) complex128 {
		return util.CheckComplex(complex(f, 0) + z)
	},
	"c+r": func(z complex128, f float64) complex128 {
		return util.CheckComplex(z + complex(f, 0))
	},
	"c-c": func(z1 complex128, z2 complex128) complex128 {
		return util.CheckComplex(z1 - z2)
	},
	"r-c": func(f float64, z complex128) complex128 {
		return util.CheckComplex(complex(f, 0) - z)
	},
	"c-r": func(z complex128, f float64) complex128 {
		return util.CheckComplex(z - complex(f, 0))
	},
	"c*c": func(z1 complex128, z2 complex128) complex128 {
		return util.CheckComplex(z1 * z2)
	},
	"r*c": func(f float64, z complex128) complex128 {
		return util.CheckComplex(complex(f, 0) * z)
	},
	"c*r": func(z complex128, f float64) complex128 {
		return util.CheckComplex(z * complex(f, 0))
	},
	"c/c": func(z1 complex128, z2 complex128) complex128 {
		return util.CheckComplex(z1 / z2)
	},
	"r/c": func(f float64, z complex128) complex128 {
		return util.CheckComplex(complex(f, 0) / z)
	},
	"c/r": func(z complex128, f float64) complex128 {
		return util.CheckComplex(z / complex(f, 0))
	},
	"[f] c+c": func(z1 complex128, z2 complex128) complex128 {
		return (z1 + z2)
	},
	"[f] r+c": func(f float64, z complex128) complex128 {
		return (complex(f, 0) + z)
	},
	"[f] c+r": func(z complex128, f float64) complex128 {
		return (z + complex(f, 0))
	},
	"[f] c-c": func(z1 complex128, z2 complex128) complex128 {
		return (z1 - z2)
	},
	"[f] r-c": func(f float64, z complex128) complex128 {
		return (complex(f, 0) - z)
	},
	"[f] c-r": func(z complex128, f float64) complex128 {
		return (z - complex(f, 0))
	},
	"[f] c*c": func(z1 complex128, z2 complex128) complex128 {
		return (z1 * z2)
	},
	"[f] r*c": func(f float64, z complex128) complex128 {
		return (complex(f, 0) * z)
	},
	"[f] c*r": func(z complex128, f float64) complex128 {
		return (z * complex(f, 0))
	},
	"[f] c/c": func(z1 complex128, z2 complex128) complex128 {
		return (z1 / z2)
	},
	"[f] r/c": func(f float64, z complex128) complex128 {
		return (complex(f, 0) / z)
	},
	"[f] c/r": func(z complex128, f float64) complex128 {
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
	"[f] complex-exp":  cmplx.Exp,
	"[f] complex-log":  cmplx.Log,
	"[f] complex-sin":  cmplx.Sin,
	"[f] complex-cos":  cmplx.Cos,
	"[f] complex-tan":  cmplx.Tan,
	"[f] complex-asin": cmplx.Asin,
	"[f] complex-acos": cmplx.Acos,
	"[f] complex-atan": cmplx.Atan,
	"[f] complex-sqrt": func(z complex128) complex128 {
		var norm, arg = cmplx.Polar(z)
		return cmplx.Rect(math.Sqrt(norm), (arg / 2.0))
	},
	"[f] complex-cbrt": func(z complex128) complex128 {
		var norm, arg = cmplx.Polar(z)
		return cmplx.Rect(math.Cbrt(norm), (arg / 3.0))
	},
}

