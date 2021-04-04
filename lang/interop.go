package lang

import (
	"os"
	"sync"
	"reflect"
	"kumachan/misc/rx"
	. "kumachan/misc/util/error"
)


type InteropContext interface {
	Call(f Value, arg Value) Value
	CallWithSyncContext(f Value, arg Value, ctx *rx.Context) Value
	SyncContext() *rx.Context
	Scheduler() rx.Scheduler
	ErrorPoint() ErrorPoint
	Environment
}
type Environment interface {
	GetSysEnv() ([] string)
	GetSysArgs() ([] string)
	GetStdIO() StdIO
	GetDebugOptions() DebugOptions
	GetEntryModulePath() string
	GetKmdApi() KmdApi
	GetRpcApi() RpcApi
	GetResources(kind string) (map[string] Resource)
}
type StdIO struct {
	Stdin   *os.File
	Stdout  *os.File
	Stderr  *os.File
}
type DebugOptions struct {
	DebugUI  bool
}
var __t = InteropContext(nil)
var __InteropContextType = reflect.TypeOf(&__t).Elem()

type NativeFunction  func(arg Value, handle InteropContext) Value
type NativeConstant  func(handle InteropContext) Value
func (c NativeConstant) ToFunction() NativeFunction {
	var mu sync.Mutex
	var evaluated = false
	var value Value
	return func(_ Value, h InteropContext) Value {
		mu.Lock()
		defer mu.Unlock()
		if !(evaluated) {
			value = c(h)
			evaluated = true
		}
		return value
	}
}

type SyncCancellationError struct {}
func (SyncCancellationError) Error() string {
	return "synchronous operation cancelled"
}

func AdaptNativeFunction(f interface{}) NativeFunction {
	var get_arg_val = func(v Value, t reflect.Type) reflect.Value {
		if v == nil {
			return reflect.New(t).Elem()
		}
		switch v.(type) {
		case reflect.Value, *reflect.Value:
			// reflect.Value should not be used as Value
			panic("something went wrong")
		default:
			return reflect.ValueOf(v)
		}
	}
	var f_fit, ok = f.(NativeFunction)
	if ok {
		return f_fit
	}
	var f_rv = reflect.ValueOf(f)
	if f_rv.Kind() != reflect.Func { panic("invalid function") }
	var t = f_rv.Type()
	var arity = t.NumIn()
	if arity == 0 {
		return func(_ Value, handle InteropContext) Value {
			return AdaptReturnValue(f_rv.Call([] reflect.Value {}))
		}
	} else {
		if t.In(arity-1).AssignableTo(__InteropContextType) {
			var net_arity = arity - 1
			if net_arity == 1 {
				return func(arg Value, handle InteropContext) Value {
					var rv_args = [] reflect.Value {
						get_arg_val(arg, t.In(0)),
						reflect.ValueOf(handle),
					}
					return AdaptReturnValue(f_rv.Call(rv_args))
				}
			} else {
				return func(arg Value, handle InteropContext) Value {
					var p = arg.(ProductValue)
					if len(p.Elements) != net_arity {
						panic("invalid input quantity")
					}
					var rv_args = make([] reflect.Value, arity)
					for i, e := range p.Elements {
						rv_args[i] = get_arg_val(e, t.In(i))
					}
					rv_args[net_arity] = reflect.ValueOf(handle)
					return AdaptReturnValue(f_rv.Call(rv_args))
				}
			}
		} else {
			if arity == 1 {
				return func(arg Value, handle InteropContext) Value {
					return AdaptReturnValue(f_rv.Call([] reflect.Value {
						get_arg_val(arg, t.In(0)),
					}))
				}
			} else {
				return func(arg Value, handle InteropContext) Value {
					var p = arg.(ProductValue)
					if len(p.Elements) != arity {
						panic("invalid input quantity")
					}
					var args = make([] reflect.Value, arity)
					for i, e := range p.Elements {
						args[i] = get_arg_val(e, t.In(i))
					}
					return AdaptReturnValue(f_rv.Call(args))
				}
			}
		}
	}
}

func AdaptReturnValue(values ([] reflect.Value)) Value {
	if len(values) == 0 {
		return nil
	} else if len(values) == 1 {
		return values[0].Interface()
	} else {
		var elements = make([]Value, len(values))
		for i, e := range values {
			elements[i] = e.Interface()
		}
		return &ValProd { elements }
	}
}
