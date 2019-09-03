package object

type ObjectCategory uint64
const (
    OC_Singleton ObjectCategory = iota
    // primitive
    OC_Bool
    OC_Byte            // uint8
    OC_Int             // int
    OC_IEEE754         // float64
    OC_BigInt
    OC_BigFloat
    OC_String          // immutable
    // native
    OC_NativeObject
    // function
    OC_Function
    // struct
    OC_Schema
    OC_Struct
    // OO
    OC_Class           // callable(construct): CLASS(arg1, ...)
    OC_Interface
    OC_Instance
    // generics
    OC_TypeTemplate
    OC_FunctionTemplate
    // misc types
    OC_PlainType       // consists of a func(Object)bool
    OC_CompoundType    // union of atomic/intersection types (may be a enum)
    OC_FunctionSignature
    // module
    OC_Module
)

func (obj Object) Category() ObjectCategory {
    return obj.__Category
}
