package vm2

import (
	"fmt"
	"sync"
	"reflect"
	"kumachan/standalone/rx"
	. "kumachan/interpreter/runtime/vm2/def"
	. "kumachan/interpreter/runtime/vm2/frame"
)


type Context = *rx.Context

type Cont = func(e interface{})

type ContVal = func(e interface{}, v Value)

func assert(ok bool, msg string) {
	if !(ok) { panic(msg) }
}

func call(ctx Context, m *Machine, f UsualFuncValue, arg Value, kv ContVal) {
	execFrame(ctx, m, CreateFrame(f, arg), kv)
}

func callAtTail(ctx Context, m *Machine, f UsualFuncValue, arg Value, u *Frame, kv ContVal) {
	execFrame(ctx, m, u.TailCall(f, arg), kv)
}

func execBranch(ctx Context, m *Machine, f UsualFuncValue, arg Value, u *Frame, kv ContVal) {
	execFrame(ctx, m, u.Branch(f, arg), kv)
}

func execFrame(ctx Context, m *Machine, u *Frame, kv ContVal) {
	var k = Cont(func(e interface{}) {
		if e != nil {
			kv(u.WrapPanic(e), nil)
		} else {
			kv(nil, u.Data(u.LastDataAddr()))
		}
	})
	if m.options.ParallelEnabled {
		execParallel(ctx, m, u, 0, k)
	} else {
		var flow = Flow { Start: 0, End: u.LastInsAddr() }
		execFlow(ctx, m, u, flow, k)
	}
}

func execParallel(ctx Context, m *Machine, u *Frame, stage uint, k0 Cont) {
	var once sync.Once
	var k = Cont(func(e interface{}) {
		once.Do(func() {
			k0(e)
		})
	})
	var stages = u.Code().Stages()
	if stage >= uint(len(stages)) {
		k(nil)
		return
	}
	var this_stage = stages[stage]
	var num_of_flows = uint(len(this_stage))
	if num_of_flows == 0 { panic("bad bytecode: empty stage") }
	if num_of_flows == 1 {
		var flow = this_stage.TheOnlyFlow()
		execFlow(ctx, m, u, flow, func(e interface{}) {
			if e != nil {
				k(e)
				return
			}
			execParallel(ctx, m, u, (stage + 1), k)
		})
	} else {
		var sem = make(chan struct{}, (num_of_flows - 1))
		this_stage.ForEachFlow(func(flow Flow) {
			m.parallel.Execute(func() {
				execFlow(ctx, m, u, flow, func(e interface{}) {
					if e != nil {
						k(e)
						return
					}
					select {
					case sem <- struct{}{}:
					default:
						execParallel(ctx, m, u, (stage + 1), k)
					}
				})
			})
		})
	}
}

func execFlow(ctx Context, m *Machine, frame *Frame, flow Flow, k Cont) {
	defer (func() {
		var e = recover()
		if e != nil {
			k(e)
		}
	})()
	execIns(ctx, m, frame, flow.Start, flow.End, k)
}

func execIns(ctx Context, m *Machine, u *Frame, i LocalAddr, end LocalAddr, k Cont) {
	var code = u.Code()
	if i > end {
		k(nil)
		return
	}
	var inst = code.Inst(i)
	var dst = u.DataDstRef(i)
	var kv_dst = ContVal(func(e interface{}, v Value) {
		if e != nil {
			k(e)
			return
		}
		*dst = v
		defer (func() {
			var e = recover()
			if e != nil {
				k(e)
			}
		})()
		execIns(ctx, m, u, (i + 1), end, k)
	})
	var op = inst.OpCode
	switch op {
	case SIZE:
		*dst = inst.ToSize()
	case ARG:
		*dst = u.Arg()
	case STATIC:
		*dst = u.Static(inst.Src)
	case CTX:
		*dst = u.Context(inst.Src)
	case FRAME:
		*dst = u.Data(inst.Src)
	case ENUM:
		*dst = &ValEnum {
			Index: inst.Idx,
			Value: u.Data(inst.Obj),
		}
	case SWITCH:
		var obj = u.Data(inst.Obj)
		var enum = obj.(EnumValue)
		var vec = CreateShortIndexVectorSingleElement(enum.Index)
		var target = code.ChooseBranch(inst.ExtIdx, vec)
		var f = code.BranchFuncValue(target)
		execBranch(ctx, m, f, enum.Value, u, kv_dst); return
	case SELECT:
		var objects_addr = inst.Obj
		var num_of_objects = u.DataGetSizeAt(objects_addr)
		assert(uint(num_of_objects) < MaxShortIndexVectorElements,
			"SELECT: too many operands")
		var objects = u.DataRange(objects_addr, num_of_objects)
		var indexes = make([] ShortIndex, num_of_objects)
		var values = make([] Value, num_of_objects)
		for n := uint(0); n < uint(num_of_objects); n += 1 {
			var enum = objects[n].(EnumValue)
			indexes[n] = enum.Index
			values[n] = enum.Value
		}
		var vec = CreateShortIndexVector(indexes)
		var target = code.ChooseBranch(inst.ExtIdx, vec)
		var f = code.BranchFuncValue(target)
		execBranch(ctx, m, f, values, u, kv_dst); return
	case BR:
		var enum = u.Data(inst.Obj).(EnumValue)
		*dst = BranchRef(enum, inst.Idx)
	case BRC:
		var base_ref = u.Data(inst.Obj)
		*dst = BranchRefFromCaseRef(base_ref, inst.Idx)
	case BRP:
		var base_ref = u.Data(inst.Obj)
		*dst = BranchRefFromProjRef(base_ref, inst.Idx)
	case TUPLE:
		var objects_addr = inst.Obj
		var num_of_objects = u.DataGetSizeAt(objects_addr)
		var elements = make([] Value, num_of_objects)
		copy(elements, u.DataRange(objects_addr, num_of_objects))
		*dst = TupleOf(elements)
	case GET:
		var tuple = u.Data(inst.Obj).(TupleValue)
		*dst = tuple.Elements[inst.Idx]
	case SET:
		var tuple = u.Data(inst.Obj).(TupleValue)
		var new_element = u.Data(inst.Src)
		var new_elements = make([] Value, len(tuple.Elements))
		copy(new_elements, tuple.Elements)
		new_elements[inst.Idx] = new_element
		*dst = TupleOf(new_elements)
	case FR:
		var tuple = u.Data(inst.Obj).(TupleValue)
		*dst = FieldRef(tuple, inst.Idx)
	case FRP:
		var base_ref = u.Data(inst.Obj)
		*dst = FieldRefFromProjRef(base_ref, inst.Idx)
	case LSV:
		var objects_addr = inst.Obj
		var num_of_objects = u.DataGetSizeAt(objects_addr)
		var list = make([] Value, num_of_objects)
		copy(list, u.DataRange(objects_addr, num_of_objects))
		*dst = list
	case LSC:
		var objects_addr = inst.Obj
		var num_of_objects = u.DataGetSizeAt(objects_addr)
		var t = GetCompactArrayType(inst.Idx)
		var length = int(num_of_objects)
		var r_list = reflect.MakeSlice(t, length, length)
		var objects = u.DataRange(objects_addr, num_of_objects)
		for index, item := range objects {
			r_list.Index(index).Set(reflect.ValueOf(item))
		}
		var list = r_list.Interface()
		*dst = list
	case MPS:
		panic("not implemented")  // TODO
	case MPI:
		panic("not implemented")  // TODO
	case CL, CLR:
		var src = inst.Src
		var src_length = u.DataGetSizeAt(src)
		var num_of_values = src_length
		if op == CLR { num_of_values += 1 }
		var context = make([] Value, num_of_values)
		copy(context, u.DataRange(src, src_length))
		var closure Value
		switch f := u.Data(inst.Obj).(type) {
		case UsualFuncValue:
			var required = f.Entity.ContextLength
			assert(num_of_values == required, "CL: invalid context length")
			assert(len(f.Context) == 0, "CL: operand is already a closure")
			closure = &ValFunc {
				Entity:  f.Entity,
				Context: context,
			}
		case NativeFuncValue:
			closure = ValNativeFunc(func(arg Value, h InteropContext) Value {
				var arg_with_context = Tuple(arg, context)
				return (*f)(arg_with_context, h)
			})
		default:
			panic("CL: invalid operand")
		}
		if op == CLR { context[num_of_values - 1] = closure }
		*dst = closure
	case CALL:
		if ctx.AlreadyCancelled() {
			panic(ExecutionCancelled {})
		}
		var f = u.Data(inst.Obj)
		var arg = u.Data(inst.Src)
		switch f := f.(type) {
		case UsualFuncValue:
			if i == end && end == u.LastInsAddr() {
				callAtTail(ctx, m, f, arg, u, kv_dst); return
			} else {
				call(ctx, m, f, arg, kv_dst); return
			}
		case NativeFuncValue:
			var h = InteropHandle { context: ctx, machine: m }
			*dst = (*f)(arg, h)
		default:
			panic("CALL: operand not callable")
		}
	default:
		panic(fmt.Sprintf("invalid instruction at %d", i))
	}
	execIns(ctx, m, u, (i + 1), end, k)
}


