package stdlib

import "math"


/* IMPORTANT: this go file should be consistent with corresponding km files */
const Core = "Core"
var core_types = []string {
	Bit, Byte, Word, Dword, Qword, Int,
	Bytes, String, Seq, Array, Heap, Set, Map,
	EffectMultiValue, Effect,
	Int64, Uint64, Int32, Uint32, Int16, Uint16, Int8, Uint8,
	Char, Nat, Float64, Float,
	Bool, Yes, No,
	Maybe, Just, Na,
	Result, Ok, Ng,
	Ordering, Smaller, Equal, Bigger,
	Debug, Never,
}
func GetCoreTypes() []string {
	return core_types
}

const Bit = "Bit"
const Byte = "Byte"
const Word = "Word"
const Dword = "Dword"
const Qword = "Qword"
const Int = "Int"
const Bytes = "Bytes"
const String = "String"
const Seq = "Seq"
const Array = "Array"
const Heap = "Heap"
const Set = "Set"
const Map = "Map"
const EffectMultiValue = "Effect*"
const Effect = "Effect"
const Int64 = "Int64"
const Uint64 = "Uint64"
const Int32 = "Int32"
const Uint32 = "Uint32"
const Int16 = "Int16"
const Uint16 = "Uint16"
const Int8 = "Int8"
const Uint8 = "Uint8"
const Char = "Char"
const Nat = "Nat"
const Float64 = "Float64"
const Float = "Float"
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

func ByteFrom(i interface{}) uint8 {
	switch x := i.(type) {
	case uint8:
		return x
	case int8:
		return uint8(x)
	default:
		panic("invalid Byte")
	}
}

func WordFrom(i interface{}) uint16 {
	switch x := i.(type) {
	case uint16:
		return x
	case int16:
		return uint16(x)
	default:
		panic("invalid Word")
	}
}

func DwordFrom(i interface{}) uint32 {
	switch x := i.(type) {
	case uint32:
		return x
	case int32:
		return uint32(x)
	default:
		panic("invalid Dword")
	}
}

func QwordFrom(i interface{}) uint64 {
	switch x := i.(type) {
	case uint64:
		return x
	case int64:
		return uint64(x)
	case float64:
		return math.Float64bits(x)
	default:
		panic("invalid Qword")
	}
}
