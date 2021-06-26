package def

import (
	"unsafe"
	"reflect"
	"kumachan/stdlib"
)


type Value = interface {}

type EnumValue = *ValEnum
type ValEnum struct {
	Index  ShortIndex
	Value  Value
}

type TupleValue = *ValTup
type ValTup struct {
	Elements  [] Value
}

type UsualFuncValue = *ValFunc
type ValFunc struct {
	Entity   *FunctionEntity
	Context  AddrSpace
}

type NativeFuncValue = *NativeFunction
func ValNativeFunc(f NativeFunction) NativeFuncValue {
	return &f
}

type InterfaceValue = *ValInterface
type ValInterface struct {
	ConcreteValue  Value
	DispatchTable  *DispatchTable
}
type DispatchTable struct {
	Methods   [] Value
	Included  [] *DispatchTable
}
type InterfaceTransformPath ([] uint)

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
func Unwrap(maybe EnumValue) (Value, bool) {
	if maybe.Index == stdlib.SomeIndex {
		return maybe.Value, true
	} else {
		return nil, false
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

// RefEqual tests the equality of its inputs.
// It basically compares pointers but looks into 1-struct / enum.
// IMPORTANT: should be only used on values of same checker.Type
func RefEqual(a Value, b Value) bool {
	if a == nil || b == nil {
		return (a == nil && b == nil)
	} else {
		return refEqual(reflect.ValueOf(a), reflect.ValueOf(b))
	}
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
			if x.Pointer() == y.Pointer() {
				return true
			} else if t == reflect.TypeOf(&ValEnum {}) {
				var a = x.Interface().(EnumValue)
				var b = y.Interface().(EnumValue)
				return (a.Index == b.Index && RefEqual(a.Value, b.Value))
			} else {
				return false
			}
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


