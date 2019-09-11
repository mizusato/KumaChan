package object

type ObjectCategory uint64
const (
    OC_Type ObjectCategory = iota
    OC_Bool
    OC_Byte            // uint8
    OC_Int             // int
    OC_IEEE754         // float64
    OC_BigInt
    OC_BigFloat
    OC_String          // immutable
    OC_NativeObject
    OC_Function
    OC_Struct
    OC_Instance
)

func (obj Object) Category() ObjectCategory {
    return obj.__Category
}
