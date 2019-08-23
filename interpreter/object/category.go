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
    // reference
    OC_Tuple           // immutable, consists of 2~4 elements
    OC_Array           // type List = Array<Any>, type Bytes = Array<Byte>
    OC_Map             // type Hash = Map<String, Any>
    OC_Function
    OC_Overload
    OC_Method
    OC_Iterator        // Iterator<T>
    OC_AsyncIterator   // AsyncIterator<T>
    OC_Source          // Source<T>, just like Rx.Observable<T>
    OC_Promise         // Promise<T>
    OC_Error           // Error<T>, T = type of payload
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
    OC_CompoundType
)

func (obj Object) Is (oc ObjectCategory) bool {
    return obj.__Category == oc
}
