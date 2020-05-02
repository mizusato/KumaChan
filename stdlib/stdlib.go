package stdlib

import (
	"math"
	"reflect"
	"math/cmplx"
)


/* IMPORTANT: this go file should be consistent with corresponding km files */
const Core = "Core"
var core_types = []string {
	// types.km
	Float, Number,
	Int64, Uint64, Int32, Uint32, Int16, Uint16, Int8, Uint8,
	Bool, Yes, No,
	Maybe, Just, Na,
	Result, Ok, Ng,
	Ordering, Smaller, Equal, Bigger,
	// binary.km
	Bit, Byte, Word, Dword, Qword, Bytes,
	// int.km
	Int,
	// containers.km
	Seq, Array, Heap, Set, Map,
	// effect.km
	EffectMultiValue, Effect, NoExceptMultiValue, NoExcept,
	// time.km
	Time,
	// complex.km
	Complex,
	// string.km
	Char, String,
	// range.km
	Range,
	// debugging.km
	Debug, Never,
}
// var core_constants = []string {}
func GetCoreScopedSymbols() []string {
	var list = make([]string, 0)
	list = append(list, core_types...)
	// Using public constants in Core violates shadowing rules
	// list = append(list, core_constants...)
	return list
}

// types.km
const Float = "Float"
const Number = "Number"
const Int64 = "Int64"
const Uint64 = "Uint64"
const Int32 = "Int32"
const Uint32 = "Uint32"
const Int16 = "Int16"
const Uint16 = "Uint16"
const Int8 = "Int8"
const Uint8 = "Uint8"
const Bool = "Bool"
const Yes = "Yes"
const No = "No"
const ( YesIndex = iota; NoIndex )
const Maybe = "Maybe"
const Just = "Just"
const Na = "N/A"
const ( JustIndex = iota; NaIndex )
const Result = "Result"
const Ok = "OK"
const Ng = "NG"
const ( OkIndex = iota; NgIndex )
const Ordering = "Ordering"
const Smaller = "<<"
const Equal = "=="
const Bigger = ">>"
const ( SmallerIndex = iota; EqualIndex; BiggerIndex )
// binary.km
const Bit = "Bit"
const Byte = "Byte"
const Word = "Word"
const Dword = "Dword"
const Qword = "Qword"
const Bytes = "Bytes"
// int.km
const Int = "Int"
// containers.km
const Seq = "Seq"
const Array = "Array"
const Heap = "Heap"
const Set = "Set"
const Map = "Map"
// effect.km
const EffectMultiValue = "Effect*"
const Effect = "Effect"
const NoExceptMultiValue = "NoExcept*"
const NoExcept = "NoExcept"
// time.km
const Time = "Time"
// complex.km
const Complex = "Complex"
// string.km
const Char = "Char"
const String = "String"
// range.km
const Range = "Range"
// debugging.km
const Debug = "Debug"
const Never = "Never"

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

func GetPrimitiveReflectType(name string) (reflect.Type, bool) {
	switch name {
	case Number:
		return reflect.TypeOf(uint(0)), true
	case Float:
		return reflect.TypeOf(float64(0)), true
	case Bit:
		return reflect.TypeOf(true), true
	case Uint8, Byte:
		return reflect.TypeOf(uint8(0)), true
	case Uint16, Word:
		return reflect.TypeOf(uint16(0)), true
	case Uint32, Dword, Char:
		return reflect.TypeOf(uint32(0)), true
	case Uint64, Qword:
		return reflect.TypeOf(uint64(0)), true
	case Int8:
		return reflect.TypeOf(int8(0)), true
	case Int16:
		return reflect.TypeOf(int16(0)), true
	case Int32:
		return reflect.TypeOf(int32(0)), true
	case Int64:
		return reflect.TypeOf(int64(0)), true
	case Complex:
		return reflect.TypeOf(complex128(complex(0,1))), true
	default:
		return nil, false
	}
}
