package util

import (
	"math"
	"math/cmplx"
)


func CheckReal(x float64) float64 {
	if math.IsNaN(x) {
		panic("Real Overflow: NaN")
	}
	if math.IsInf(x, 0) {
		panic("Real Overflow: Infinity")
	}
	return x
}

func IsValidReal(x float64) bool {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return false
	}
	return true
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

func IsValidComplex(z complex128) bool {
	if cmplx.IsNaN(z) || cmplx.IsInf(z) {
		return false
	}
	return true
}

