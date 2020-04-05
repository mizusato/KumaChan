package common

import (
	. "kumachan/error"
	"reflect"
)


type Value = interface {}

type SumValue struct {
	Index  Short
	Value  Value
}

type ProductValue struct {
	Elements  [] Value
}

type FunctionValue struct {
	Underlying     *Function
	ContextValues  [] Value
}

type NativeFunctionValue  NativeFunction


func Inspect(v Value) ErrorMessage {
	var rv = reflect.ValueOf(v)
	var msg = make(ErrorMessage, 0)
	if rv.Type().AssignableTo(reflect.TypeOf([]rune{})) {
		msg.WriteText(TS_BOLD, "String")
		msg.WriteEndText(TS_NORMAL, string(v.([]rune)))
	} else {
		// TODO: more fancy representations
		msg.WriteText(TS_NORMAL, rv.String())
	}
	return msg
}

func BoolFrom(p SumValue) bool {
	// should be consistent with `stdlib/core.km`
	if p.Value != nil { panic("something went wrong") }
	if p.Index == 0 {
		return true
	} else if p.Index == 1 {
		return false
	} else {
		panic("something went wrong")
	}
}

func ToBool(p bool) SumValue {
	// should be consistent with `stdlib/core.km`
	if p == true {
		return SumValue { Index: 0 }
	} else {
		return SumValue { Index: 1 }
	}
}

func ToOrdering(o Ordering) SumValue {
	// should be consistent with `stdlib/core.km`
	switch o {
	case Smaller:
		return SumValue { Index: 0 }
	case Equal:
		return SumValue { Index: 1 }
	case Bigger:
		return SumValue { Index: 2 }
	default:
		panic("impossible branch")
	}
}

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
	default:
		panic("invalid Qword")
	}
}
