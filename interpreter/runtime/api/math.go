package api

import (
	"math"
	"math/big"
	"math/cmplx"
	"kumachan/standalone/util"
	. "kumachan/interpreter/base"
)


var MathFunctions = map[string] interface{} {
	"Number?": func(n *big.Int) EnumValue {
		if util.IsNonNegative(n) {
			return Some(n)
		} else {
			return None()
		}
	},
	"NormalFloat?": func(x float64) EnumValue {
		if util.IsNormalFloat(x) {
			return Some(x)
		} else {
			return None()
		}
	},
	"NormalComplex?": func(z complex128) EnumValue {
		if util.IsNormalComplex(z) {
			return Some(z)
		} else {
			return None()
		}
	},
	// basic
	"i+i": func(a *big.Int, b *big.Int) *big.Int {
		var c big.Int
		return c.Add(a, b)
	},
	"i-i": func(a *big.Int, b *big.Int) *big.Int {
		var c big.Int
		return c.Sub(a, b)
	},
	"-i": func(n *big.Int) *big.Int {
		var m big.Int
		m.Neg(n)
		return &m
	},
	"i*i": func(a *big.Int, b *big.Int) *big.Int {
		var c big.Int
		return c.Mul(a, b)
	},
	"i/i": func(a *big.Int, b *big.Int) *big.Int {
		var q, r big.Int
		q.QuoRem(a, b, &r)
		return &q
	},
	"i%i": func(a *big.Int, b *big.Int) *big.Int {
		var q, r big.Int
		q.QuoRem(a, b, &r)
		return &r
	},
	"i**i": func(a *big.Int, b *big.Int) *big.Int {
		var c big.Int
		c.Exp(a, b, nil)
		return &c
	},
	"modexp": func(a *big.Int, b *big.Int, M *big.Int) *big.Int {
		var c big.Int
		c.Exp(a, b, M)
		return &c
	},
	"quorem": func(a *big.Int, b *big.Int) (*big.Int, *big.Int) {
		var q, r big.Int
		q.QuoRem(a, b, &r)
		return &q, &r
	},
	"divmod": func(a *big.Int, b *big.Int) (*big.Int, *big.Int) {
		var q, r big.Int
		q.DivMod(a, b, &r)
		return &q, &r
	},
	"n-!n": func(a *big.Int, b *big.Int) *big.Int {
		var c big.Int
		c.Sub(a, b)
		if c.Cmp(big.NewInt(0)) >= 0 {
			return &c
		} else {
			panic("number subtraction underflow")
		}
	},
	"f+f": func(a float64, b float64) float64 {
		return a + b
	},
	"f-f":  func(a float64, b float64) float64 {
		return a - b
	},
	"-f": func(x float64) float64 {
		return -x
	},
	"f*f": func(a float64, b float64) float64 {
		return a * b
	},
	"f/f": func(a float64, b float64) float64 {
		return a / b
	},
	"f%f": func(a float64, b float64) float64 {
		return math.Mod(a, b)
	},
	"f**f": math.Pow,
	// advanced
	"floor":      math.Floor,
	"ceil":       math.Ceil,
	"round":      math.Round,
	"float-sqrt": math.Sqrt,
	"float-cbrt": math.Cbrt,
	"float-exp":  math.Exp,
	"float-log":  math.Log,
	"float-sin":  math.Sin,
	"float-cos":  math.Cos,
	"float-tan":  math.Tan,
	"float-asin": math.Asin,
	"float-acos": math.Acos,
	"float-atan": math.Atan,
	"atan2":     math.Atan2,
	// complex
	"Complex": func(re float64, im float64) complex128 {
		return complex(re, im)
	},
	"FloatComplex": func(re float64, im float64) complex128 {
		return complex(re, im)
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
		return (z1 + z2)
	},
	"f+c": func(f float64, z complex128) complex128 {
		return (complex(f, 0) + z)
	},
	"c+f": func(z complex128, f float64) complex128 {
		return (z + complex(f, 0))
	},
	"c-c": func(z1 complex128, z2 complex128) complex128 {
		return (z1 - z2)
	},
	"f-c": func(f float64, z complex128) complex128 {
		return (complex(f, 0) - z)
	},
	"c-f": func(z complex128, f float64) complex128 {
		return (z - complex(f, 0))
	},
	"-c": func(z complex128) complex128 {
		return -z
	},
	"c*c": func(z1 complex128, z2 complex128) complex128 {
		return (z1 * z2)
	},
	"f*c": func(f float64, z complex128) complex128 {
		return (complex(f, 0) * z)
	},
	"c*f": func(z complex128, f float64) complex128 {
		return (z * complex(f, 0))
	},
	"c/c": func(z1 complex128, z2 complex128) complex128 {
		return (z1 / z2)
	},
	"f/c": func(f float64, z complex128) complex128 {
		return (complex(f, 0) / z)
	},
	"c/f": func(z complex128, f float64) complex128 {
		return (z / complex(f, 0))
	},
	"complex-exp":  cmplx.Exp,
	"complex-log":  cmplx.Log,
	"complex-sin":  cmplx.Sin,
	"complex-cos":  cmplx.Cos,
	"complex-tan":  cmplx.Tan,
	"complex-asin": cmplx.Asin,
	"complex-acos": cmplx.Acos,
	"complex-atan": cmplx.Atan,
	"complex-sqrt": func(z complex128) complex128 {
		var norm, arg = cmplx.Polar(z)
		return cmplx.Rect(math.Sqrt(norm), (arg / 2.0))
	},
	"complex-cbrt": func(z complex128) complex128 {
		var norm, arg = cmplx.Polar(z)
		return cmplx.Rect(math.Cbrt(norm), (arg / 3.0))
	},
}

