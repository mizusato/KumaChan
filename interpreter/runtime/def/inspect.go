package def

import (
	"fmt"
	"reflect"
	"unsafe"
	"strconv"
	. "kumachan/standalone/util/error"
)


type Inspectable interface {
	Inspect(inspect func(Value)(ErrorMessage)) ErrorMessage
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
	case UsualFuncValue:
		var name = v.Entity.Name
		msg.WriteText(TS_NORMAL, fmt.Sprintf("[func %s]", strconv.Quote(name)))
	case NativeFuncValue:
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

