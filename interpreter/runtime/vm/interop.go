package vm

import (
	"os"
	"strings"
	. "kumachan/interpreter/runtime/def"
	"kumachan/standalone/rx"
)


type InteropHandle struct {
	context   Context
	machine   *Machine
	location  Location
}

func (h InteropHandle) Call(v Value, arg Value) Value {
	switch f := v.(type) {
	case UsualFuncValue:
		return h.machine.Call(h.context, f, arg)
	case NativeFuncValue:
		return (*f)(arg, h)
	default:
		panic("cannot call a non-callable value")
	}
}

func (h InteropHandle) CallWithContext(ctx Context, f Value, arg Value) Value {
	return (InteropHandle {
		context:  ctx,
		machine:  h.machine,
		location: h.location,
	}).Call(f, arg)
}

func (h InteropHandle) Context() *rx.Context {
	return h.context
}
func (h InteropHandle) Location() Location {
	return h.location
}
func (h InteropHandle) Scheduler() rx.Scheduler {
	return h.machine.scheduler
}

func (h InteropHandle) GetSysEnv() ([] string) {
	return h.machine.options.SysEnv
}
func (h InteropHandle) GetSysArgs() ([] string) {
	return h.machine.options.SysArgs
}
func (h InteropHandle) GetStdIO() StdIO {
	return h.machine.options.StdIO
}
func (h InteropHandle) GetDebugOptions() DebugOptions {
	return h.machine.options.DebugOptions
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
func (h InteropHandle) GetResource(kind string, path string) (Resource, bool) {
	var category = h.machine.resources[kind]
	var res, exists = category[path]
	return res, exists
}

