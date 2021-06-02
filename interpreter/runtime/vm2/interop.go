package vm2

import (
	. "kumachan/interpreter/runtime/vm2/def"
)


type InteropHandle struct {
	context  Context
	machine  *Machine
}

func (h InteropHandle) Call(f Value, arg Value) Value {
	switch f := f.(type) {
	case UsualFuncValue:
		var ret = make(chan func() (interface{}, Value), 1)
		call(h.context, h.machine, f, arg, func(e interface{}, v Value) {
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
	case NativeFuncValue:
		return (*f)(arg, h)
	default:
		panic("cannot call a non-callable value")
	}
}

