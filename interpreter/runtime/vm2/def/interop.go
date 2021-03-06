package def

import (
	"reflect"
	"kumachan/standalone/rx"
)

type ExecutionCancelled struct {}
func (_ ExecutionCancelled) Error() string {
	return "execution cancelled"
}

type InteropContext interface {
	Call(f Value, arg Value) Value
	CallWithContext(ctx *rx.Context, f Value, arg Value) Value
	Context() *rx.Context
	Location() Location
	Scheduler() rx.Scheduler
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
	GetResource(kind string, path string) (Resource, bool)
}
type StdIO struct {
	Stdin   rx.File
	Stdout  rx.File
	Stderr  rx.File
}
type DebugOptions struct {
	DebugUI  bool
}
var __t = InteropContext(nil)
var __InteropContextType = reflect.TypeOf(&__t).Elem()

type NativeFunction func(arg Value, handle InteropContext) Value
type NativeConstant  func(handle InteropContext) Value
func (c NativeConstant) ToFunction() NativeFunction {
	return func(_ Value, h InteropContext) Value {
		return c(h)
	}
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
					var p = arg.(TupleValue)
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
					var p = arg.(TupleValue)
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
		return TupleOf(elements)
	}
}


