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

func (m *Machine) Call(f FunctionValue, arg Value) Value {
	return CallInNewThread(f, arg, m)
}

