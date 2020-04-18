package common

import "reflect"
import . "kumachan/error"


type MachineHandle interface {
	Call(fv Value, arg Value) Value
	GetErrorPoint() ErrorPoint
}
var __t = MachineHandle(nil)
var __MachineHandleType = reflect.TypeOf(&__t).Elem()

type NativeFunction  func(arg Value, handle MachineHandle) Value


func AdaptNativeFunction(f interface{}) NativeFunction {
	var get_arg_val = func(v Value) reflect.Value {
		if v == nil {
			return reflect.ValueOf(struct{}{})
		} else {
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
		return func(_ Value, handle MachineHandle) Value {
			return AdaptReturnValue(f_rv.Call([]reflect.Value {}))
		}
	} else {
		if t.In(arity-1).AssignableTo(__MachineHandleType) {
			var net_arity = arity - 1
			if net_arity == 1 {
				return func(arg Value, handle MachineHandle) Value {
					var rv_args = []reflect.Value {
						get_arg_val(arg),
						reflect.ValueOf(handle),
					}
					return AdaptReturnValue(f_rv.Call(rv_args))
				}
			} else {
				return func(arg Value, handle MachineHandle) Value {
					var p = arg.(ProductValue)
					if len(p.Elements) != net_arity {
						panic("invalid input quantity")
					}
					var rv_args = make([]reflect.Value, arity)
					for i, e := range p.Elements {
						rv_args[i] = get_arg_val(e)
					}
					rv_args[net_arity] = reflect.ValueOf(handle)
					return AdaptReturnValue(f_rv.Call(rv_args))
				}
			}
		} else {
			if arity == 1 {
				return func(arg Value, handle MachineHandle) Value {
					return AdaptReturnValue(f_rv.Call([]reflect.Value {
						get_arg_val(arg),
					}))
				}
			} else {
				return func(arg Value, handle MachineHandle) Value {
					var p = arg.(ProductValue)
					if len(p.Elements) != arity {
						panic("invalid input quantity")
					}
					var args = make([]reflect.Value, arity)
					for i, e := range p.Elements {
						args[i] = get_arg_val(e)
					}
					return AdaptReturnValue(f_rv.Call(args))
				}
			}
		}
	}
}

func AdaptReturnValue(values []reflect.Value) Value {
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
