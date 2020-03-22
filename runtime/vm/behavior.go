package vm

import (
	"fmt"
	. "kumachan/runtime/common"
	"kumachan/runtime/common/rx"
	"kumachan/runtime/lib"
)


func Execute(p Program, m *Machine) {
	var L = len(p.Functions) + len(p.Constants) + len(p.Constants)
	if L > GlobalSlotMaxSize { panic("maximum registry size exceeded") }
	var N = lib.NativeFunctions
	m.GlobalSlot = make([]Value, 0, L)
	for _, v := range p.DataValues {
		m.GlobalSlot = append(m.GlobalSlot, v.ToValue())
	}
	for i, _ := range p.Functions {
		var f = &(p.Functions[i])
		m.GlobalSlot = append(m.GlobalSlot, f.ToValue(N))
	}
	for i, _ := range p.Closures {
		var f = &(p.Closures[i])
		m.GlobalSlot = append(m.GlobalSlot, f.ToValue(nil))
	}
	for i, _ := range p.Constants {
		var f = &(p.Constants[i])
		var v = f.ToValue(N)
		m.GlobalSlot = append(m.GlobalSlot, m.Call(v, nil))
	}
	var sched = rx.TrivialScheduler { EventLoop: m.EventLoop }
	var ctx = rx.Background()
	for i, _ := range p.Effects {
		var f = &(p.Effects[i])
		var v = f.ToValue(nil)
		var e = (m.Call(v, nil)).(rx.Effect)
		rx.RunEffect(e, sched, ctx, nil, nil)
	}
}


func CallFunction (f FunctionValue, arg Value, m *Machine) Value {
	var ec = m.ContextPool.Get().(*ExecutionContext)
	defer (func() {
		var err = recover()
		if err != nil {
			PrintRuntimeErrorMessage(err, ec)
			panic(err)
		}
	}) ()
	ec.PushCall(f, arg)
	outer: for len(ec.CallStack) > 0 {
		var code = ec.WorkingFrame.Function.Code
		var base_addr = ec.WorkingFrame.BaseAddr
		var inst_ptr_ref = &(ec.WorkingFrame.InstPtr)
		for *inst_ptr_ref < uint(len(code)) {
			var inst = code[*inst_ptr_ref]
			*inst_ptr_ref += 1
			switch inst.OpCode {
			case NOP:
				// do nothing
			case GLOBAL:
				var id = inst.GetRegIndex()
				var gv = m.GlobalSlot[id]
				ec.PushValue(gv)
			case LOAD:
				var offset = inst.GetOffset()
				var value = ec.DataStack[base_addr + offset]
				ec.PushValue(value)
			case STORE:
				var offset = inst.GetOffset()
				var value = ec.PopValue()
				ec.DataStack[base_addr + offset] = value
			case SUM:
				var index = inst.GetShortIndexOrSize()
				var value = ec.PopValue()
				ec.PushValue(SumValue {
					Index: index,
					Value: value,
				})
			case JIF:
				switch sum := ec.GetCurrentValue().(type) {
				case SumValue:
					if sum.Index == inst.GetShortIndexOrSize() {
						var new_inst_ptr = inst.GetDestAddr()
						assert(new_inst_ptr < uint(len(code)),
							"JIF: invalid address")
						*inst_ptr_ref = new_inst_ptr
					} else {
						// do nothing
					}
				default:
					panic("JIF: cannot execute on non-sum value")
				}
			case JMP:
				var new_inst_ptr = inst.GetDestAddr()
				assert(new_inst_ptr < uint(len(code)),
					"JMP: invalid address")
				*inst_ptr_ref = new_inst_ptr
			case PROD:
				var size = inst.GetIndexOrSize()
				var elements = make([]Value, size)
				for i := uint(0); i < size; i += 1 {
					elements[size-1-i] = ec.PopValue()
				}
				ec.PushValue(ProductValue {
					Elements: elements,
				})
			case GET:
				var index = inst.GetIndexOrSize()
				switch prod := ec.GetCurrentValue().(type) {
				case ProductValue:
					assert(index < uint(len(prod.Elements)),
						"GET: invalid index")
					ec.PushValue(prod.Elements[index])
				default:
					panic("GET: cannot execute on non-product value")
				}
			case SET:
				var index = inst.GetIndexOrSize()
				var value = ec.PopValue()
				switch prod := ec.PopValue().(type) {
				case ProductValue:
					var L = uint(len(prod.Elements))
					assert(index < L, "SET: invalid index")
					var draft = make([]Value, L)
					copy(draft, prod.Elements)
					draft[index] = value
					ec.PushValue(ProductValue {
						Elements: draft,
					})
				default:
					panic("GET: cannot execute on non-product value")
				}
			case CTX:
				var is_recursive = (inst.Arg1 != 0)
				switch prod := ec.PopValue().(type) {
				case ProductValue:
					var ctx = prod.Elements
					switch f := ec.PopValue().(type) {
					case FunctionValue:
						var required = int(f.Underlying.BaseSize.Context)
						var given = len(ctx)
						if is_recursive { given += 1 }
						assert(given == required, "CTX: invalid context size")
						assert((len(f.ContextValues) == 0), "CTX: context already injected")
						if is_recursive { ctx = append(ctx, nil) }
						var fv = FunctionValue {
							Underlying:    f.Underlying,
							ContextValues: ctx,
						}
						if is_recursive { ctx[len(ctx)-1] = fv }
						ec.PushValue(fv)
					default:
						panic("CTX: cannot inject context for non-function value")
					}
				default:
					panic("CTX: cannot use non-product value as context")
				}
			case CALL:
				switch f := ec.PopValue().(type) {
				case FunctionValue:
					// check if the function is valid
					var required = int(f.Underlying.BaseSize.Context)
					var current = len(f.ContextValues)
					assert(current == required,
						"CALL: missing correct context")
					var arg = ec.PopValue()
					// tail call optimization
					var L = uint(len(code))
					var next_inst_ptr = *inst_ptr_ref
					if next_inst_ptr < L {
						var next = code[next_inst_ptr]
						if next.OpCode == JMP && next.GetDestAddr() == L-1 {
							ec.PopTailCall()
						}
					} else {
						ec.PopTailCall()
					}
					// push the function to call stack
					ec.PushCall(f, arg)
					// check if call stack size exceeded
					var stack_size = uint(len(ec.DataStack))
					assert(stack_size < m.MaxStackSize,
						"CALL: stack overflow")
					// work on the pushed new frame
					continue outer
				case NativeFunctionValue:
					var arg = ec.PopValue()
					var ret = f(arg, m)
					ec.PushValue(ret)
				default:
					panic("CALL: cannot execute on non-callable value")
				}
			default:
				panic(fmt.Sprintf("invalid instruction %+v", inst))
			}
		}
		ec.PopCall()
	}
	var ret = ec.PopValue()
	ec.WorkingFrame = CallStackFrame {}
	for i, _ := range ec.CallStack {
		ec.CallStack[i] = CallStackFrame {}
	}
	ec.CallStack = ec.CallStack[:0]
	for i, _ := range ec.DataStack {
		ec.DataStack[i] = nil
	}
	ec.DataStack = ec.DataStack[:0]
	m.ContextPool.Put(ec)
	return ret
}


func (ec *ExecutionContext) GetCurrentValue() Value {
	return ec.DataStack[len(ec.DataStack) - 1]
}

func (ec *ExecutionContext) PushValue(v Value) {
	ec.DataStack = append(ec.DataStack, v)
}

func (ec *ExecutionContext) PopValue() Value {
	var L = len(ec.DataStack)
	assert(L > 0, "cannot pop empty data stack")
	var cur = (L - 1)
	var popped = ec.DataStack[cur]
	ec.DataStack[cur] = nil
	ec.DataStack = ec.DataStack[:cur]
	return popped
}

func (ec *ExecutionContext) PopValuesTo(addr uint) {
	var L = uint(len(ec.DataStack))
	assert(L > 0, "cannot pop empty data stack")
	assert(addr < L, "invalid data stack address")
	for i := addr; i < L; i += 1 {
		ec.DataStack[i] = nil
	}
	ec.DataStack = ec.DataStack[:addr]
}

func (ec *ExecutionContext) PushCall(f FunctionValue, arg Value) {
	var context_size = int(f.Underlying.BaseSize.Context)
	var reserved_size = int(f.Underlying.BaseSize.Reserved)
	assert(context_size == len(f.ContextValues),
		"invalid number of context values")
	var new_base_addr = uint(len(ec.DataStack))
	for i := 0; i < context_size; i += 1 {
		ec.PushValue(f.ContextValues[i])
	}
	for i := 0; i < reserved_size; i += 1 {
		ec.PushValue(nil)
	}
	ec.PushValue(arg)
	ec.CallStack = append(ec.CallStack, ec.WorkingFrame)
	ec.WorkingFrame = CallStackFrame {
		Function: f.Underlying,
		BaseAddr: new_base_addr,
		InstPtr:  0,
	}
}

func (ec *ExecutionContext) PopCall() {
	var L = len(ec.CallStack)
	assert(L > 0, "cannot pop empty call stack")
	var cur = (L - 1)
	var popped = ec.CallStack[cur]
	ec.CallStack[cur] = CallStackFrame {}
	ec.CallStack = ec.CallStack[:cur]
	var ret = ec.PopValue()
	ec.PopValuesTo(ec.WorkingFrame.BaseAddr)
	ec.PushValue(ret)
	ec.WorkingFrame = popped
}

func (ec *ExecutionContext) PopTailCall() {
	var L = len(ec.CallStack)
	assert(L > 0, "cannot pop empty call stack")
	var cur = (L - 1)
	var popped = ec.CallStack[cur]
	ec.CallStack[cur] = CallStackFrame {}
	ec.CallStack = ec.CallStack[:cur]
	ec.PopValuesTo(ec.WorkingFrame.BaseAddr)
	ec.WorkingFrame = popped
}


func assert(ok bool, msg string) {
	if !ok { panic(msg) }
}
