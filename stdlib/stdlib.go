package stdlib

import (
	"math"
	"reflect"
	"math/cmplx"
)


/* IMPORTANT: this go file should be consistent with corresponding km files */
const Core = "Core"
var core_types = []string {
	Bit, Byte, Word, Dword, Qword, Number, Float, Int,
	Seq, Array, Heap, Set, Map,
	EffectMultiValue, Effect, NoExceptMultiValue, NoExcept,
	Int64, Uint64, Int32, Uint32, Int16, Uint16, Int8, Uint8,
	Complex, Char, Range, String, Bytes,
	Bool, Yes, No,
	Maybe, Just, Na,
	Result, Ok, Ng,
	Ordering, Smaller, Equal, Bigger,
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

const Bit = "Bit"
const Byte = "Byte"
const Word = "Word"
const Dword = "Dword"
const Qword = "Qword"
const Number = "Number"
const Int = "Int"
const Float = "Float"
const Seq = "Seq"
const Array = "Array"
const Heap = "Heap"
const Set = "Set"
const Map = "Map"
const EffectMultiValue = "Effect*"
const Effect = "Effect"
const NoExceptMultiValue = "NoExcept*"
const NoExcept = "NoExcept"
const Int64 = "Int64"
const Uint64 = "Uint64"
const Int32 = "Int32"
const Uint32 = "Uint32"
const Int16 = "Int16"
const Uint16 = "Uint16"
const Int8 = "Int8"
const Uint8 = "Uint8"
const Complex = "Complex"
const Char = "Char"
const Range = "Range"
const String = "String"
const Bytes = "Bytes"
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
