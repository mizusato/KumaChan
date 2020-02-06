package vm

func Call (f V_Function, arg Value, global *GlobalState) Value {
	var state = &CallState {
		GlobalState: global,
		CallStack:   make([]CallStackFrame, 0),
		DataStack:   make([]Value, 0),
	}
	state.PushCall(f, arg)
	for {
		var code = state.Function.Code
		var inst = code[state.InstPtr]
		switch inst.OpCode {
			// TODO
		}
	}
}

func (state *CallState) PushValue(v Value) {
	state.DataStack = append(state.DataStack, v)
}

func (state *CallState) PopValue() Value {
	var L = len(state.DataStack)
	assert(L > 0, "cannot pop empty data stack")
	assert(state.BaseAddr <= L-1, "cannot pop below base address")
	var popped = state.DataStack[L - 1]
	state.DataStack[L - 1] = nil
	state.DataStack = state.DataStack[:L-1]
	return popped
}

func (state *CallState) PopValuesTo(addr int) {
	var L = len(state.DataStack)
	assert(L > 0, "cannot pop empty data stack")
	assert(addr < L, "invalid data stack address")
	assert(state.BaseAddr <= addr, "cannot pop below base address")
	for i := addr; i < L; i += 1 {
		state.DataStack[i] = nil
	}
	state.DataStack = state.DataStack[:addr]
}

func (state *CallState) PushCall(f V_Function, arg Value) {
	var context_size = int(f.Underlying.BaseSize.Context)
	var reserved_size = int(f.Underlying.BaseSize.Reserved)
	assert(context_size == len(f.ContextValues), "invalid number of context values")
	var new_base_addr = len(state.DataStack)
	for i := 0; i < context_size; i += 1 {
		state.PushValue(f.ContextValues[i])
	}
	for i := 0; i < reserved_size; i += 1 {
		state.PushValue(nil)
	}
	state.PushValue(arg)
	state.CallStack = append(state.CallStack, CallStackFrame {
		SavedFunction: state.Function,
		SavedBaseAddr: state.BaseAddr,
		SavedInstPtr:  state.InstPtr,
	})
	state.Function = f.Underlying
	state.BaseAddr = new_base_addr
	state.InstPtr = 0
}

func (state *CallState) PopCall() {
	var L = len(state.CallStack)
	assert(L > 0, "cannot pop empty call stack")
	var popped_frame = state.CallStack[L-1]
	state.CallStack[L-1] = CallStackFrame {}
	state.CallStack = state.CallStack[:L-1]
	var return_value = state.PopValue()
	state.PopValuesTo(state.BaseAddr)
	state.PushValue(return_value)
	state.Function = popped_frame.SavedFunction
	state.BaseAddr = popped_frame.SavedBaseAddr
	state.InstPtr = popped_frame.SavedInstPtr
}

