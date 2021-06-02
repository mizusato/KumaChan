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
func (r *Frame) Func() UsualFuncValue {
	return r.function
}
func (r *Frame) Arg() Value {
	return r.argument
}
func (r *Frame) Code() *Code {
	return &(r.function.Entity.Code)
}
func (r *Frame) Last() LocalAddr {
	return LocalAddr(uint(len(r.data)) - 1)
}
func (r *Frame) Branch(f UsualFuncValue, arg Value) *Frame {
	return &Frame {
		function: f,
		argument: arg,
		data:     r.data,
	}
}
func (r *Frame) Static(addr LocalAddr) Value {
	return r.function.Entity.Code.Static[addr]
}
func (r *Frame) Context(addr LocalAddr) Value {
	return r.function.Context[addr]
}
func (r *Frame) Data(addr LocalAddr) Value {
	return r.data[addr]
}
func (r *Frame) DataRange(addr LocalAddr, size LocalSize) ([] Value) {
	var u = (addr + 1)
	return r.data[u: (u + size)]
}
func (r *Frame) DataDstRef(instruction_index LocalAddr) *Value {
	return &(r.data[(instruction_index + r.function.Entity.Code.Offset)])
}
func (r *Frame) WrapPanic(e interface{}) interface{} {
	return e  // TODO
}

