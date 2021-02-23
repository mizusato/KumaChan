package vm

import (
	"os"
	"strings"
	"kumachan/rx"
	. "kumachan/lang"
	. "kumachan/util/error"
)


type InteropHandle struct {
	machine   *Machine
	locator   InteropErrorPointLocator
	sync_ctx  *rx.Context
}

type InteropErrorPointLocator (func() ErrorPoint)
func InteropErrorPointLocatorFromStatic(p ErrorPoint) InteropErrorPointLocator {
	return InteropErrorPointLocator(func() ErrorPoint {
		return p
	})
}
func InteropErrorPointLocatorFromExecutionContext(ec *ExecutionContext) InteropErrorPointLocator {
	return InteropErrorPointLocator(func() ErrorPoint {
		return GetFrameErrorPoint(ec.workingFrame)
	})
}

func (h InteropHandle) Call(f Value, arg Value) Value {
	switch f := f.(type) {
	case FunctionValue:
		return call(f, arg, h.machine, h.sync_ctx)
	case NativeFunctionValue:
		return f(arg, h)
	default:
		panic("cannot call a non-callable value")
	}
}

func (h InteropHandle) CallWithSyncContext(f Value, arg Value, ctx *rx.Context) Value {
	switch f := f.(type) {
	case FunctionValue:
		return call(f, arg, h.machine, ctx)
	case NativeFunctionValue:
		return f(arg, InteropHandle {
			machine:  h.machine,
			locator:  h.locator,
			sync_ctx: ctx,
		})
	default:
		panic("cannot call a non-callable value")
	}
}

func (h InteropHandle) SyncContext() *rx.Context {
	return h.sync_ctx
}

func (h InteropHandle) Scheduler() rx.Scheduler {
	return h.machine.scheduler
}

func (h InteropHandle) GetSysEnv() ([] string) {
	return h.machine.options.Environment
}

func (h InteropHandle) GetSysArgs() ([] string) {
	return h.machine.options.Arguments
}

func (h InteropHandle) GetStdIO() StdIO {
	return h.machine.options.StdIO
}

func (h InteropHandle) GetDebugOptions() DebugOptions {
	return h.machine.options.DebugOptions
}

func (h InteropHandle) ErrorPoint() ErrorPoint {
	return h.locator()
}

func (h InteropHandle) GetEntryModulePath() string {
	var raw = h.machine.program.MetaData.EntryModulePath
	return strings.TrimRight(raw, string([] rune { os.PathSeparator }))
}

func (h InteropHandle) GetKmdApi() KmdApi {
	return h.machine.kmdApi
}

func (h InteropHandle) GetRpcApi() RpcApi {
	return h.machine.rpcApi
}

func (h InteropHandle) GetResources(kind string) (map[string] Resource) {
	var clone = make(map[string] Resource)
	for path, item := range h.machine.resources[kind] {
		clone[path] = item
	}
	return clone
}

