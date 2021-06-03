package frame

import (
	"unsafe"
	. "kumachan/interpreter/runtime/vm2/def"
)


type Frame struct {
	function  UsualFuncValue
	argument  Value
	data      AddrSpace
}

func CreateFrame(f UsualFuncValue, arg Value) *Frame {
	const CacheLinePadLength = LocalSize(64 / unsafe.Sizeof(Value(nil)))
	var req_size = f.Entity.Code.FrameRequiredSize()
	var alloc_size = (req_size + CacheLinePadLength)
	var data = make(AddrSpace, req_size, alloc_size)
	return &Frame {
		function: f,
		argument: arg,
		data:     data,
	}
}

func (u *Frame) Func() UsualFuncValue {
	return u.function
}

func (u *Frame) Arg() Value {
	return u.argument
}

func (u *Frame) Code() *Code {
	return &(u.function.Entity.Code)
}

func (u *Frame) LastDataAddr() LocalAddr {
	return LocalAddr(uint(len(u.data)) - 1)
}

func (u *Frame) LastInsAddr() LocalAddr {
	return LocalAddr(uint(len(u.function.Entity.Code.InsSeq)) - 1)
}

func (u *Frame) Branch(f UsualFuncValue, arg Value) *Frame {
	return &Frame {
		function: &ValFunc {
			Entity:  f.Entity,
			Context: u.function.Context,
		},
		argument: arg,
		data:     u.data,
	}
}

func (u *Frame) TailCall(f UsualFuncValue, arg Value) *Frame {
	if u.function.Entity == f.Entity {
		return &Frame {
			function: f,
			argument: arg,
			data:     u.data,
		}
	} else {
		var new_frame = CreateFrame(f, arg)
		u.data = new_frame.data
		return new_frame
	}
}

func (u *Frame) Static(addr LocalAddr) Value {
	return u.function.Entity.Code.Static[addr]
}

func (u *Frame) Context(addr LocalAddr) Value {
	return u.function.Context[addr]
}

func (u *Frame) Data(addr LocalAddr) Value {
	return u.data[addr]
}

func (u *Frame) DataRange(addr LocalAddr, size LocalSize) ([] Value) {
	var start = (addr + 1)
	return u.data[start: (start + size)]
}

func (u *Frame) DataDstRef(instruction_index LocalAddr) *Value {
	return &(u.data[(instruction_index + u.function.Entity.Code.Offset)])
}

func (u *Frame) WrapPanic(e interface{}) interface{} {
	return e  // TODO
}


