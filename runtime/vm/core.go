package vm

import (
	"fmt"
	"kumachan/runtime/lib"
	. "kumachan/runtime/common"
)

const InitialDataStackCapacity = 32
const InitialCallStackCapacity = 8

func CallInNewThread (f FunctionValue, arg Value, m *Machine) Value {
	var thread = &Thread {
		Machine:   m,
		DataStack: make([]Value, 0, InitialDataStackCapacity),
		CallStack: make([]CallStackFrame, 0, InitialCallStackCapacity),
	}
	thread.PushCall(f, arg)
	for len(thread.CallStack) > 0 {
		var code = thread.WorkingFrame.Function.Code
		var base_addr = thread.WorkingFrame.BaseAddr
		var inst_ptr_ref = &(thread.WorkingFrame.InstPtr)
		for *inst_ptr_ref < len(code) {
			var inst = code[*inst_ptr_ref]
			*inst_ptr_ref += 1
			switch inst.OpCode {
			case NOP:
				// do nothing
			case GLOBAL:
				var id = inst.GetRegIndex()
				var gv = thread.Machine.GlobalValues[id]
				thread.PushValue(gv)
			case LOAD:
				var offset = inst.GetOffset()
				var value = thread.DataStack[base_addr + offset]
				thread.PushValue(value)
			case STORE:
				var offset = inst.GetOffset()
				var value = thread.PopValue()
				thread.DataStack[base_addr + offset] = value
			case SUM:
				var index = inst.GetShortIndexOrSize()
				var value = thread.PopValue()
				thread.PushValue(SumValue {
					Index: index,
					Value: value,
				})
			case JIF:
				switch sum := thread.GetCurrentValue().(type) {
				case SumValue:
					if sum.Index == inst.GetShortIndexOrSize() {
						var new_inst_ptr = inst.GetJumpAddr()
						assert(new_inst_ptr < len(code), "JIF: invalid address")
						*inst_ptr_ref = new_inst_ptr
					} else {
						// do nothing
					}
				default:
					panic("JIF: cannot execute on non-sum value")
				}
			case JMP:
				var new_inst_ptr = inst.GetJumpAddr()
				assert(new_inst_ptr < len(code), "JMP: invalid address")
				*inst_ptr_ref = new_inst_ptr
			case PROD:
				var size = inst.GetIndexOrSize()
				var elements = make([]Value, size)
				for i := 0; i < size; i += 1 {
					elements[size-1-i] = thread.PopValue()
				}
				thread.PushValue(ProductValue {
					Elements: elements,
				})
			case GET:
				var index = inst.GetIndexOrSize()
				switch prod := thread.GetCurrentValue().(type) {
				case ProductValue:
					assert(index < len(prod.Elements), "GET: invalid index")
					thread.PushValue(prod.Elements[index])
				default:
					panic("GET: cannot execute on non-product value")
				}
			case SET:
				var index = inst.GetIndexOrSize()
				var value = thread.PopValue()
				switch prod := thread.PopValue().(type) {
				case ProductValue:
					var L = len(prod.Elements)
					assert(index < L, "SET: invalid index")
					var draft = make([]Value, L)
					copy(draft, prod.Elements)
					draft[index] = value
					thread.PushValue(ProductValue {
						Elements: draft,
					})
				default:
					panic("GET: cannot execute on non-product value")
				}
			case CTX:
				switch prod := thread.PopValue().(type) {
				case ProductValue:
					var ctx = prod.Elements
					switch f := thread.PopValue().(type) {
					case FunctionValue:
						var required = int(f.Underlying.BaseSize.Context)
						var given = len(ctx)
						assert(given == required, "CTX: invalid context size")
						assert((len(f.ContextValues) == 0), "CTX: context already injected")
						thread.PushValue(FunctionValue {
							Underlying:    f.Underlying,
							ContextValues: ctx,
						})
					default:
						panic("CTX: cannot inject context for non-function value")
					}
				default:
					panic("CTX: cannot use non-product value as context")
				}
			case CALL:
				switch f := thread.PopValue().(type) {
				case FunctionValue:
					var required = int(f.Underlying.BaseSize.Context)
					var current = len(f.ContextValues)
					assert(current == required, "CALL: missing correct context")
					var arg = thread.PopValue()
					thread.PushCall(f, arg)
				default:
					panic("CALL: cannot execute on non-function value")
				}
			case NATIVE:
				var id = inst.GetRegIndex()
				var f = lib.NativeFunctions[id]
				var arg = thread.PopValue()
				var ret = f(arg, m)
				thread.PushValue(ret)
			default:
				panic(fmt.Sprintf("invalid instruction %+v", inst))
			}
		}
		thread.PopCall()
	}
	return thread.PopValue()
}

func (thread *Thread) GetCurrentValue() Value {
	return thread.DataStack[len(thread.DataStack) - 1]
}

func (thread *Thread) PushValue(v Value) {
	thread.DataStack = append(thread.DataStack, v)
}

func (thread *Thread) PopValue() Value {
	var L = len(thread.DataStack)
	assert(L > 0, "cannot pop empty data stack")
	var cur = (L - 1)
	var popped = thread.DataStack[cur]
	thread.DataStack[cur] = nil
	thread.DataStack = thread.DataStack[:cur]
	return popped
}

func (thread *Thread) PopValuesTo(addr int) {
	var L = len(thread.DataStack)
	assert(L > 0, "cannot pop empty data stack")
	assert(addr < L, "invalid data stack address")
	for i := addr; i < L; i += 1 {
		thread.DataStack[i] = nil
	}
	thread.DataStack = thread.DataStack[:addr]
}

func (thread *Thread) PushCall(f FunctionValue, arg Value) {
	var context_size = int(f.Underlying.BaseSize.Context)
	var reserved_size = int(f.Underlying.BaseSize.Reserved)
	assert(context_size == len(f.ContextValues), "invalid number of context values")
	var new_base_addr = len(thread.DataStack)
	for i := 0; i < context_size; i += 1 {
		thread.PushValue(f.ContextValues[i])
	}
	for i := 0; i < reserved_size; i += 1 {
		thread.PushValue(nil)
	}
	thread.PushValue(arg)
	thread.CallStack = append(thread.CallStack, thread.WorkingFrame)
	thread.WorkingFrame = CallStackFrame {
		Function: f.Underlying,
		BaseAddr: new_base_addr,
		InstPtr:  0,
	}
}

func (thread *Thread) PopCall() {
	var L = len(thread.CallStack)
	assert(L > 0, "cannot pop empty call stack")
	var cur = (L - 1)
	var popped = thread.CallStack[cur]
	thread.CallStack[cur] = CallStackFrame {}
	thread.CallStack = thread.CallStack[:cur]
	var ret = thread.PopValue()
	thread.PopValuesTo(thread.WorkingFrame.BaseAddr)
	thread.PushValue(ret)
	thread.WorkingFrame = popped
}

