package vm

import (
	. "kumachan/runtime/common"
	"kumachan/runtime/common/rx"
	"sync"
)


const InitialDataStackCapacity = 16
const InitialCallStackCapacity = 4

type Machine struct {
	program       Program
	globalSlot    [] Value
	contextPool   *sync.Pool
	scheduler     rx.Scheduler
	maxStackSize  uint
}

func Execute(p Program, max_stack_size uint) *Machine {
	var sched = rx.TrivialScheduler {
		EventLoop: rx.SpawnEventLoop(),
	}
	var pool = &sync.Pool { New: func() interface{} {
		return &ExecutionContext {
			dataStack: make([]Value, 0, InitialDataStackCapacity),
			callStack: make([]CallStackFrame, 0, InitialCallStackCapacity),
		}
	} }
	var m = &Machine {
		program:      p,
		globalSlot:   nil,
		contextPool:  pool,
		scheduler:    sched,
		maxStackSize: max_stack_size,
	}
	execute(p, m)
	return m
}

type ExecutionContext struct {
	dataStack     DataStack
	callStack     CallStack
	workingFrame  CallStackFrame
}

type DataStack  [] Value
type CallStack  [] CallStackFrame
type CallStackFrame struct {
	function  *Function
	baseAddr  uint
	instPtr   uint
}

func (m *Machine) Call(fv Value, arg Value) Value {
	switch f := fv.(type) {
	case FunctionValue:
		return call(f, arg, m)
	case NativeFunctionValue:
		return f(arg, m)
	default:
		panic("cannot call a non-callable value")
	}
}
