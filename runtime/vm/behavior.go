package vm

import (
	"fmt"
	"reflect"
	. "kumachan/runtime/common"
	"kumachan/runtime/rx"
	"kumachan/runtime/lib"
)


func assert(ok bool, msg string) {
	if !ok { panic(msg) }
}

func execute(p Program, m *Machine) {
	var L = len(p.Functions) + len(p.Constants) + len(p.Constants)
	assert(L <= GlobalSlotMaxSize, "maximum registry size exceeded")
	var F = func(index int) Value {
		// TODO: change raw array NativeFunctions to GetNativeFunction(i)
		return NativeFunctionValue(lib.NativeFunctions[index])
	}
	var C = func(index int) Value {
		return lib.NativeConstants[index]
	}
	m.globalSlot = make([]Value, 0, L)
	for _, v := range p.DataValues {
		m.globalSlot = append(m.globalSlot, v.ToValue())
	}
	for i, _ := range p.Functions {
		var f = p.Functions[i]
		m.globalSlot = append(m.globalSlot, f.ToValue(F))
	}
	for i, _ := range p.Closures {
		var f = p.Closures[i]
		m.globalSlot = append(m.globalSlot, f.ToValue(nil))
	}
	for i, _ := range p.Constants {
		var f = p.Constants[i]
		switch f.Kind {
		case F_USER:
			var v = f.ToValue(nil).(FunctionValue)
			m.globalSlot = append(m.globalSlot, call(v, nil, m))
		case F_NATIVE:
			m.globalSlot = append(m.globalSlot, C(f.NativeIndex))
		case F_PREDEFINED:
			m.globalSlot = append(m.globalSlot, f.ToValue(nil))
		}
	}
	var ctx = rx.Background()
	var wg = make(chan struct{}, len(p.Effects))
	for i, _ := range p.Effects {
		var f = p.Effects[i]
		var v = f.ToValue(nil).(FunctionValue)
		var e = (call(v, nil, m)).(rx.Effect)
		m.scheduler.RunTopLevel(e, rx.Receiver {
			Context:   ctx,
			Values:    nil,
			Error:     nil,
			Terminate: wg,
		})
	}
	for i := 0; i < len(p.Effects); i += 1 {
		<- wg
	}
}

func call(f FunctionValue, arg Value, m *Machine) Value {
	var ec = m.contextPool.Get().(*ExecutionContext)
	defer (func() {
		var err = recover()
		if err != nil {
			// TODO: innermost panic message is printed first, fix it
			PrintRuntimeErrorMessage(err, ec)
			panic(err)
		}
	}) ()
	ec.pushCall(f, arg)
	outer: for len(ec.callStack) > 0 {
		var code = ec.workingFrame.function.Code
		var base_addr = ec.workingFrame.baseAddr
		var inst_ptr_ref = &(ec.workingFrame.instPtr)
		for *inst_ptr_ref < uint(len(code)) {
			var inst = code[*inst_ptr_ref]
			*inst_ptr_ref += 1
			switch inst.OpCode {
			case NOP:
				// do nothing
			case NIL:
				ec.pushValue(nil)
			case POP:
				ec.popValue()
			case GLOBAL:
				var id = inst.GetGlobalIndex()
				var gv = m.globalSlot[id]
				ec.pushValue(gv)
			case LOAD:
				var offset = inst.GetOffset()
				var value = ec.dataStack[base_addr + offset]
				ec.pushValue(value)
			case STORE:
				var offset = inst.GetOffset()
				var value = ec.popValue()
				ec.dataStack[base_addr + offset] = value
			case SUM:
				var index = inst.GetRawShortIndexOrSize()
				var value = ec.popValue()
				ec.pushValue(&ValSum {
					Index: index,
					Value: value,
				})
			case JIF:
				var sum, ok = ec.getCurrentValue().(SumValue)
				assert(ok, "JIF: cannot execute on non-sum value")
				if sum.Index == inst.GetRawShortIndexOrSize() {
					ec.popValue()
					ec.pushValue(sum.Value)
					var new_inst_ptr = inst.GetDestAddr()
					assert(new_inst_ptr < uint(len(code)),
						"JIF: invalid address")
					*inst_ptr_ref = new_inst_ptr
				} else {
					// do nothing
				}
			case JMP:
				var new_inst_ptr = inst.GetDestAddr()
				assert(new_inst_ptr < uint(len(code)),
					"JMP: invalid address")
				*inst_ptr_ref = new_inst_ptr
			case PROD:
				var size = inst.GetShortIndexOrSize()
				var elements = make([]Value, size)
				for i := uint(0); i < size; i += 1 {
					elements[size-1-i] = ec.popValue()
				}
				ec.pushValue(&ValProd {
					Elements: elements,
				})
			case GET:
				var index = inst.GetShortIndexOrSize()
				switch prod := ec.getCurrentValue().(type) {
				case ProductValue:
					assert(index < uint(len(prod.Elements)),
						"GET: invalid index")
					ec.pushValue(prod.Elements[index])
				default:
					panic("GET: cannot execute on non-product value")
				}
			case POPGET:
				var index = inst.GetShortIndexOrSize()
				switch prod := ec.popValue().(type) {
				case ProductValue:
					assert(index < uint(len(prod.Elements)),
						"POPGET: invalid index")
					ec.pushValue(prod.Elements[index])
				default:
					panic("POPGET: cannot execute on non-product value")
				}
			case SET:
				var index = inst.GetShortIndexOrSize()
				var value = ec.popValue()
				switch prod := ec.popValue().(type) {
				case ProductValue:
					var L = uint(len(prod.Elements))
					assert(index < L, "SET: invalid index")
					var draft = make([]Value, L)
					copy(draft, prod.Elements)
					draft[index] = value
					ec.pushValue(&ValProd {
						Elements: draft,
					})
				default:
					panic("SET: cannot execute on non-product value")
				}
			case CTX:
				var is_recursive = (inst.Arg0 != 0)
				switch prod := ec.popValue().(type) {
				case ProductValue:
					var ctx = prod.Elements
					switch f := ec.popValue().(type) {
					case FunctionValue:
						var required = int(f.Underlying.BaseSize.Context)
						var given = len(ctx)
						if is_recursive { given += 1 }
						assert(given == required, "CTX: invalid context size")
						assert((len(f.ContextValues) == 0), "CTX: context already injected")
						if is_recursive { ctx = append(ctx, nil) }
						var fv = &ValFunc {
							Underlying:    f.Underlying,
							ContextValues: ctx,
						}
						if is_recursive { ctx[len(ctx)-1] = fv }
						ec.pushValue(fv)
					default:
						panic("CTX: cannot inject context for non-function value")
					}
				default:
					panic("CTX: cannot use non-product value as context")
				}
			case CALL:
				switch f := ec.popValue().(type) {
				case FunctionValue:
					// check if the function is valid
					var required = int(f.Underlying.BaseSize.Context)
					var current = len(f.ContextValues)
					assert(current == required,
						"CALL: missing correct context")
					var arg = ec.popValue()
					// tail call optimization
					var L = uint(len(code))
					var next_inst_ptr = *inst_ptr_ref
					if next_inst_ptr < L {
						var next = code[next_inst_ptr]
						if next.OpCode == JMP && next.GetDestAddr() == L-1 {
							ec.popTailCall()
						}
					} else {
						ec.popTailCall()
					}
					// push the function to call stack
					ec.pushCall(f, arg)
					// check if call stack size exceeded
					var stack_size = uint(len(ec.dataStack))
					assert(stack_size < m.maxStackSize,
						"CALL: stack overflow")
					// work on the pushed new frame
					continue outer
				case NativeFunctionValue:
					var arg = ec.popValue()
					var ret = f(arg, Handle { context: ec, machine: m })
					ec.pushValue(ret)
				default:
					panic("CALL: cannot execute on non-callable value")
				}
			case ARRAY:
				var id = inst.GetGlobalIndex()
				var info = m.globalSlot[id].(ArrayInfo)
				var t = reflect.SliceOf(info.ItemType)
				var rv = reflect.MakeSlice(t, 0, int(info.Length))
				var v = rv.Interface()
				ec.pushValue(v)
			case APPEND:
				var item = ec.popValue()
				var item_rv = reflect.ValueOf(item)
				var arr = ec.popValue()
				var arr_rv = reflect.ValueOf(arr)
				if arr_rv.Kind() == reflect.Slice {
					var appended_rv = reflect.Append(arr_rv, item_rv)
					var appended = appended_rv.Interface()
					ec.pushValue(appended)
				} else {
					panic("APPEND: cannot append to non-slice value")
				}
			case MS:
				ec.indexBufLen = 0
			case MSI:
				assert(ec.indexBufLen < ProductMaxSize,
					"MSI: index buffer overflow")
				var index = inst.GetRawShortIndexOrSize()
				ec.indexBuf[ec.indexBufLen] = index
				ec.indexBufLen += 1
			case MSD:
				assert(ec.indexBufLen < ProductMaxSize,
					"MSD: index buffer overflow")
				ec.indexBuf[ec.indexBufLen] = ^(Short(0))
				ec.indexBufLen += 1
			case MSJ:
				var prod, ok = ec.getCurrentValue().(ProductValue)
				assert(ok, "MSJ: cannot execute on non-product value")
				assert(uint(len(prod.Elements)) == ec.indexBufLen,
					"MSJ: wrong index quantity")
				var matching = true
				for i, e := range prod.Elements {
					var sum, ok = e.(SumValue)
					assert(ok, "MSJ: non-sum element value occurred")
					var desired = ec.indexBuf[i]
					if desired == ^(Short(0)) {
						continue
					} else {
						if sum.Index == desired {
							continue
						} else {
							matching = false
							break
						}
					}
				}
				if matching {
					ec.popValue()
					var narrowed = make([] Value, len(prod.Elements))
					for i, e := range prod.Elements {
						narrowed[i] = e.(SumValue).Value
					}
					ec.pushValue(&ValProd { Elements: narrowed })
					var new_inst_ptr = inst.GetDestAddr()
					assert(new_inst_ptr < uint(len(code)),
						"MSJ: invalid address")
					*inst_ptr_ref = new_inst_ptr
				} else {
					// do nothing
				}
			default:
				panic(fmt.Sprintf("invalid instruction %+v", inst))
			}
		}
		ec.popCall()
	}
	var ret = ec.popValue()
	ec.workingFrame = CallStackFrame {}
	for i, _ := range ec.callStack {
		ec.callStack[i] = CallStackFrame {}
	}
	ec.callStack = ec.callStack[:0]
	for i, _ := range ec.dataStack {
		ec.dataStack[i] = nil
	}
	ec.dataStack = ec.dataStack[:0]
	m.contextPool.Put(ec)
	return ret
}


func (ec *ExecutionContext) getCurrentValue() Value {
	return ec.dataStack[len(ec.dataStack) - 1]
}

func (ec *ExecutionContext) pushValue(v Value) {
	ec.dataStack = append(ec.dataStack, v)
}

func (ec *ExecutionContext) popValue() Value {
	var L = len(ec.dataStack)
	assert(L > 0, "cannot pop empty data stack")
	var cur = (L - 1)
	var popped = ec.dataStack[cur]
	ec.dataStack[cur] = nil
	ec.dataStack = ec.dataStack[:cur]
	return popped
}

func (ec *ExecutionContext) popValuesTo(addr uint) {
	var L = uint(len(ec.dataStack))
	assert(L > 0, "cannot pop empty data stack")
	assert(addr < L, "invalid data stack address")
	for i := addr; i < L; i += 1 {
		ec.dataStack[i] = nil
	}
	ec.dataStack = ec.dataStack[:addr]
}

func (ec *ExecutionContext) pushCall(f FunctionValue, arg Value) {
	var context_size = int(f.Underlying.BaseSize.Context)
	var reserved_size = int(f.Underlying.BaseSize.Reserved)
	assert(context_size == len(f.ContextValues),
		"invalid number of context values")
	var new_base_addr = uint(len(ec.dataStack))
	for i := 0; i < context_size; i += 1 {
		ec.pushValue(f.ContextValues[i])
	}
	for i := 0; i < reserved_size; i += 1 {
		ec.pushValue(nil)
	}
	ec.pushValue(arg)
	ec.callStack = append(ec.callStack, ec.workingFrame)
	ec.workingFrame = CallStackFrame {
		function: f.Underlying,
		baseAddr: new_base_addr,
		instPtr:  0,
	}
}

func (ec *ExecutionContext) popCall() {
	var L = len(ec.callStack)
	assert(L > 0, "cannot pop empty call stack")
	var cur = (L - 1)
	var popped = ec.callStack[cur]
	ec.callStack[cur] = CallStackFrame {}
	ec.callStack = ec.callStack[:cur]
	var ret = ec.popValue()
	ec.popValuesTo(ec.workingFrame.baseAddr)
	ec.pushValue(ret)
	ec.workingFrame = popped
}

func (ec *ExecutionContext) popTailCall() {
	var L = len(ec.callStack)
	assert(L > 0, "cannot pop empty call stack")
	var cur = (L - 1)
	var popped = ec.callStack[cur]
	ec.callStack[cur] = CallStackFrame {}
	ec.callStack = ec.callStack[:cur]
	ec.popValuesTo(ec.workingFrame.baseAddr)
	ec.workingFrame = popped
}
