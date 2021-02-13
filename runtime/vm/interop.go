package vm

import (
	"os"
	"strings"
	"kumachan/rx"
	. "kumachan/lang"
	. "kumachan/util/error"
)


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

func (h MachineContextHandle) GetKmdApi() KmdApi {
	return h.machine.kmdApi
}

func (h MachineContextHandle) GetResources(kind string) (map[string] Resource) {
	var res = make(map[string] Resource)
	for path, item := range h.machine.options.Resources {
		if item.Kind == kind {
			res[path] = item
		}
	}
	return res
}
