package common

import (
	"math"
	"math/big"
)

type Value interface { RuntimeValue() }

func (impl PlainValue) RuntimeValue() {}
type PlainValue struct {
	Inline   uint64
	Pointer  interface{}
}

func (impl SumValue) RuntimeValue() {}
type SumValue struct {
	Index Short
	Value Value
}

func (impl ProductValue) RuntimeValue() {}
type ProductValue struct {
	Elements  [] Value
}

func (impl FunctionValue) RuntimeValue() {}
type FunctionValue struct {
	Underlying     *Function
	ContextValues  [] Value
}

func Tuple2(a Value, b Value) Value {
	return ProductValue { Elements: []Value { a, b } }
}

func FromTuple2(v Value) (Value, Value) {
	var tuple = v.(ProductValue)
	return tuple.Elements[0], tuple.Elements[1]
}

func BitValue(p bool) Value {
	if p {
		return PlainValue { Inline: 1 }
	} else {
		return PlainValue { Inline: 0 }
	}
}
func Int8Value(n int8) Value {
	return PlainValue { Inline: uint64(uint8(n)) }
}
func Uint8Value(n uint8) Value {
	return PlainValue { Inline: uint64(n) }
}
func Int16Value(n int16) Value {
	return PlainValue { Inline: uint64(uint16(n)) }
}
func Uint16Value(n uint16) Value {
	return PlainValue { Inline: uint64(n) }
}
func Int32Value(n int32) Value {
	return PlainValue { Inline: uint64(uint32(n)) }
}
func Uint32Value(n uint32) Value {
	return PlainValue { Inline: uint64(n) }
}
func Int64Value(n int64) Value {
	return PlainValue { Inline: uint64(n) }
}
func Uint64Value(n uint64) Value {
	return PlainValue { Inline: n }
}
func IntValue(n *big.Int) Value {
	return PlainValue { Pointer: n }
}
func Float64Value(x float64) Value {
	return PlainValue { Inline: math.Float64bits(x) }
}
func FloatValue(x float64) Value {
	if math.IsNaN(x) {
		panic("Float Overflow: NaN")
	}
	if math.IsInf(x, 0) {
		panic("Float Overflow: Infinity")
	}
	return Float64Value(x)
}

func BitFrom(v Value) bool {
	return (v.(PlainValue).Inline & 1 != 0)
}
func Int8From(v Value) int8 {
	return int8(v.(PlainValue).Inline)
}
func Uint8From(v Value) uint8 {
	return uint8(v.(PlainValue).Inline)
}
func Int16From(v Value) int16 {
	return int16(v.(PlainValue).Inline)
}
func Uint16From(v Value) uint16 {
	return uint16(v.(PlainValue).Inline)
}
func Int32From(v Value) int32 {
	return int32(v.(PlainValue).Inline)
}
func Uint32From(v Value) uint32 {
	return uint32(v.(PlainValue).Inline)
}
func Int64From(v Value) int64 {
	return int64(v.(PlainValue).Inline)
}
func Uint64From(v Value) uint64 {
	return v.(PlainValue).Inline
}
func IntFrom(v Value) *big.Int {
	return v.(PlainValue).Pointer.(*big.Int)
}
func Float64From(v Value) float64 {
	return math.Float64frombits(v.(PlainValue).Inline)
}
func FloatFrom(v Value) float64 {
	var x = Float64From(v)
	if math.IsNaN(x) || math.IsInf(x, 0) {
		panic("bad Float value")
	}
	return x
}
