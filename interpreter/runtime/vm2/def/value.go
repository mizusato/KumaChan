package def

import "kumachan/stdlib"

type Value = interface {}

type EnumValue = *ValEnum
type ValEnum struct {
	Index  ShortIndex
	Value  Value
}

type TupleValue = *ValTup
type ValTup struct {
	Elements  [] Value
}

type UsualFuncValue = *ValFunc
type ValFunc struct {
	Entity   *FunctionEntity
	Context  AddrSpace
}

type NativeFuncValue = *NativeFunction
func ValNativeFunc(f NativeFunction) NativeFuncValue {
	return &f
}

func Tuple(elements... Value) TupleValue {
	return &ValTup { elements }
}

func TupleOf(elements ([] Value)) TupleValue {
	return &ValTup { elements }
}

func Some(v Value) EnumValue {
	return &ValEnum {
		Index: stdlib.SomeIndex,
		Value: v,
	}
}

func None() EnumValue {
	return &ValEnum {
		Index: stdlib.NoneIndex,
		Value: nil,
	}
}

func Unwrap(maybe EnumValue) (Value, bool) {
	if maybe.Index == stdlib.SomeIndex {
		return maybe.Value, true
	} else {
		return nil, false
	}
}

