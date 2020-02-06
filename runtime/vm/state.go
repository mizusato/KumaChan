package vm

type GlobalState struct {
	Program       Program
	GlobalValues  [] Value
}

type CallState struct {
	GlobalState  *GlobalState
	CallStack    CallStack
	DataStack    DataStack
	Function     *Function
	BaseAddr     int
	InstPtr      int
}

type CallStack  [] CallStackFrame

type CallStackFrame struct {
	SavedFunction  *Function
	SavedBaseAddr  int
	SavedInstPtr   int
}

type DataStack  [] Value



