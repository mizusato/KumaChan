package vm

import . "kumachan/runtime/common"

type Machine struct {
	Program      Program
	GlobalValues [] Value
}

type Thread struct {
	Machine       *Machine
	DataStack     DataStack
	CallStack     CallStack
	WorkingFrame  CallStackFrame
}

type DataStack  [] Value
type CallStack  [] CallStackFrame
type CallStackFrame struct {
	Function  *Function
	BaseAddr  int
	InstPtr   int
}

func (m *Machine) Call(fv Value, arg Value) Value {
	switch f := fv.(type) {
	case FunctionValue:
		return CallInNewThread(f, arg, m)
	case NativeFunctionValue:
		return f(arg, m)
	default:
		panic("cannot call a non-callable value")
	}
}

