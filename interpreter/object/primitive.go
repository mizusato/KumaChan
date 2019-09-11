package object

import "unsafe"
import ."kumachan/interpreter/assertion"

const __InvalidUnwrap = "Primitive: unable to unwrap object of wrong category"
const __UnwrapNil = "Primitive: nil pointer occurred during unwrapping"

func NewInt (x int) Object {
    return Object {
        __Category: OC_Int,
        __Inline: uint64(*(*uint)(unsafe.Pointer(&x))),
    }
}

func UnwrapInt (o Object) int {
    Assert(o.__Category == OC_Int, __InvalidUnwrap)
    var t = uint(o.__Inline)
    return *(*int)(unsafe.Pointer(&t))
}

func NewIEEE754 (x float64) Object {
    return Object {
        __Category: OC_IEEE754,
        __Inline: *(*uint64)(unsafe.Pointer(&x)),
    }
}

func UnwrapIEEE754 (o Object) float64 {
    Assert(o.__Category == OC_IEEE754, __InvalidUnwrap)
    return *(*float64)(unsafe.Pointer(&o.__Inline))
}

func NewBool (x bool) Object {
    var num uint64
    if x {
        num = 1
    } else {
        num = 0
    }
    return Object {
        __Category: OC_Bool,
        __Inline: num,
    }
}

func UnwrapBool (o Object) bool {
    Assert(o.__Category == OC_Bool, __InvalidUnwrap)
    Assert(o.__Inline <= 1, "Primitive: invalid boolean value")
    return (o.__Inline != 0)
}

func NewString (x string) Object {
    return Object {
        __Category: OC_String,
        __Pointer: unsafe.Pointer(&x),
    }
}

func UnwrapString (o Object) string {
    Assert(o.__Category == OC_String, __InvalidUnwrap)
    Assert(o.__Pointer != nil, __UnwrapNil)
    return *(*string)(o.__Pointer)
}
