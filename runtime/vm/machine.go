package vm

import (
	. "kumachan/error"
	. "kumachan/runtime/common"
	"kumachan/runtime/rx"
	"sync"
	"kumachan/kmd"
	"kumachan/runtime/lib"
)


const InitialDataStackCapacity = 16
const InitialCallStackCapacity = 4

type Machine struct {
	program         Program
	arguments       [] string
	globalSlot      [] Value
	contextPool     *sync.Pool
	scheduler       rx.Scheduler
	maxStackSize    uint
	kmdTransformer  kmd.Transformer
}

func Execute(p Program, args ([] string), max_stack_size uint) *Machine {
	var sched = rx.TrivialScheduler {
		EventLoop: rx.SpawnEventLoop(),
	}
	var pool = &sync.Pool { New: func() interface{} {
		return &ExecutionContext {
			dataStack: make([] Value, 0, InitialDataStackCapacity),
			callStack: make([] CallStackFrame, 0, InitialCallStackCapacity),
		}
	} }
	var m = &Machine {
		program:      p,
		arguments:    args,
		globalSlot:   nil,
		contextPool:  pool,
		scheduler:    sched,
		maxStackSize: max_stack_size,
	}
	m.kmdTransformer = lib.KmdTransformer(m)
	execute(p, m)
	return m
}

type ExecutionContext struct {
	dataStack     DataStack
	callStack     CallStack
	workingFrame  CallStackFrame
	indexBufLen   uint
	indexBuf      [ProductMaxSize] Short
}

type DataStack  [] Value
type CallStack  [] CallStackFrame
type CallStackFrame struct {
	function  *Function
	baseAddr  uint
	instPtr   uint
}

type MachineContextHandle struct {
	machine  *Machine
	context  *ExecutionContext
}

func (h MachineContextHandle) Call(fv Value, arg Value) Value {
	switch f := fv.(type) {
	case FunctionValue:
		return call(f, arg, h.machine)
	case NativeFunctionValue:
		return f(arg, h)
	default:
		panic("cannot call a non-callable value")
	}
}

func (h MachineContextHandle) GetScheduler() rx.Scheduler {
	return h.machine.scheduler
}

func (h MachineContextHandle) GetArgs() ([] string) {
	return h.machine.arguments
}

func (h MachineContextHandle) GetErrorPoint() ErrorPoint {
	return GetFrameErrorPoint(h.context.workingFrame)
}

func (h MachineContextHandle) KmdSerialize(v Value, t *kmd.Type) ([] byte, error) {
	return lib.KmdSerialize(v, t, h.machine.kmdTransformer)
}

func (h MachineContextHandle) KmdDeserialize(binary ([] byte), t *kmd.Type) (Value, error) {
	return lib.KmdDeserialize(binary, t, h.machine.kmdTransformer)
}

func (m *Machine) KmdGetConfig() KmdConfig {
	return m.program.KmdConfig
}

func (m *Machine) KmdGetAdapter(index uint) Value {
	var adapter = m.globalSlot[index]
	var _, ok = adapter.(FunctionValue)
	if !(ok) { panic("something went wrong") }
	return adapter
}

func (m *Machine) KmdCallAdapter(f Value, x Value) Value {
	return call(f.(FunctionValue), x, m)
}

func (m *Machine) KmdCallValidator(f Value, x Value) bool {
	return BoolFrom(call(f.(FunctionValue), x, m).(SumValue))
}
