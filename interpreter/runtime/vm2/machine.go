package vm2

import (
	"kumachan/standalone/rx"
	. "kumachan/interpreter/runtime/vm2/def"
	"kumachan/interpreter/runtime/lib/librpc"
	"kumachan/interpreter/runtime/api"
	"kumachan/interpreter/runtime/lib/ui"
)


type Machine struct {
	program    Program
	options    Options
	pool       GoroutinePool
	scheduler  rx.Scheduler
	GeneratedObjects
}

type Options struct {
	ParallelEnabled  bool
	Resources        map[string] Resource
	MaxStackSize     uint
	InjectedEffects  [] rx.Observable
	SysEnv           [] string
	SysArgs          [] string
	DebugOptions
	StdIO
}

type GeneratedObjects struct {
	functions  [] Value
	funcMap    map[Symbol] uint
	effects    [] uint
	kmdApi     KmdApi
	rpcApi     RpcApi
	resources  map[string] map[string] Resource  // kind -> path -> res
}

func Execute(p Program, opts Options, ret (chan <- *Machine)) {
	var sched = rx.TrivialScheduler {
		EventLoop: rx.SpawnEventLoop(),
	}
	var pool = CreateGoroutinePool()
	var m = &Machine {
		program:   p,
		options:   opts,
		pool:      pool,
		scheduler: sched,
	}
	generateObjects(m)
	if ret != nil {
		ret <- m
	}
	runAllEffects(m)
}

func generateObjects(m *Machine) {
	var program_functions = m.program.Functions
	var functions = make([] Value, len(program_functions))
	var funcMap = make(map[Symbol] uint)
	var effects = make([] uint, 0)
	for i, seed := range program_functions {
		var info = seed.GetInfo()
		funcMap[info.Symbol] = uint(i)
	}
	var ctx = FunctionSeedEvaluator {
		ProcessEntity: func(entity *FunctionEntity) {
			if entity.IsEffect {
				var index, exists = funcMap[entity.Symbol]
				if !(exists) { panic("something went wrong") }
				effects = append(effects, index)
			}
		},
		GetNativeFuncValue: func(id string) Value {
			return api.GetNativeFunctionValue(id)
		},
		GeneratedNativeFunctionSeedEvaluator: GeneratedNativeFunctionSeedEvaluator {
			EvaluateUiObjectSeed: ui.EvaluateObjectThunk,
		},
		StaticValueSeedEvaluator: StaticValueSeedEvaluator {
			GetFunctionReference: func(sym Symbol) *Value {
				var index, exists = funcMap[sym]
				if !(exists) { panic("something went wrong") }
				return &(functions[index])
			},
		},
	}
	for i, seed := range program_functions {
		functions[i] = seed.Evaluate(ctx)
	}
	m.GeneratedObjects = GeneratedObjects {
		functions: functions,
		funcMap:   funcMap,
		effects:   effects,
		kmdApi:    librpc.CreateKmdApi(m),
		rpcApi:    librpc.CreateRpcApi(m),
		resources: CategorizeResources(m.options.Resources),
	}
}

func runAllEffects(m *Machine) {
	var effects = make([] rx.Observable, 0)
	for _, e := range m.options.InjectedEffects {
		effects = append(effects, e)
	}
	for _, index := range m.effects {
		var f = m.functions[index].(UsualFuncValue)
		var v = m.Call(rx.Background(), f, nil)
		var e = v.(rx.Observable)
		effects = append(effects, e)
	}
	var wg = make(chan bool, len(effects))
	for _, e := range effects {
		rx.Schedule(e, m.scheduler, rx.Receiver {
			Context:   rx.Background(),
			Terminate: wg,
		})
	}
	for range effects {
		<- wg
	}
}

func (m *Machine) Call(ctx Context, f UsualFuncValue, arg Value) Value {
	var ret = make(chan func() (interface{}, Value), 1)
	call(ctx, m, f, arg, func(e interface{}, v Value) {
		select {
		case ret <- (func() (interface{}, Value) {
			return e, v
		}):
		default:
			panic("something went wrong")
		}
	})
	var e, v = (<- ret)()
	select {
	case ret <- nil:
	default:
		panic("something went wrong")
	}
	if e != nil {
		panic(e)
	}
	return v
}

func (m *Machine) GetFuncValue(sym Symbol) (Value, bool) {
	var index, exists = m.GeneratedObjects.funcMap[sym]
	if exists {
		return m.functions[index], true
	} else {
		return nil, false
	}
}

func (m *Machine) GetScheduler() rx.Scheduler {
	return m.scheduler
}

func (m *Machine) KmdGetInfo() KmdInfo {
	return m.program.KmdInfo
}

func (m *Machine) GetRpcInfo() RpcInfo {
	return m.program.RpcInfo
}

func (m *Machine) KmdCallAdapter(info KmdAdapterInfo, x Value) Value {
	var f, exists = m.GetFuncValue(info.Symbol)
	if !(exists) { panic("something went wrong") }
	return m.Call(rx.Background(), f.(UsualFuncValue), x)
}

func (m *Machine) KmdCallValidator(info KmdValidatorInfo, x Value) bool {
	var f, exists = m.GetFuncValue(info.Symbol)
	if !(exists) { panic("something went wrong") }
	return FromBool(m.Call(rx.Background(), f.(UsualFuncValue), x).(EnumValue))
}


