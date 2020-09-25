package util

import (
	"math"
	"math/cmplx"
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

func CheckComplex(z complex128) complex128 {
	if cmplx.IsNaN(z) {
		panic("Complex Overflow: NaN")
	}
	if cmplx.IsInf(z) {
		panic("Complex Overflow: Infinity")
	}
	return z
}

