package object

import "fmt"
import "strconv"

func RepresentCategory (oc ObjectCategory) string {
    switch oc {
    case OC_Type:
        return "Type"
    case OC_Bool:
        return "Bool"
    case OC_Byte:
        return "Byte"
    case OC_Int:
        return "Int"
    case OC_IEEE754:
        return "IEEE754"
    case OC_BigInt:
        return "BigInt"
    case OC_BigFloat:
        return "BigFloat"
    case OC_String:
        return "String"
    case OC_NativeObject:
        return "NativeObject"
    case OC_Function:
        return "Function"
    case OC_Struct:
        return "Struct"
    case OC_Instance:
        return "Instance"
    case OC_Module:
        return "Module"
    default:
        panic("invalid object category")
    }
}

func Represent (object Object, context *ObjectContext) string {
    var category = RepresentCategory(object.__Category)
    var description = "..."
    switch object.__Category {
    case OC_Type:
        description = context.GetType(int(object.__Inline)).__Name
    case OC_Int:
        description = fmt.Sprintf("%v", UnwrapInt(object))
    case OC_IEEE754:
        description = fmt.Sprintf("%v", UnwrapIEEE754(object))
    case OC_Bool:
        description = fmt.Sprintf("%v", UnwrapBool(object))
    case OC_String:
        description = strconv.Quote(UnwrapString(object))
    }
    return fmt.Sprintf("[%v %v]", category, description)
}
