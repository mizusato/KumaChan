package object

import "unsafe"
import ."kumachan/interpreter/assertion"

type NativeClassId int
type NativeClassMethod = func(unsafe.Pointer, [MAX_ARGS]Object) Result
type NativeClassMethodList = map[Identifier]NativeClassMethod

type NativeClass struct {
    __Name      string
    __Id        NativeClassId
    __Methods   NativeClassMethodList
}

func NewNativeClass (
    context   *ObjectContext,
    name      string,
    methods   NativeClassMethodList,
) *NativeClass {
    return context.__RegisterNativeClass(name, methods)
}

func (class *NativeClass) NewObject(data unsafe.Pointer) Object {
    return Object {
        __Category: OC_NativeObject,
        __Inline: uint64(class.__Id),
        __Pointer: data,
    }
}

func (class *NativeClass) CallMethod (
    object   Object,
    method   Identifier,
    argv     [MAX_ARGS]Object,
) (Result, bool) {
    Assert (
        object.__Category == OC_NativeObject,
        "NativeClass: cannot call the method on a non-native object",
    )
    var f, exists = class.__Methods[method]
    if exists {
        return f(object.__Pointer, argv), true
    } else {
        return Result{}, false
    }
}
