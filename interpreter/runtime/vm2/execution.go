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

func assert(ok bool, msg string) {
	if !(ok) { panic(msg) }
}

func call(ctx Context, m *Machine, f UsualFuncValue, arg Value) Value {
	return execFrame(ctx, m, CreateFrame(f, arg))
}

func execBranch(ctx Context, m *Machine, f UsualFuncValue, arg Value, frame *Frame) Value {
	return execFrame(ctx, m, frame.Branch(f, arg))
}

func execFrame(ctx Context, m *Machine, frame *Frame) Value {
	var f = frame.Func()
	if m.options.ParallelEnabled {
		for _, flows := range f.Entity.Code.Stages {
			var num_of_flows = uint(len(flows))
			if num_of_flows == 1 {
				var flow = flows[0]
				execFlow(ctx, m, frame, flow)
			} else {
				var wg sync.WaitGroup
				var err interface{}
				wg.Add(int(num_of_flows))
				for _, flow := range flows {
					var ok = m.parallel.Execute(func() {
						defer (func() {
							err = recover()
							wg.Done()
						})()
						execFlow(ctx, m, frame, flow)
					})
					if !(ok) {
						execFlow(ctx, m, frame, flow)
						wg.Done()
					}
				}
				wg.Wait()
				if err != nil {
					panic(err)
				}
			}
		}
	} else {
		var flow = Flow { Start: 0, End: frame.Last() }
		execFlow(ctx, m, frame, flow)
	}
	return frame.Data(frame.Last())
}

func execFlow(ctx Context, m *Machine, frame *Frame, flow Flow) {
	var code = frame.Code()
	for i := flow.Start; i <= flow.End; i += 1 {
		var this = code.InsSeq[i]
		var dst = frame.DataDstRef(i)
		switch this.OpCode {
		case SIZE:
			// do nothing
		case ARG:
			*dst = frame.Arg()
		case STATIC:
			*dst = frame.Static(this.Src)
		case CTX:
			*dst = frame.Context(this.Src)
		case FRAME:
			*dst = frame.Data(this.Src)
		case ENUM:
			*dst = &ValEnum {
				Index: this.Idx,
				Value: frame.Data(this.Obj),
			}
		case SWITCH:
			var branches_addr = this.Src
			var obj = frame.Data(this.Obj)
			var enum = obj.(EnumValue)
			var vec = CreateShortIndexVectorSingleElement(enum.Index)
			var target = code.ExtMap.ChooseBranch(this.ExtIdx, vec)
			var num_of_branches = code.GetSizeInsValue(branches_addr)
			assert(target < num_of_branches, "SWITCH: invalid branch index")
			var branch = frame.Data(branches_addr + 1 + target)
			var f, ok = branch.(UsualFuncValue)
			assert(ok, "SWITCH: invalid branch")
			*dst = execBranch(ctx, m, f, enum.Value, frame)
		case SELECT:
			var objects_addr = this.Obj
			var branches_addr = this.Src
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
			var target = code.ExtMap.ChooseBranch(this.ExtIdx, vec)
			var num_of_branches = code.GetSizeInsValue(branches_addr)
			assert(target < num_of_branches, "SELECT: invalid branch index")
			var branch = frame.Data(branches_addr + 1 + target)
			var f, ok = branch.(UsualFuncValue)
			assert(ok, "SELECT: invalid branch")
			*dst = execBranch(ctx, m, f, values, frame)
		case BR:
			var enum = frame.Data(this.Obj).(EnumValue)
			*dst = BranchRef(enum, this.Idx)
		case BRC:
			var base_ref = frame.Data(this.Obj)
			*dst = BranchRefFromCaseRef(base_ref, this.Idx)
		case BRP:
			var base_ref = frame.Data(this.Obj)
			*dst = BranchRefFromProjRef(base_ref, this.Idx)
		case TUPLE:
			var objects_addr = this.Obj
			var num_of_objects = code.GetSizeInsValue(objects_addr)
			var elements = make([] Value, num_of_objects)
			copy(elements, frame.DataRange(objects_addr, num_of_objects))
			*dst = TupleOf(elements)
		case GET:
			var tuple = frame.Data(this.Obj).(TupleValue)
			*dst = tuple.Elements[this.Idx]
		case SET:
			var tuple = frame.Data(this.Obj).(TupleValue)
			var new_element = frame.Data(this.Src)
			var new_elements = make([] Value, len(tuple.Elements))
			copy(new_elements, tuple.Elements)
			new_elements[this.Idx] = new_element
			*dst = TupleOf(new_elements)
		case FR:
			var tuple = frame.Data(this.Obj).(TupleValue)
			*dst = FieldRef(tuple, this.Idx)
		case FRP:
			var base_ref = frame.Data(this.Obj)
			*dst = FieldRefFromProjRef(base_ref, this.Idx)
		case LSV:
			var objects_addr = this.Obj
			var num_of_objects = code.GetSizeInsValue(objects_addr)
			var list = make([] Value, num_of_objects)
			copy(list, frame.DataRange(objects_addr, num_of_objects))
			*dst = list
		case LSC:
			var objects_addr = this.Obj
			var num_of_objects = code.GetSizeInsValue(objects_addr)
			var t = GetCompactArrayType(this.Idx)
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
			var values_addr = this.Src
			var num_of_values = code.GetSizeInsValue(values_addr)
			var context = make([] Value, num_of_values)
			copy(context, frame.DataRange(values_addr, num_of_values))
			var f = frame.Data(this.Obj)
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
			var f, ok = frame.Data(this.Obj).(UsualFuncValue)
			assert(ok, "CLR: invalid operand")
			var usual_values_addr = this.Src
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
			var f = frame.Data(this.Obj)
			var arg = frame.Data(this.Src)
			switch f := f.(type) {
			case UsualFuncValue:
				*dst = call(ctx, m, f, arg)
			case NativeFuncValue:
				var h = InteropHandle { context: ctx, machine: m }
				*dst = (*f)(arg, h)
			default:
				panic("CALL: operand not callable")
			}
		default:
			panic(fmt.Sprintf("invalid instruction at %d", i))
		}
	}
}


