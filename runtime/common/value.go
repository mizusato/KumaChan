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
