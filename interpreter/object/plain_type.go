package object

import "math"
import "unsafe"
import ."kumachan/interpreter/assertion"

const __PlainTypeInvalidInit = "PlainType: invalid default type initialization"

// category types
var O_Type = GetTypeObject(6)
var O_Bool = GetTypeObject(7)
var O_Byte = GetTypeObject(8)
var O_Int = GetTypeObject(9)
var O_IEEE754 = GetTypeObject(10)
var O_BigInt = GetTypeObject(11)
var O_BigFloat = GetTypeObject(12)
var O_String = GetTypeObject(13)
var O_NativeObject = GetTypeObject(14)
var O_Function = GetTypeObject(15)
var O_Struct = GetTypeObject(16)
var O_Instance = GetTypeObject(17)
// subtypes of Int
var O_Size = GetTypeObject(18)
// subtypes of IEEE754
var O_Float = GetTypeObject(19)
var O_FloatSize = GetTypeObject(20)
// subtypes of String
var O_Char = GetTypeObject(21)

func __InitDefaultPlainTypes (context *ObjectContext) {
    var O_Type_ = __NewCategoryType(context, OC_Type)
    var O_Bool_ = __NewCategoryType(context, OC_Bool)
    var O_Byte_ = __NewCategoryType(context, OC_Byte)
    var O_Int_ = __NewCategoryType(context, OC_Int)
    var O_IEEE754_ = __NewCategoryType(context, OC_IEEE754)
    var O_BigInt_ = __NewCategoryType(context, OC_BigInt)
    var O_BigFloat_ = __NewCategoryType(context, OC_BigFloat)
    var O_String_ = __NewCategoryType(context, OC_String)
    var O_NativeObject_ = __NewCategoryType(context, OC_NativeObject)
    var O_Function_ = __NewCategoryType(context, OC_Function)
    var O_Struct_ = __NewCategoryType(context, OC_Struct)
    var O_Instance_ = __NewCategoryType(context, OC_Instance)
    Assert(O_Type_ == O_Type, __PlainTypeInvalidInit)
    Assert(O_Bool_ == O_Bool, __PlainTypeInvalidInit)
    Assert(O_Byte_ == O_Byte, __PlainTypeInvalidInit)
    Assert(O_Int_ == O_Int, __PlainTypeInvalidInit)
    Assert(O_IEEE754_ == O_IEEE754, __PlainTypeInvalidInit)
    Assert(O_BigInt_ == O_BigInt, __PlainTypeInvalidInit)
    Assert(O_BigFloat_ == O_BigFloat, __PlainTypeInvalidInit)
    Assert(O_String_ == O_String, __PlainTypeInvalidInit)
    Assert(O_NativeObject_ == O_NativeObject, __PlainTypeInvalidInit)
    Assert(O_Function_ == O_Function, __PlainTypeInvalidInit)
    Assert(O_Struct_ == O_Struct, __PlainTypeInvalidInit)
    Assert(O_Instance_ == O_Instance, __PlainTypeInvalidInit)
    var O_Size_ = __NewPlainSubType (
        context, "Size", UnwrapType(O_Int),
        func (o Object) bool {
            return UnwrapInt(o) >= 0
        },
    )
    var O_Float_ = __NewPlainSubType (
        context, "Float", UnwrapType(O_IEEE754),
        func (o Object) bool {
            var f = UnwrapIEEE754(o)
            return (!math.IsNaN(f) && !math.IsInf(f, 0))
        },
    )
    var O_FloatSize_ = __NewPlainSubType (
        context, "FloatSize", UnwrapType(O_Float),
        func (o Object) bool {
            return UnwrapIEEE754(o) >= 0
        },
    )
    var O_Char_ = __NewPlainSubType (
        context, "Char", UnwrapType(O_String),
        func (o Object) bool {
            var s = UnwrapString(o)
            if len(s) == 0 {
                return false
            } else {
                for i, _ := range s {
                    if i > 0 {
                        return false
                    }
                }
                return true
            }
        },
    )
    Assert(O_Size_ == O_Size, __PlainTypeInvalidInit)
    Assert(O_Float_ == O_Float, __PlainTypeInvalidInit)
    Assert(O_FloatSize_ == O_FloatSize, __PlainTypeInvalidInit)
    Assert(O_Char_ == O_Char, __PlainTypeInvalidInit)
}

func __NewCategoryType (context *ObjectContext, oc ObjectCategory) Object {
    var T = (*TypeInfo)(unsafe.Pointer(&T_Plain {
        __TypeInfo: TypeInfo {
            __Kind: TK_Plain,
            __Name: RepresentCategory(oc),
            __Initialized: true,
        },
        __Category: oc,
        __Checker: func (object Object) bool {
            return object.__Category == oc
        },
        __Parent: -1,
    }))
    context.__RegisterType(T)
    return GetTypeObject(T.__Id)
}

func __NewPlainSubType (
    context  *ObjectContext,
    name     string,
    parent   int,
    checker  func(Object)bool,
) Object {
    var parent_type = context.GetType(parent)
    Assert(parent_type.__Kind == TK_Plain, "PlainTypes: invalid parent")
    var parent_plain_type = (*T_Plain)(unsafe.Pointer(parent_type))
    var oc = parent_plain_type.__Category
    var T = (*TypeInfo)(unsafe.Pointer(&T_Plain {
        __TypeInfo: TypeInfo {
            __Kind: TK_Plain,
            __Name: name,
            __Initialized: true,
        },
        __Category: oc,
        __Checker: func (object Object) bool {
            return (object.__Category == oc) && checker(object)
        },
        __Parent: parent,
    }))
    context.__RegisterType(T)
    return GetTypeObject(T.__Id)
}
