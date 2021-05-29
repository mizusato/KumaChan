package vm

import (
	"fmt"
	"reflect"
	"kumachan/misc/rx"
	"kumachan/runtime/api"
	. "kumachan/lang"
	"kumachan/runtime/lib/ui"
)


type ExecutionContext struct {
	dataStack     DataStack
	callStack     CallStack
	workingFrame  CallStackFrame
	indexBufLen   uint
	indexBuf      [ProductMaxSize] uint
}
type DataStack  [] Value
type CallStack  [] CallStackFrame
type CallStackFrame struct {
	function  *Function
	baseAddr  uint
	instPtr   uint
}

func execute(p Program, m *Machine) {
	var L = len(p.DataValues) + len(p.Functions) + len(p.Closures)
	assert(L <= GlobalSlotMaxSize, "maximum global slot size exceeded")
	m.globalSlot = make([] Value, L)
	var ptr = 0
	var add_global = func(v Value) {
		m.globalSlot[ptr] = v
		ptr += 1
	}
	for _, data := range p.DataValues {
		add_global(data.ToValue())
	}
	for _, f := range p.Functions {
		switch f.Kind {
		case F_USER:
			add_global(&ValFun { Underlying: f })
		case F_NATIVE:
			add_global(api.GetNativeFunctionValue(f.NativeId))
		case F_GENERATED:
			add_global(f.Generated)
		case F_RUNTIME_GENERATED:
			switch seed := f.Generated.(type) {
			case UiObjectThunk:
				add_global(ui.EvaluateObjectThunk(seed))
			default:
				panic("unknown runtime function generation seed")
			}
		default:
			panic("impossible branch")
		}
	}
	for _, f := range p.Closures {
		if f.Kind != F_USER { panic("something went wrong") }
		var v = &ValFun { Underlying: f }
		add_global(v)
	}
	var background = rx.Background()
	var wg = make(chan bool, len(p.Effects))
	for _, f := range p.Effects {
		switch f.Kind {
		case F_USER:
			var evaluate = &ValFun { Underlying: f }
			var e = (call(evaluate, nil, m, background)).(rx.Observable)
			rx.Schedule(e, m.scheduler, rx.Receiver {
				Context:   background,
				Terminate: wg,
			})
		case F_RUNTIME_GENERATED:
			var  e = f.Generated.(rx.Observable)
			rx.Schedule(e, m.scheduler, rx.Receiver {
				Context:   background,
				Terminate: wg,
			})
		default:
			panic("something went wrong")
		}
	}
	for i := 0; i < len(p.Effects); i += 1 {
		<- wg
	}
}

func call(f UserFunctionValue, arg Value, m *Machine, sync_ctx *rx.Context) Value {
	var ec = m.contextPool.Get().(*ExecutionContext)
	var l = InteropErrorPointLocatorFromExecutionContext(ec)
	var h = InteropHandle { machine: m, locator: l, sync_ctx: sync_ctx }
	defer (func() {
		var err = recover()
		if err != nil {
			var _, is_cancel = err.(SyncCancellationError)
			if is_cancel {
				ec.clear()
				m.contextPool.Put(ec)
				panic(err)
			} else {
				PrintRuntimeErrorMessage(err, ec)
				panic(err)
			}
		}
	}) ()
	ec.pushCall(f, arg)
	outer: for len(ec.callStack) > 0 {
		if sync_ctx.AlreadyCancelled() {
			panic(SyncCancellationError {})
		}
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
				var index = inst.GetGlobalIndex()
				var v, exists = m.GetGlobalValue(index)
				if !(exists) { panic("GLOBAL: value index out of range") }
				ec.pushValue(v)
			case LOAD:
				var offset = inst.GetOffset()
				var value = ec.dataStack[base_addr + offset]
				ec.pushValue(value)
			case STORE:
				var offset = inst.GetOffset()
				var value = ec.popValue()
				ec.dataStack[base_addr + offset] = value
			case SUM:
				var index = inst.GetShortIndexOrSize()
				var value = ec.popValue()
				ec.pushValue(&ValEnum {
					Index: index,
					Value: value,
				})
			case JIF:
				var sum, ok = ec.getCurrentValue().(EnumValue)
				assert(ok, "JIF: cannot execute on non-enum value")
				if sum.Index == inst.GetShortIndexOrSize() {
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
				var ok = new_inst_ptr < uint(len(code))
				assert(ok, "JMP: invalid address")
				*inst_ptr_ref = new_inst_ptr
			case PROD:
				var size = inst.GetShortIndexOrSize()
				var elements = make([] Value, size)
				for i := uint(0); i < size; i += 1 {
					elements[size-1-i] = ec.popValue()
				}
				ec.pushValue(TupleOf(elements))
			case GET:
				var index = inst.GetShortIndexOrSize()
				switch v := ec.getCurrentValue().(type) {
				case TupleValue:
					var prod = v
					assert(index < uint(len(prod.Elements)),
						"GET: invalid index")
					ec.pushValue(prod.Elements[index])
				default:
					panic("GET: cannot execute on non-tuple value")
				}
			case POPGET:
				var index = inst.GetShortIndexOrSize()
				switch v := ec.popValue().(type) {
				case TupleValue:
					var prod = v
					assert(index < uint(len(prod.Elements)),
						"POPGET: invalid index")
					ec.pushValue(prod.Elements[index])
				default:
					panic("POPGET: cannot execute on non-tuple value")
				}
			case SET:
				var index = inst.GetShortIndexOrSize()
				var value = ec.popValue()
				switch prod := ec.popValue().(type) {
				case TupleValue:
					var L = uint(len(prod.Elements))
					assert(index < L, "SET: invalid index")
					var draft = make([] Value, L)
					copy(draft, prod.Elements)
					draft[index] = value
					ec.pushValue(TupleOf(draft))
				default:
					panic("SET: cannot execute on non-tuple value")
				}
			case BRS, BRB, BRF, FRP, FRF:
				createRef(ec, inst)
			case CTX:
				var is_recursive = (inst.Arg0 != 0)
				switch prod := ec.popValue().(type) {
				case TupleValue:
					var ctx = prod.Elements
					switch f := ec.popValue().(type) {
					case UserFunctionValue:
						var required = int(f.Underlying.BaseSize.Context)
						var given = len(ctx)
						if is_recursive { given += 1 }
						assert(given == required, "CTX: invalid context size")
						assert((len(f.ContextValues) == 0), "CTX: context already injected")
						if is_recursive { ctx = append(ctx, nil) }
						var fv = &ValFun {
							Underlying:    f.Underlying,
							ContextValues: ctx,
						}
						if is_recursive { ctx[len(ctx)-1] = fv }
						ec.pushValue(fv)
					case NativeFunctionValue:
						var wrapped = ValNativeFun(func(arg Value, h InteropContext) Value {
							var arg_with_context = Tuple(arg, prod)
							return (*f)(arg_with_context, h)
						})
						ec.pushValue(wrapped)
					default:
						panic("CTX: cannot inject context for non-function value")
					}
				default:
					panic("CTX: cannot use non-tuple value as context")
				}
			case CALL:
				switch f := ec.popValue().(type) {
				case UserFunctionValue:
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
					assert(stack_size < m.options.MaxStackSize,
						"CALL: stack overflow")
					// work on the pushed new frame
					continue outer
				case NativeFunctionValue:
					var arg = ec.popValue()
					var ret = (*f)(arg, h)
					ec.pushValue(ret)
				default:
					panic("CALL: cannot execute on non-callable value")
				}
			case ARRAY:
				var id = inst.GetGlobalIndex()
				var info_v, exists = m.GetGlobalValue(id)
				if !(exists) { panic("ARRAY: info index out of range") }
				var info = info_v.(ArrayInfo)
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
				var index = inst.GetShortIndexOrSize()
				ec.indexBuf[ec.indexBufLen] = index
				ec.indexBufLen += 1
			case MSD:
				assert(ec.indexBufLen < ProductMaxSize,
					"MSD: index buffer overflow")
				ec.indexBuf[ec.indexBufLen] = ^(uint(0))
				ec.indexBufLen += 1
			case MSJ:
				var prod, ok = ec.getCurrentValue().(TupleValue)
				assert(ok, "MSJ: cannot execute on non-tuple value")
				assert(uint(len(prod.Elements)) == ec.indexBufLen,
					"MSJ: wrong index quantity")
				var matching = true
				for i, e := range prod.Elements {
					var sum, ok = e.(EnumValue)
					assert(ok, "MSJ: non-enum element value occurred")
					var desired = ec.indexBuf[i]
					if desired == ^(uint(0)) {
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
						narrowed[i] = e.(EnumValue).Value
					}
					ec.pushValue(TupleOf(narrowed))
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
	ec.clear()
	m.contextPool.Put(ec)
	return ret
}


func assert(ok bool, msg string) {
	if !ok { panic(msg) }
}

func (ec *ExecutionContext) clear() {
	ec.workingFrame = CallStackFrame {}
	for i, _ := range ec.callStack {
		ec.callStack[i] = CallStackFrame {}
	}
	ec.callStack = ec.callStack[:0]
	for i, _ := range ec.dataStack {
		ec.dataStack[i] = nil
	}
	ec.dataStack = ec.dataStack[:0]
	ec.indexBufLen = 0
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
	assert(addr <= L, "invalid data stack address")
	for i := addr; i < L; i += 1 {
		ec.dataStack[i] = nil
	}
	ec.dataStack = ec.dataStack[:addr]
}

func (ec *ExecutionContext) pushCall(f UserFunctionValue, arg Value) {
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

func createRef(ec *ExecutionContext, inst Instruction) {
	var index = inst.GetShortIndexOrSize()
	switch inst.OpCode {
	case BRS:
		var sum, ok = ec.popValue().(EnumValue)
		assert(ok, "BRS: invalid operand")
		ec.pushValue(ValNativeFun(func(arg Value, _ InteropContext) Value {
			var new_value, update = Unwrap(arg.(EnumValue))
			if update {
				return Tuple(&ValEnum {
					Index: index,
					Value: new_value,
				}, arg)
			} else {
				if sum.Index == index {
					return Tuple(sum, Some(sum.Value))
				} else {
					return Tuple(sum, None())
				}
			}
		}))
	case BRB:
		switch base := ec.popValue().(type) {
		case UserFunctionValue, NativeFunctionValue:
			ec.pushValue(ValNativeFun(func(arg Value, h InteropContext) Value {
				var t = h.Call(base, None())
				var pair = t.(TupleValue).Elements
				var base_sum = pair[0]
				var base_branch = pair[1]
				var value, has_value = Unwrap(base_branch.(EnumValue))
				var new_value, update = Unwrap(arg.(EnumValue))
				if has_value {
					if update {
						var u = h.Call(base, Some(&ValEnum{
							Index: index,
							Value: new_value,
						}))
						return Tuple(u.(TupleValue).Elements[0], arg)
					} else {
						var sum = value.(EnumValue)
						if sum.Index == index {
							return Tuple(base_sum, Some(sum.Value))
						} else {
							return Tuple(base_sum, None())
						}
					}
				} else {
					return Tuple(base_sum, None())
				}
			}))
		default:
			panic("BRB: invalid operand")
		}
	case BRF:
		switch base := ec.popValue().(type) {
		case UserFunctionValue, NativeFunctionValue:
			ec.pushValue(ValNativeFun(func(arg Value, h InteropContext) Value {
				var new_value, update = Unwrap(arg.(EnumValue))
				if update {
					var t = h.Call(base, Some(&ValEnum {
						Index: index,
						Value: new_value,
					}))
					return Tuple(t.(TupleValue).Elements[0], arg)
				} else {
					var value = h.Call(base, None())
					var pair = value.(TupleValue).Elements
					var base_prod = pair[0]
					var base_field = pair[1]
					var sum = base_field.(EnumValue)
					if sum.Index == index {
						return Tuple(base_prod, Some(sum.Value))
					} else {
						return Tuple(base_prod, None())
					}
				}
			}))
		default:
			panic("BRF: invalid operand")
		}
	case FRP:
		var prod, ok = ec.popValue().(TupleValue)
		assert(ok, "FRP: invalid operand")
		var L = uint(len(prod.Elements))
		assert(index < L, "FRP: invalid index")
		ec.pushValue(ValNativeFun(func(arg Value, _ InteropContext) Value {
			var new_value, update = Unwrap(arg.(EnumValue))
			if update {
				var draft =  make([] Value, L)
				copy(draft, prod.Elements)
				draft[index] = new_value
				return Tuple(TupleOf(draft), new_value)
			} else {
				return Tuple(prod, prod.Elements[index])
			}
		}))
	case FRF:
		switch base := ec.popValue().(type) {
		case UserFunctionValue, NativeFunctionValue:
			ec.pushValue(ValNativeFun(func(arg Value, h InteropContext) Value {
				var t = h.Call(base, None())
				var prod = t.(TupleValue).Elements[1].(TupleValue)
				var L = uint(len(prod.Elements))
				assert(index < L, "FRF: invalid index")
				var new_field_value, update = Unwrap(arg.(EnumValue))
				if update {
					var draft =  make([] Value, L)
					copy(draft, prod.Elements)
					draft[index] = new_field_value
					var new_prod_value = TupleOf(draft)
					var t = h.Call(base, Some(new_prod_value))
					var new_base_value = t.(TupleValue).Elements[0]
					return Tuple(new_base_value, new_field_value)
				} else {
					return Tuple(prod, prod.Elements[index])
				}
			}))
		default:
			panic("FRF: invalid operand")
		}
	default:
		panic("something went wrong")
	}
}


