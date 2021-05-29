package vm

import (
	"sync"
	"kumachan/misc/rx"
	"kumachan/runtime/lib/librpc"
	. "kumachan/lang"
)


const InitialDataStackCapacity = 16
const InitialCallStackCapacity = 4

type Machine struct {
	program      Program
	options      Options
	globalSlot   [] Value
	extraSlot    [] Value
	extraLock    *sync.Mutex
	contextPool  *sync.Pool
	scheduler    rx.Scheduler
	GeneratedObjects
}

type Options struct {
	Resources     map[string] Resource
	MaxStackSize  uint
	Environment   [] string
	Arguments     [] string
	DebugOptions
	StdIO
}

type GeneratedObjects struct {
	kmdApi     KmdApi
	rpcApi     RpcApi
	resources  map[string] map[string] Resource  // kind -> path -> res
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
	m.generateObjects()
	if m_signal != nil {
		m_signal <- m
	}
	execute(p, m)
}

func (m *Machine) generateObjects() {
	m.GeneratedObjects = GeneratedObjects {
		kmdApi:    librpc.CreateKmdApi(m),
		rpcApi:    librpc.CreateRpcApi(m),
		resources: CategorizeResources(m.options.Resources),
	}
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

func (m *Machine) Call(f UserFunctionValue, arg Value, ctx *rx.Context) Value {
	return call(f, arg, m, ctx)
}

func (m *Machine) GetScheduler() rx.Scheduler {
	return m.scheduler
}

func (m *Machine) InjectExtraGlobals(values ([] Value)) {
	m.extraLock.Lock()
	defer m.extraLock.Unlock()
	m.extraSlot = append(m.extraSlot, values...)
}

func (m *Machine) KmdGetInfo() KmdInfo {
	return m.program.KmdInfo
}

func (m *Machine) GetRpcInfo() RpcInfo {
	return m.program.RpcInfo
}

func (m *Machine) KmdCallAdapter(info KmdAdapterInfo, x Value) Value {
	var f, exists = m.GetGlobalValue(info.Index)
	if !(exists) { panic("something went wrong") }
	return call(f.(UserFunctionValue), x, m, rx.Background())
}

func (m *Machine) KmdCallValidator(info KmdValidatorInfo, x Value) bool {
	var f, exists = m.GetGlobalValue(info.Index)
	if !(exists) { panic("something went wrong") }
	return FromBool(call(f.(UserFunctionValue), x, m, rx.Background()).(EnumValue))
}

