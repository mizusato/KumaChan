package object

type ObjectCategory int
const (
    OC_Singleton ObjectCategory = 0
    // primitive
    OC_Bool
    OC_Byte            // uint8
    OC_Int             // int
    OC_Float           // float64
    OC_BigInt
    OC_BigFloat
    OC_String          // immutable
    // native
    OC_NativeObject
    // struct
    OC_Schema          // callable(cast): SCHEMA(struct)
    OC_Struct
    // OO
    OC_Class           // callable(construct): CLASS(arg1, ...)
    OC_Interface
    OC_Instance
    // misc
    OC_Module
    OC_Function
    OC_TypeTemplate    // { __Inflater: Object, __HasDefault: bool, ... }
    OC_OneElementType
    OC_SimpleType      // consists of a func(Object)bool
    OC_SignatureType
    OC_CompoundType    // union of atomic/intersection types (may be a enum)
)

func (obj Object) Is (oc ObjectCategory) bool {
    return obj.__Category == oc
}
