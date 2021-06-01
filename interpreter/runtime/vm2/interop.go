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
		return call(h.context, h.machine, f, arg)
	case NativeFuncValue:
		return (*f)(arg, h)
	default:
		panic("cannot call a non-callable value")
	}
}

