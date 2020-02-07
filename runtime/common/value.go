package common

type Value interface { RuntimeValue() }

func (impl PlainValue) RuntimeValue() {}
type PlainValue struct {
	Pointer interface{}
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
