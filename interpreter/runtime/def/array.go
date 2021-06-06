package def

import (
	"reflect"
	"math/big"
)


var compactArrayTypes = [] reflect.Type {
	reflect.TypeOf(big.NewInt(0)),
	reflect.TypeOf(float64(0)),
	reflect.TypeOf(complex128(complex(0,1))),
	reflect.TypeOf(uint8(0)),
	reflect.TypeOf(int32(0)),
}

var compactArrayTypeMap = (func() map[reflect.Type] ShortIndex {
	if !(uint(len(compactArrayTypes)) < ShortSizeMax) {
		panic("too many compact array types")
	}
	var m = make(map[reflect.Type] ShortIndex)
	for i, t := range compactArrayTypes {
		m[t] = ShortIndex(i)
	}
	return m
})()

func GetCompactArrayType(idx ShortIndex) reflect.Type {
	return compactArrayTypes[idx]
}

func GetCompactArrayTypeIndex(t reflect.Type) (ShortIndex, bool) {
	var idx, exists = compactArrayTypeMap[t]
	return idx, exists
}


