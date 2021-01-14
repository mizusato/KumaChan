package vm

import (
	"fmt"
	"reflect"
	"kumachan/rx"
	"kumachan/runtime/api"
	. "kumachan/lang"
)


func assert(ok bool, msg string) {
	if !ok { panic(msg) }
}

func execute(p Program, m *Machine) {
	var L = len(p.Functions) + len(p.Constants) + len(p.Constants)
	assert(L <= GlobalSlotMaxSize, "maximum global slot size exceeded")
	m.globalSlot = make([]Value, 0, L)
	for _, v := range p.DataValues {
		m.globalSlot = append(m.globalSlot, v.ToValue())
	}
	for i, _ := range p.Functions {
		var f = p.Functions[i]
		m.globalSlot = append(m.globalSlot, f.ToValue(api.GetNativeFunction))
	}
	for i, _ := range p.Closures {
		var f = p.Closures[i]
		m.globalSlot = append(m.globalSlot, f.ToValue(nil))
	}
	var nil_ctx_handle = MachineContextHandle { machine: m, context: nil }
	for i, _ := range p.Constants {
		var f = p.Constants[i]
		switch f.Kind {
		case F_USER:
			var fv = f.ToValue(nil).(FunctionValue)
			m.globalSlot = append(m.globalSlot, call(fv, nil, m))
		case F_NATIVE:
			var v = api.GetNativeConstant(f.NativeId, nil_ctx_handle)
			m.globalSlot = append(m.globalSlot, v)
		case F_PREDEFINED:
			m.globalSlot = append(m.globalSlot, f.ToValue(nil))
		}
	}
	var ctx = rx.Background()
	var wg = make(chan bool, len(p.Effects))
	for i, _ := range p.Effects {
		var f = p.Effects[i]
		var e = (func() rx.Effect {
			var v = f.ToValue(nil)
			switch v := v.(type) {
			case FunctionValue:
				return (call(v, nil, m)).(rx.Effect)
			case rx.Effect:
				return v
			default:
				panic("something went wrong")
			}
		})()
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
				var ok = new_inst_ptr < uint(len(code))
				assert(ok, "JMP: invalid address")
				*inst_ptr_ref = new_inst_ptr
			case RSW:
				arg, ok := ec.popValue().(rx.Reactive)
				assert(ok, "RSW: cannot execute on non-reactive value")
				p, ok := ec.popValue().(ProductValue)
				assert(ok, "RSW: invalid branches")
				var consumers = make(map[uint] (func(rx.Reactive) rx.Effect))
				var default_consumer (func(rx.Effect) rx.Effect)
				for _, element := range p.Elements {
					pair, ok := element.(ProductValue)
					assert(ok, "RSW: invalid branch")
					assert(len(pair.Elements) == 2, "RSW: invalid branch")
					consumer, ok := pair.Elements[1].(FunctionValue)
					assert(ok, "RSW: invalid branch")
					if pair.Elements[0] == nil {
						assert(default_consumer == nil,
							"RSW: duplicate default branch")
						default_consumer = func(eff rx.Effect) rx.Effect {
							return call(consumer, eff, m).(rx.Effect)
						}
					} else {
						index, ok := pair.Elements[0].(uint)
						assert(ok, "RSW: invalid branch")
						consumers[index] = func(r rx.Reactive) rx.Effect {
							return call(consumer, r, m).(rx.Effect)
						}
					}
				}
				var eff = ConsumeReactiveSum(arg, consumers, default_consumer)
				ec.pushValue(eff)
			case PROD:
				var size = inst.GetShortIndexOrSize()
				var elements = make([] Value, size)
				for i := uint(0); i < size; i += 1 {
					elements[size-1-i] = ec.popValue()
				}
				ec.pushValue(&ValProd {
					Elements: elements,
				})
			case GET:
				var index = inst.GetShortIndexOrSize()
				switch v := ec.getCurrentValue().(type) {
				case ProductValue:
					var prod = v
					assert(index < uint(len(prod.Elements)),
						"GET: invalid index")
					ec.pushValue(prod.Elements[index])
				case rx.Reactive:
					var r = v
					ec.pushValue(ProjectReactiveProduct(r, index))
				default:
					panic("GET: cannot execute on non-product value")
				}
			case POPGET:
				var index = inst.GetShortIndexOrSize()
				switch v := ec.popValue().(type) {
				case ProductValue:
					var prod = v
					assert(index < uint(len(prod.Elements)),
						"POPGET: invalid index")
					ec.pushValue(prod.Elements[index])
				case rx.Reactive:
					var r = v
					ec.pushValue(ProjectReactiveProduct(r, index))
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
					var draft = make([] Value, L)
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
					assert(stack_size < m.options.MaxStackSize,
						"CALL: stack overflow")
					// work on the pushed new frame
					continue outer
				case NativeFunctionValue:
					var arg = ec.popValue()
					var ret = f(arg, MachineContextHandle { context: ec, machine: m })
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

func ProjectReactiveProduct(r rx.Reactive, index uint) rx.Reactive {
	var in = func(old_state rx.Object) func(rx.Object) rx.Object {
		return func(obj rx.Object) rx.Object {
			var prod = old_state.(ProductValue)
			var L = uint(len(prod.Elements))
			var draft = make([] Value, L)
			copy(draft, prod.Elements)
			draft[index] = obj
			return &ValProd { Elements: draft }
		}
	}
	var out = func(state rx.Object) rx.Object {
		var prod = state.(ProductValue)
		return prod.Elements[index]
	}
	return rx.ReactiveProject(r, in, out, &rx.KeyChain { Key: index })
}

func ConsumeReactiveSum (
	r                 rx.Reactive,
	consumers         (map[uint] (func(rx.Reactive) rx.Effect)),
	default_consumer  (func(rx.Effect) rx.Effect),
) rx.Effect {
	var branches = make(map[uint] rx.Reactive)
	for _i, _ := range consumers {
		var case_index = _i
		var in = func(obj rx.Object) rx.Object {
			return &ValSum {
				Index: Short(case_index),
				Value: obj,
			}
		}
		var out = func(obj rx.Object) (rx.Object, bool) {
			var sum = obj.(SumValue)
			if uint(sum.Index) == case_index {
				return sum.Value, true
			} else {
				return nil, false
			}
		}
		branches[case_index] = rx.ReactiveBranch(r, in, out)
	}
	var default_branch = r.Watch().Filter(func(obj rx.Object) bool {
		var sum = obj.(SumValue)
		var _, exists = branches[uint(sum.Index)]
		return !(exists)
	})
	// TODO: distinctUntilChanged operator
	var changing_index = r.Watch().Map(func(obj rx.Object) rx.Object {
		var sum = obj.(SumValue)
		return uint(sum.Index)
	}).Scan(func(acc rx.Object, this rx.Object)(rx.Object) {
		if acc == nil {
			return rx.Pair {
				First:  nil,
				Second: this,
			}
		} else {
			var pair = acc.(rx.Pair)
			return rx.Pair {
				First:  pair.Second,
				Second: this,
			}
		}
	}, nil).FilterMap(func(acc rx.Object) (rx.Object, bool) {
		var pair = acc.(rx.Pair)
		var prev, has_prev = pair.First.(uint)
		var this = pair.Second.(uint)
		if has_prev && prev == this {
			return nil, false
		} else {
			return this, true
		}
	})
	return changing_index.SwitchMap(func(obj rx.Object) rx.Effect {
		var case_index = obj.(uint)
		var branch, exists = branches[case_index]
		if exists {
			var consumer = consumers[case_index]
			return consumer(branch)
		} else {
			if default_consumer != nil {
				return default_consumer(default_branch)
			} else {
				panic("something went wrong")
			}
		}
	})
}

