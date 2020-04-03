package common


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


func Bool(v bool) SumValue {
	// This function should be consistent with `stdlib/core.km`
	if v == true {
		return SumValue { Index: 0 }
	} else {
		return SumValue { Index: 1 }
	}
}
