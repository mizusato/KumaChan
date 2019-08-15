package object

type ObjectCategory uint32
const (
    OC_Singleton ObjectCategory = iota
    OC_Bool
    OC_Float
    OC_Integer
    OC_String
    OC_Buffer
    OC_List
    OC_Hash
    OC_Function
    OC_Overload
    OC_Iterator
    OC_AsyncIterator
    OC_Observer
    OC_Promise
    OC_Error
    OC_Module
    OC_SubModule
    OC_Class
    OC_Instance
    OC_Interface
    OC_Schema
    OC_Struct
    OC_Enum
    OC_AnyType
    OC_NeverType
    OC_NativeType
    OC_GenericType
    OC_SignatureType
    OC_SubType
    OC_CompoundType
)

func (obj Object) Is (oc ObjectCategory) bool {
    return obj.__Category == oc
}
