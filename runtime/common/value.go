package common

import (
	. "kumachan/error"
	"kumachan/stdlib"
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

func Tuple2From(t ProductValue) (Value, Value) {
	if len(t.Elements) != 2 { panic("tuple size is not 2") }
	return t.Elements[0], t.Elements[1]
}

func ToTuple2(a Value, b Value) ProductValue {
	return ProductValue { [] Value { a, b } }
}

func SingleValueFromBundle(b ProductValue) Value {
	if len(b.Elements) != 1 { panic("bundle size is not 1") }
	return b.Elements[0]
}

func BoolFrom(p SumValue) bool {
	if p.Value != nil { panic("something went wrong") }
	if p.Index == stdlib.YesIndex {
		return true
	} else if p.Index == stdlib.NoIndex {
		return false
	} else {
		panic("something went wrong")
	}
}

func ToBool(p bool) SumValue {
	if p == true {
		return SumValue { Index: stdlib.YesIndex }
	} else {
		return SumValue { Index: stdlib.NoIndex }
	}
}

func ToOrdering(o Ordering) SumValue {
	switch o {
	case Smaller:
		return SumValue { Index: stdlib.SmallerIndex }
	case Equal:
		return SumValue { Index: stdlib.EqualIndex }
	case Bigger:
		return SumValue { Index: stdlib.BiggerIndex }
	default:
		panic("impossible branch")
	}
}

func Just(v Value) SumValue {
	return SumValue {
		Index: stdlib.JustIndex,
		Value: v,
	}
}

func Na() SumValue {
	return SumValue {
		Index: stdlib.NaIndex,
		Value: nil,
	}
}

func ByteFrom(v Value) uint8 {
	return stdlib.ByteFrom(v)
}

func WordFrom(v Value) uint16 {
	return stdlib.WordFrom(v)
}

func DwordFrom(v Value) uint32 {
	return stdlib.DwordFrom(v)
}

func QwordFrom(v Value) uint64 {
	return stdlib.QwordFrom(v)
}
