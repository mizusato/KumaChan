package common

import (
	"fmt"
	"reflect"
	"strconv"
	. "kumachan/error"
	"kumachan/stdlib"
	"unsafe"
)


type Value = interface {}
var __ValueReflectType = reflect.TypeOf((*Value)(nil)).Elem()
func ValueReflectType() reflect.Type { return __ValueReflectType }

type SumValue = *ValSum
type ValSum struct {
	Index  Short
	Value  Value
}

type ProductValue = *ValProd
type ValProd struct {
	Elements  [] Value
}

type FunctionValue = *ValFunc
type ValFunc struct {
	Underlying     *Function
	ContextValues  [] Value
}

type NativeFunctionValue  NativeFunction

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
		var new_path = make([]uintptr, len(path))
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
	case SumValue:
		return inspect(v.Value, path)
	case ProductValue:
		msg.WriteText(TS_NORMAL, "(")
		for i, e := range v.Elements {
			msg.WriteAll(inspect(e, path))
			if i != len(v.Elements)-1 {
				msg.WriteText(TS_NORMAL, ", ")
			}
		}
		msg.WriteText(TS_NORMAL, ")")
	case FunctionValue:
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

func Tuple2From(t ProductValue) (Value, Value) {
	if len(t.Elements) != 2 { panic("tuple size is not 2") }
	return t.Elements[0], t.Elements[1]
}

func ToTuple2(a Value, b Value) ProductValue {
	return &ValProd { [] Value { a, b } }
}

func SingleValueFromBundle(b ProductValue) Value {
	if len(b.Elements) != 1 { panic("bundle size is not 1") }
	return b.Elements[0]
}

func BoolFrom(p SumValue) bool {
	if p.Value != nil { panic("something went wrong") }
	if p.Index == stdlib.YesIndex {
		return true
	} else if p.Index == stdlib.NoIndex {
		return false
	} else {
		panic("something went wrong")
	}
}

func ToBool(p bool) SumValue {
	if p == true {
		return &ValSum { Index: stdlib.YesIndex }
	} else {
		return &ValSum { Index: stdlib.NoIndex }
	}
}

func ToOrdering(o Ordering) SumValue {
	switch o {
	case Smaller:
		return &ValSum { Index: stdlib.SmallerIndex }
	case Equal:
		return &ValSum { Index: stdlib.EqualIndex }
	case Bigger:
		return &ValSum { Index: stdlib.BiggerIndex }
	default:
		panic("impossible branch")
	}
}

func Just(v Value) SumValue {
	return &ValSum {
		Index: stdlib.JustIndex,
		Value: v,
	}
}

func Na() SumValue {
	return &ValSum {
		Index: stdlib.NaIndex,
		Value: nil,
	}
}

func ByteFrom(v Value) uint8 {
	return v.(uint8)
}

func WordFrom(v Value) uint16 {
	return v.(uint16)
}

func DwordFrom(v Value) uint32 {
	return v.(uint32)
}

func QwordFrom(v Value) uint64 {
	return v.(uint64)
}
