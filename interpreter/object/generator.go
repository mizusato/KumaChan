package object

import "unsafe"

type GeneratorPause struct {
    __Kind          GeneratorPauseKind
    __Value         Object
}

type GeneratorPauseKind int
const (
    GPK_Yield GeneratorPauseKind = iota
    GPK_Await
    GPK_Return
)

type GeneratorPauseObject struct {
    __NativeObject  NativeObject
    __Data          GeneratorPause
}

var GeneratorPauseClass = NativeClass {
    Name: "GeneratorPause",
    New: func(data interface{}) *NativeObject {
        return (*NativeObject)(unsafe.Pointer(&GeneratorPauseObject {
            __Data: *(data.(*GeneratorPause)),
        }))
    },
    Methods: map[string] NativeMethod {
        "is_yield": func (self unsafe.Pointer, _ []Object) Object {
            var pause_kind = (*GeneratorPauseObject)(self).__Data.__Kind
            return NewBool(pause_kind == GPK_Yield)
        },
        "is_await": func (self unsafe.Pointer, _ []Object) Object {
            var pause_kind = (*GeneratorPauseObject)(self).__Data.__Kind
            return NewBool(pause_kind == GPK_Await)
        },
        "is_return": func (self unsafe.Pointer, _ []Object) Object {
            var pause_kind = (*GeneratorPauseObject)(self).__Data.__Kind
            return NewBool(pause_kind == GPK_Return)
        },
        "value": func (self unsafe.Pointer, _ []Object) Object {
            return (*GeneratorPauseObject)(self).__Data.__Value
        },
    },
}
