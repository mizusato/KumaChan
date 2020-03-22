package vm

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/common/rx"
	"sync"
)

const InitialDataStackCapacity = 16
const InitialCallStackCapacity = 4

type Machine struct {
	Program       Program
	GlobalSlot    [] Value
	ContextPool   *sync.Pool
	EventLoop     *rx.EventLoop
	MaxStackSize  uint
}

func CreateMachine(p Program, max_stack_size uint) *Machine {
	var m = &Machine {
		Program:    p,
		GlobalSlot: nil,
		ContextPool:  &sync.Pool { New: func() interface{} {
			return &ExecutionContext {
				DataStack: make([]Value, 0, InitialDataStackCapacity),
				CallStack: make([]CallStackFrame, 0, InitialCallStackCapacity),
			}
		} },
		EventLoop:    rx.SpawnEventLoop(),
		MaxStackSize: max_stack_size,
	}
	Execute(p, m)
	return m
}

type ExecutionContext struct {
	DataStack     DataStack
	CallStack     CallStack
	WorkingFrame  CallStackFrame
}

type DataStack  [] Value
type CallStack  [] CallStackFrame
type CallStackFrame struct {
	Function  *Function
	BaseAddr  uint
	InstPtr   uint
}

func (m *Machine) Call(fv Value, arg Value) Value {
	switch f := fv.(type) {
	case FunctionValue:
		return CallFunction(f, arg, m)
	case NativeFunctionValue:
		return f(arg, m)
	default:
		panic("cannot call a non-callable value")
	}
}

func (m *Machine) CallAsync(fv Value, arg Value, cb func(Value)) {
	go (func() {
		var ret = m.Call(fv, arg)
		cb(ret)
	})()
}
