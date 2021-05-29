package util

import (
	"math"
	"math/big"
	"math/cmplx"
)


func GetNumberUint(small uint) *big.Int {
	var n big.Int
	n.SetUint64(uint64(small))
	return &n
}

func GetNumberUint64(small uint64) *big.Int {
	var n big.Int
	n.SetUint64(small)
	return &n
}

func GetUintNumber(n *big.Int) uint {
	if n.Cmp(big.NewInt(0)) < 0 { panic("something went wrong") }
	var limit big.Int
	limit.SetUint64(uint64(^uint(0)))
	if n.Cmp(&limit) <= 0 {
		return uint(n.Uint64())
	} else {
		panic("given number too big")
	}
}

func GetInt64Integer(n *big.Int) int64 {
	if n.IsInt64() {
		return n.Int64()
	} else {
		panic("given number too big")
	}
}

func IsNonNegative(n *big.Int) bool {
	return (n.Cmp(big.NewInt(0)) >= 0)
}

func IsNormalFloat(x float64) bool {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return false
	}
	return true
}

func IsNormalComplex(z complex128) bool {
	if cmplx.IsNaN(z) || cmplx.IsInf(z) {
		return false
	}
	return true
}

