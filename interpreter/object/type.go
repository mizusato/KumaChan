package object

import "unsafe"
import ."kumachan/interpreter/assertion"

type TypeId int

type TypeKind int
const (
    TK_Placeholder TypeKind = iota
    TK_Singleton
    TK_Plain
    TK_Function
    TK_Union
    TK_Trait
    TK_Schema
    TK_Class
    TK_Interface
)

type TypeInfo struct {
    __Kind          TypeKind
    __Id            int
    __Name          string
    __Initialized   bool
    __FromGeneric   bool
    __GenericId     int
    __GenericArgs   [] int
}

type T_Placeholder struct {
    __TypeInfo    TypeInfo
    __UpperBound  int
}

type T_Plain struct {
    __TypeInfo   TypeInfo
    __Category   ObjectCategory
    __Checker    func(Object)bool
    __Parent     int
}

type T_Function struct {
    __TypeInfo    TypeInfo
    __Items       [] T_Function_Item
}

type T_Function_Item struct {
    __Parameters    [] int
    __ReturnValue   int
}

type T_Union struct {
    __TypeInfo   TypeInfo
    __Elements   [] int
}

type T_Trait struct {
    __TypeInfo     TypeInfo
    __Constraints  [] int
}

type T_Schema struct {
    __TypeInfo      TypeInfo
    __Immutable     bool
    __Bases         [] int
    __Supers        [] int
    __Fields        [] SchemaField
    __OffsetTable   map[Identifier] int
}

type SchemaField struct {
    __Name           Identifier
    __Type           int
    __HasDefault     bool
    __DefaultValue   *Object
    __From           int
}

type T_Class struct {
    __TypeInfo          TypeInfo
    __Methods           map[Identifier] MethodInfo
    __BaseClasses       [] int
    __BaseInterfaces    [] int
    __SuperClasses      [] int
    __SuperInterfaces   [] int
}

type MethodInfo struct {
    __Type      int
    __From      int
    __Offset    int
    __FunInfo   int
}

type T_Interface struct {
    __TypeInfo      TypeInfo
    __MethodTypes   map[Identifier] int
}

func GetTypeObject (id int) Object {
    return Object {
        __Category: OC_Type,
        __Inline: uint64(id),
    }
}

func UnwrapType (object Object) int {
    Assert (
        object.__Category == OC_Type,
        "Type: unable to unwrap object of wrong category",
    )
    return int(object.__Inline)
}

type Triple int
const (
    False Triple = iota
    True
    Unknown
)

func (T *TypeInfo) IsSubTypeOf(U *TypeInfo, ctx *ObjectContext) Triple {
    // TODO: Any Type, Never Type
    /*
    TK_Placeholder
    TK_Function
    TK_Union
    TK_Trait
    TK_Schema
    TK_Class
    TK_Interface
    */
    if T.__Kind == TK_Placeholder {
        var B = ctx.GetType((*T_Placeholder)(unsafe.Pointer(T)).__UpperBound)
        return B.IsSubTypeOf(U, ctx)
    }
    // TODO: LowerBound
    switch T.__Kind {
    case TK_Singleton:
        if U.__Kind == TK_Singleton {
            if T.__Id == U.__Id {
                return True
            } else {
                return False
            }
        } else if U.__Kind == TK_Union {
            return (*T_Union)(unsafe.Pointer(U)).HasSubType(T, ctx)
        } else {
            return False
        }
    case TK_Plain:
        if U.__Kind == TK_Plain {
            var Tp = (*T_Plain)(unsafe.Pointer(T))
            var Up = (*T_Plain)(unsafe.Pointer(U))
            if Tp.__Category == Up.__Category {
                if Up.__Parent == -1 {
                    return True
                } else {
                    var current = Tp
                    for current != nil {
                        if current.__Parent == U.__Id {
                            return True
                        }
                        if current.__Parent != -1 {
                            var next = ctx.GetType(current.__Parent)
                            Assert(next.__Kind == TK_Plain, "Type: bad parent")
                            current = (*T_Plain)(unsafe.Pointer(next))
                        } else {
                            current = nil
                        }
                    }
                    return False
                }
            } else {
                return False
            }
        } else if U.__Kind == TK_Union {
            return (*T_Union)(unsafe.Pointer(U)).HasSubType(T, ctx)
        } else {
            return False
        }
    case TK_Function:
        if U.__Kind == TK_Plain {
            var Up = (*T_Plain)(unsafe.Pointer(U))
            if Up.__Category == OC_Function {
                if Up.__Parent == -1 {
                    return True
                } else {
                    return Unknown
                }
            } else {
                return False
            }
        } else if U.__Kind == TK_Function {
            if T.__Id == U.__Id {
                return True
            } else {
                return False
            }
        } else {
            return False
        }
    default:
        panic("impossible branch")
    }
}

func (U *T_Union) HasSubType(T *TypeInfo, ctx *ObjectContext) Triple {
    var super_exists = false
    for _, element := range U.__Elements {
        var E = ctx.GetType(element)
        if T.IsSubTypeOf(E, ctx) == True {
            if !super_exists {
                super_exists = true
            } else {
                return False
            }
        }
    }
    if super_exists {
        // there exists only 1 super type
        return True
    } else {
        if T.__Kind == TK_Singleton {
            return False
        } else {
            return Unknown
        }
    }
}
