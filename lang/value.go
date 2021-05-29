package lang

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
	. "kumachan/misc/util/error"
	"kumachan/stdlib"
)


type Value = interface {}
var __ValueReflectType = reflect.TypeOf((*Value)(nil)).Elem()
func ValueReflectType() reflect.Type { return __ValueReflectType }

type EnumValue = *ValEnum
type ValEnum struct {
	Index  uint
	Value  Value
}

type TupleValue = *ValTup
type ValTup struct {
	Elements  [] Value
}

type UserFunctionValue = *ValFun
type ValFun struct {
	Underlying     *Function
	ContextValues  [] Value
}

type NativeFunctionValue = *NativeFunction
func ValNativeFun(f NativeFunction) NativeFunctionValue {
	return &f
}

func RefEqual(a Value, b Value) bool {
	return refEqual(reflect.ValueOf(a), reflect.ValueOf(b))
}

func refEqual(x reflect.Value, y reflect.Value) bool {
	var tx = x.Type()
	var ty = y.Type()
	if tx == ty {
		var t = tx
		switch t.Kind() {
		case reflect.Struct:
			if t.NumField() == 0 {
				return true
			} else if t.NumField() == 1 {
				return refEqual(x.Field(0), y.Field(0))
			} else {
				return false
			}
		case reflect.Array:
			return false
		case reflect.Interface:
			panic("impossible branch")
		case reflect.Ptr:
			return (x.Pointer() == y.Pointer())
		case reflect.UnsafePointer:
			return (x.Pointer() == y.Pointer())
		case reflect.Chan:
			return (x.Pointer() == y.Pointer())
		case reflect.Func:
			// WARNING: this will read a private field of reflect.Value
			var u = reflect.ValueOf(&x).Elem().FieldByName("ptr").Pointer()
			var v = reflect.ValueOf(&y).Elem().FieldByName("ptr").Pointer()
			return (u == v)
		case reflect.Slice:
			return (x.Pointer() == y.Pointer() && x.Len() == y.Len())
		case reflect.Map:
			return (x.Pointer() == y.Pointer())
		case reflect.String:
			var u = x.String()
			var v = y.String()
			var uh = (*reflect.StringHeader)(unsafe.Pointer(&u))
			var vh = (*reflect.StringHeader)(unsafe.Pointer(&v))
			return (uh.Data == vh.Data && uh.Len == vh.Len)
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			return x.Int() == y.Int()
		case reflect.Uint, reflect.Uintptr, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			return x.Uint() == y.Uint()
		case reflect.Bool:
			return x.Bool() == y.Bool()
		case reflect.Float64, reflect.Float32:
			return x.Float() == y.Float()
		case reflect.Complex128, reflect.Complex64:
			return x.Complex() == y.Complex()
		default:
			return false
		}
	} else {
		return false
	}
}


type ArrayInfo struct {
	Length    uint
	ItemType  reflect.Type
}

type Inspectable interface {
	Inspect(inspect func(Value)ErrorMessage) ErrorMessage
}

func Inspect(value Value) ErrorMessage {
	return inspect(value, make([]uintptr, 0))
}

func inspect(value Value, path []uintptr) ErrorMessage {
	var rv = reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.UnsafePointer:
		var ptr = rv.Pointer()
		for _, visited := range path {
			if ptr == visited {
				// NOTE: circular reference is rare
				//       but could happen in mutable containers
				var msg = make(ErrorMessage, 0)
				msg.WriteText(TS_BOLD, "(Circular)")
				return msg
			}
		}
		var new_path = make([] uintptr, len(path))
		new_path = append(new_path, ptr)
		path = new_path
	}
	var msg = make(ErrorMessage, 0)
	switch v := value.(type) {
	case nil, struct{}:
		msg.WriteText(TS_NORMAL, "()")
	case [] uint32:
		var go_str = string(*(*([] rune))(unsafe.Pointer(&v)))
		msg.WriteText(TS_NORMAL, strconv.Quote(go_str))
	case reflect.Value:
		msg.WriteText(TS_NORMAL, "reflect.Value")
	case *reflect.Value:
		msg.WriteText(TS_NORMAL, "*reflect.Value")
	case EnumValue:
		// TODO: v.Index should be shown
		return inspect(v.Value, path)
	case TupleValue:
		msg.WriteText(TS_NORMAL, "(")
		for i, e := range v.Elements {
			msg.WriteAll(inspect(e, path))
			if i != len(v.Elements)-1 {
				msg.WriteText(TS_NORMAL, ", ")
			}
		}
		msg.WriteText(TS_NORMAL, ")")
	case UserFunctionValue:
		var name = v.Underlying.Info.Name
		msg.WriteText(TS_NORMAL, fmt.Sprintf("[func %s]", name))
	case NativeFunctionValue:
		msg.WriteText(TS_NORMAL, "[func (native)]")
	case Inspectable:
		msg.WriteAll(v.Inspect(func(value Value) ErrorMessage {
			return inspect(value, path)
		}))
	default:
		var rv = reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			var L = rv.Len()
			msg.WriteText(TS_NORMAL, "[")
			if L > 0 {
				msg.Write(T_LF)
			}
			for i := 0; i < L; i += 1 {
				var item = rv.Index(i).Interface()
				msg.WriteAllWithIndent(inspect(item, path), 1)
				if i != L-1 {
					msg.WriteText(TS_NORMAL, ",")
				}
				msg.Write(T_LF)
			}
			msg.WriteText(TS_NORMAL, "]")
		} else {
			msg.WriteText(TS_NORMAL,
				fmt.Sprintf("[%s %v]", rv.Type().String(), v))
		}
	}
	return msg
}

func Tuple2From(t TupleValue) (Value, Value) {
	if len(t.Elements) != 2 { panic("tuple size is not 2") }
	return t.Elements[0], t.Elements[1]
}

func Tuple(elements... Value) TupleValue {
	return &ValTup { elements }
}

func TupleOf(elements ([] Value)) TupleValue {
	return &ValTup { elements }
}

func SingleValueFromRecord(b TupleValue) Value {
	if len(b.Elements) != 1 { panic("record size is not 1") }
	return b.Elements[0]
}

func FromBool(p EnumValue) bool {
	if p.Value != nil { panic("something went wrong") }
	if p.Index == stdlib.YesIndex {
		return true
	} else if p.Index == stdlib.NoIndex {
		return false
	} else {
		panic("something went wrong")
	}
}

func ToBool(p bool) EnumValue {
	if p == true {
		return &ValEnum { Index: stdlib.YesIndex }
	} else {
		return &ValEnum { Index: stdlib.NoIndex }
	}
}

func FromOrdering(o EnumValue) Ordering {
	if o.Value != nil { panic("something went wrong") }
	switch o.Index {
	case stdlib.SmallerIndex:
		return Smaller
	case stdlib.EqualIndex:
		return Equal
	case stdlib.BiggerIndex:
		return Bigger
	default:
		panic("something went wrong")
	}
}

func ToOrdering(o Ordering) EnumValue {
	switch o {
	case Smaller:
		return &ValEnum { Index: stdlib.SmallerIndex }
	case Equal:
		return &ValEnum { Index: stdlib.EqualIndex }
	case Bigger:
		return &ValEnum { Index: stdlib.BiggerIndex }
	default:
		panic("impossible branch")
	}
}

func Some(v Value) EnumValue {
	return &ValEnum {
		Index: stdlib.SomeIndex,
		Value: v,
	}
}

func None() EnumValue {
	return &ValEnum {
		Index: stdlib.NoneIndex,
		Value: nil,
	}
}

func Ok(v Value) EnumValue {
	return &ValEnum {
		Index: stdlib.SuccessIndex,
		Value: v,
	}
}

func Ng(v Value) EnumValue {
	return &ValEnum {
		Index: stdlib.FailureIndex,
		Value: v,
	}
}

func Unwrap(maybe EnumValue) (Value, bool) {
	if maybe.Index == stdlib.SomeIndex {
		return maybe.Value, true
	} else {
		return nil, false
	}
}

func FromByte(v Value) uint8 {
	return v.(uint8)
}

func FromWord(v Value) uint16 {
	return v.(uint16)
}

func FromDword(v Value) uint32 {
	return v.(uint32)
}

func FromQword(v Value) uint64 {
	return v.(uint64)
}

func Struct2Prod(v interface{}) TupleValue {
	var rv = reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		panic("struct expected")
	}
	var elements = make([] Value, rv.NumField())
	for i := 0; i < rv.NumField(); i += 1 {
		elements[i] = ToValue(rv.Field(i).Interface())
	}
	return &ValTup { elements }
}

func ToValue(go_value interface{}) Value {
	switch v := go_value.(type) {
	case bool:
		return ToBool(v)
	// TODO: `rune` should be converted but it is a type alias of int32,
	//       consider how to fix this problem
	default:
		return v
	}
}

