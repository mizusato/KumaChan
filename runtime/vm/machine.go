package vm

import (
	"os"
	"sync"
	"strings"
	"kumachan/rx"
	"kumachan/util"
	"kumachan/rpc/kmd"
	. "kumachan/lang"
	. "kumachan/util/error"
	"kumachan/runtime/lib/librpc"
)


const InitialDataStackCapacity = 16
const InitialCallStackCapacity = 4

type Machine struct {
	program         Program
	options         Options
	globalSlot      [] Value
	extraSlot       [] Value
	extraLock       *sync.Mutex
	contextPool     *sync.Pool
	scheduler       rx.Scheduler
	kmdTransformer  kmd.Transformer
}

type Options struct {
	Resources     map[string] util.Resource
	MaxStackSize  uint
	Environment   [] string
	Arguments     [] string
	DebugOptions
	StdIO
}

func Execute(p Program, opts Options, m_signal (chan <- *Machine)) {
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
		options:      opts,
		globalSlot:   nil,
		extraSlot:    make([] Value, 0),
		extraLock:    &sync.Mutex {},
		contextPool:  pool,
		scheduler:    sched,
	}
	m.kmdTransformer = librpc.KmdTransformer(m)
	if m_signal != nil {
		m_signal <- m
	}
	execute(p, m)
}

func (m *Machine) GetGlobalValue(index uint) (Value, bool) {
	var L = uint(len(m.globalSlot))
	if index < L {
		return m.globalSlot[index], true
	} else {
		m.extraLock.Lock()
		defer m.extraLock.Unlock()
		var offset = (index - L)
		if offset < uint(len(m.extraSlot)) {
			return m.extraSlot[offset], true
		} else {
			return nil, false
		}
	}
}

func (m *Machine) Call(f FunctionValue, arg Value) Value {
	return call(f, arg, m)
}

func (m *Machine) GetScheduler() rx.Scheduler {
	return m.scheduler
}

func (m *Machine) InjectExtraGlobals(values ([] Value)) {
	m.extraLock.Lock()
	defer m.extraLock.Unlock()
	m.extraSlot = append(m.extraSlot, values...)
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
	return FromBool(call(f.(FunctionValue), x, m).(SumValue))
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
		return h.machine.Call(f, arg)
	case NativeFunctionValue:
		return f(arg, h)
	default:
		panic("cannot call a non-callable value")
	}
}

func (h MachineContextHandle) GetScheduler() rx.Scheduler {
	return h.machine.scheduler
}

func (h MachineContextHandle) GetEnv() ([] string) {
	return h.machine.options.Environment
}

func (h MachineContextHandle) GetArgs() ([] string) {
	return h.machine.options.Arguments
}

func (h MachineContextHandle) GetStdIO() StdIO {
	return h.machine.options.StdIO
}

func (h MachineContextHandle) GetDebugOptions() DebugOptions {
	return h.machine.options.DebugOptions
}

func (h MachineContextHandle) GetErrorPoint() ErrorPoint {
	return GetFrameErrorPoint(h.context.workingFrame)
}

func (h MachineContextHandle) GetEntryModulePath() string {
	var raw = h.machine.program.MetaData.EntryModulePath
	return strings.TrimRight(raw, string([] rune { os.PathSeparator }))
}

func (h MachineContextHandle) KmdGetTypeFromId(id kmd.TypeId) *kmd.Type {
	return h.machine.program.KmdConfig.GetTypeFromId(id)
}

func (h MachineContextHandle) KmdSerialize(v Value, t *kmd.Type) ([] byte, error) {
	return librpc.KmdSerialize(v, t, h.machine.kmdTransformer)
}

func (h MachineContextHandle) KmdDeserialize(binary ([] byte), t *kmd.Type) (Value, error) {
	return librpc.KmdDeserialize(binary, t, h.machine.kmdTransformer)
}

func (h MachineContextHandle) GetResources(kind string) (map[string] util.Resource) {
	var res = make(map[string] util.Resource)
	for path, item := range h.machine.options.Resources {
		if item.Kind == kind {
			res[path] = item
		}
	}
	return res
}

