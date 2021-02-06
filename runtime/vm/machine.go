package vm

import (
	"sync"
	"kumachan/rx"
	"kumachan/util"
	"kumachan/rpc/kmd"
	"kumachan/runtime/lib/librpc"
	. "kumachan/lang"
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

