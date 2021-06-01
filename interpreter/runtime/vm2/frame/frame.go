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
func (ctx *Frame) Func() UsualFuncValue {
	return ctx.function
}
func (ctx *Frame) Arg() Value {
	return ctx.argument
}
func (ctx *Frame) Code() *Code {
	return &(ctx.function.Entity.Code)
}
func (ctx *Frame) Last() LocalAddr {
	return LocalAddr(uint(len(ctx.data)) - 1)
}
func (ctx *Frame) Branch(f UsualFuncValue, arg Value) *Frame {
	return &Frame {
		function: f,
		argument: arg,
		data:     ctx.data,
	}
}
func (ctx *Frame) Static(addr LocalAddr) Value {
	return ctx.function.Entity.Code.Static[addr]
}
func (ctx *Frame) Context(addr LocalAddr) Value {
	return ctx.function.Context[addr]
}
func (ctx *Frame) Data(addr LocalAddr) Value {
	return ctx.data[addr]
}
func (ctx *Frame) DataRange(addr LocalAddr, size LocalSize) ([] Value) {
	var u = (addr + 1)
	return ctx.data[u: (u + size)]
}
func (ctx *Frame) DataDstRef(instruction_index LocalAddr) *Value {
	return &(ctx.data[(instruction_index + ctx.function.Entity.Code.Offset)])
}

