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

func callTailRec(ctx Context, m *Machine, f UsualFuncValue, arg Value, frame *Frame, kv ContVal) {
	execFrame(ctx, m, frame.TailRec(f, arg), kv)
}

func execBranch(ctx Context, m *Machine, f UsualFuncValue, arg Value, frame *Frame, kv ContVal) {
	execFrame(ctx, m, frame.Branch(f, arg), kv)
}

func execFrame(ctx Context, m *Machine, frame *Frame, kv ContVal) {
	var k = Cont(func(e interface{}) {
		if e != nil {
			kv(frame.WrapPanic(e), nil)
		} else {
			kv(nil, frame.Data(frame.LastDataAddr()))
		}
	})
	if m.options.ParallelEnabled {
		execParallel(ctx, m, frame, 0, k)
	} else {
		var flow = Flow { Start: 0, End: frame.LastInsAddr() }
		execFlow(ctx, m, frame, flow, k)
	}
}

func execParallel(ctx Context, m *Machine, frame *Frame, stage uint, k0 Cont) {
	var once sync.Once
	var k = Cont(func(e interface{}) {
		once.Do(func() {
			k0(e)
		})
	})
	var stages = frame.Code().Stages
	if stage >= uint(len(stages)) {
		k(nil)
		return
	}
	var this_stage = stages[stage]
	var num_of_flows = uint(len(this_stage))
	if num_of_flows == 0 { panic("bad bytecode: empty stage") }
	if num_of_flows == 1 {
		var flow = this_stage.TheOnlyFlow()
		execFlow(ctx, m, frame, flow, func(e interface{}) {
			if e != nil {
				k(e)
				return
			}
			execParallel(ctx, m, frame, (stage + 1), k)
		})
	} else {
		var sem = make(chan struct{}, (num_of_flows - 1))
		this_stage.ForEachFlow(func(flow Flow) {
			m.parallel.Execute(func() {
				execFlow(ctx, m, frame, flow, func(e interface{}) {
					if e != nil {
						k(e)
						return
					}
					select {
					case sem <- struct{}{}:
					default:
						execParallel(ctx, m, frame, (stage + 1), k)
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

func execIns(ctx Context, m *Machine, frame *Frame, i LocalAddr, end LocalAddr, k Cont) {
	var code = frame.Code()
	if i > end {
		k(nil)
		return
	}
	var ins = code.InsSeq[i]
	var dst = frame.DataDstRef(i)
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
		execIns(ctx, m, frame, (i + 1), end, k)
	})
	switch ins.OpCode {
	case SIZE:
		// do nothing
	case ARG:
		*dst = frame.Arg()
	case STATIC:
		*dst = frame.Static(ins.Src)
	case CTX:
		*dst = frame.Context(ins.Src)
	case FRAME:
		*dst = frame.Data(ins.Src)
	case ENUM:
		*dst = &ValEnum {
			Index: ins.Idx,
			Value: frame.Data(ins.Obj),
		}
	case SWITCH:
		var branches_addr = ins.Src
		var obj = frame.Data(ins.Obj)
		var enum = obj.(EnumValue)
		var vec = CreateShortIndexVectorSingleElement(enum.Index)
		var target = code.ExtMap.ChooseBranch(ins.ExtIdx, vec)
		var num_of_branches = code.GetSizeInsValue(branches_addr)
		assert(target < num_of_branches, "SWITCH: invalid branch index")
		var branch = frame.Data(branches_addr + 1 + target)
		var f, ok = branch.(UsualFuncValue)
		assert(ok, "SWITCH: invalid branch")
		execBranch(ctx, m, f, enum.Value, frame, kv_dst); return
	case SELECT:
		var objects_addr = ins.Obj
		var branches_addr = ins.Src
		var num_of_objects = code.GetSizeInsValue(objects_addr)
		assert(uint(num_of_objects) < MaxShortIndexVectorElements,
			"SELECT: too many operands")
		var objects = frame.DataRange(objects_addr, num_of_objects)
		var indexes = make([] ShortIndex, num_of_objects)
		var values = make([] Value, num_of_objects)
		for n := uint(0); n < uint(num_of_objects); n += 1 {
			var enum = objects[n].(EnumValue)
			indexes[n] = enum.Index
			values[n] = enum.Value
		}
		var vec = CreateShortIndexVector(indexes)
		var target = code.ExtMap.ChooseBranch(ins.ExtIdx, vec)
		var num_of_branches = code.GetSizeInsValue(branches_addr)
		assert(target < num_of_branches, "SELECT: invalid branch index")
		var branch = frame.Data(branches_addr + 1 + target)
		var f, ok = branch.(UsualFuncValue)
		assert(ok, "SELECT: invalid branch")
		execBranch(ctx, m, f, values, frame, kv_dst); return
	case BR:
		var enum = frame.Data(ins.Obj).(EnumValue)
		*dst = BranchRef(enum, ins.Idx)
	case BRC:
		var base_ref = frame.Data(ins.Obj)
		*dst = BranchRefFromCaseRef(base_ref, ins.Idx)
	case BRP:
		var base_ref = frame.Data(ins.Obj)
		*dst = BranchRefFromProjRef(base_ref, ins.Idx)
	case TUPLE:
		var objects_addr = ins.Obj
		var num_of_objects = code.GetSizeInsValue(objects_addr)
		var elements = make([] Value, num_of_objects)
		copy(elements, frame.DataRange(objects_addr, num_of_objects))
		*dst = TupleOf(elements)
	case GET:
		var tuple = frame.Data(ins.Obj).(TupleValue)
		*dst = tuple.Elements[ins.Idx]
	case SET:
		var tuple = frame.Data(ins.Obj).(TupleValue)
		var new_element = frame.Data(ins.Src)
		var new_elements = make([] Value, len(tuple.Elements))
		copy(new_elements, tuple.Elements)
		new_elements[ins.Idx] = new_element
		*dst = TupleOf(new_elements)
	case FR:
		var tuple = frame.Data(ins.Obj).(TupleValue)
		*dst = FieldRef(tuple, ins.Idx)
	case FRP:
		var base_ref = frame.Data(ins.Obj)
		*dst = FieldRefFromProjRef(base_ref, ins.Idx)
	case LSV:
		var objects_addr = ins.Obj
		var num_of_objects = code.GetSizeInsValue(objects_addr)
		var list = make([] Value, num_of_objects)
		copy(list, frame.DataRange(objects_addr, num_of_objects))
		*dst = list
	case LSC:
		var objects_addr = ins.Obj
		var num_of_objects = code.GetSizeInsValue(objects_addr)
		var t = GetCompactArrayType(ins.Idx)
		var length = int(num_of_objects)
		var r_list = reflect.MakeSlice(t, length, length)
		var objects = frame.DataRange(objects_addr, num_of_objects)
		for index, item := range objects {
			r_list.Index(index).Set(reflect.ValueOf(item))
		}
		var list = r_list.Interface()
		*dst = list
	case MPS:
		panic("not implemented")  // TODO
	case MPI:
		panic("not implemented")  // TODO
	case CL:
		var values_addr = ins.Src
		var num_of_values = code.GetSizeInsValue(values_addr)
		var context = make([] Value, num_of_values)
		copy(context, frame.DataRange(values_addr, num_of_values))
		var f = frame.Data(ins.Obj)
		switch f := f.(type) {
		case UsualFuncValue:
			var required = f.Entity.ContextLength
			assert(required == num_of_values, "CL: invalid context length")
			assert(len(f.Context) == 0, "CL: operand is already a closure")
			*dst = &ValFunc {
				Entity:  f.Entity,
				Context: context,
			}
		case NativeFuncValue:
			*dst = ValNativeFunc(func(arg Value, h InteropContext) Value {
				var arg_with_context = Tuple(arg, context)
				return (*f)(arg_with_context, h)
			})
		default:
			panic("CL: invalid operand")
		}
	case CLR:
		var f, ok = frame.Data(ins.Obj).(UsualFuncValue)
		assert(ok, "CLR: invalid operand")
		var usual_values_addr = ins.Src
		var num_of_usual_values = code.GetSizeInsValue(usual_values_addr)
		var num_of_values = uint(num_of_usual_values + 1)
		var required = uint(f.Entity.ContextLength)
		assert(required == num_of_values, "CLR: invalid context length")
		assert(len(f.Context) == 0, "CLR: operand is already a closure")
		var context = make([] Value, num_of_values)
		copy(context, frame.DataRange(usual_values_addr, num_of_usual_values))
		var closure = &ValFunc {
			Entity:  f.Entity,
			Context: context,
		}
		context[len(context) - 1] = closure
		*dst = closure
	case CALL:
		if ctx.AlreadyCancelled() {
			panic(ExecutionCancelled {})
		}
		var f = frame.Data(ins.Obj)
		var arg = frame.Data(ins.Src)
		switch f := f.(type) {
		case UsualFuncValue:
			if i == end && end == frame.LastInsAddr() &&
			frame.Func().Entity == f.Entity {
				callTailRec(ctx, m, f, arg, frame, kv_dst); return
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
	execIns(ctx, m, frame, (i + 1), end, k)
}


