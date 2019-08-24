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
    // container
    OC_Array           // type List = Array<Any>, type Bytes = Array<Byte>
    OC_Map             // type Hash = Map<String, Any>
    OC_ArrayView       // immutable reference of mutable array
    OC_MapView         // immutable reference of mutable map
    // pseudo container
    OC_Iterator        // Iterator<T>
    OC_Stream          // Stream<T>, async iterator
    OC_Source          // Source<T>, just like Rx.Observable<T>
    OC_Promise         // Promise<T>
    // error
    OC_Error           // Error<T> in which T = type of payload, immutable
    // module
    OC_Module
    // function-like
    OC_Function
    OC_TypeTemplate    // { __Inflater: Object, __HasDefault: bool, ... }
    // struct
    OC_Schema          // callable(cast): SCHEMA(struct)
    OC_Struct
    // OO
    OC_Class           // callable(construct): CLASS(arg1, ...)
    OC_Interface
    OC_Instance
    // misc types
    OC_OneElementType
    OC_SimpleType      // consists of a func(Object)bool
    OC_SignatureType
    OC_CompoundType    // union of atomic/intersection types (may be a enum)
)

func (obj Object) Is (oc ObjectCategory) bool {
    return obj.__Category == oc
}
