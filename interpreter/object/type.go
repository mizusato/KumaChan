package object

import "unsafe"
import ."kumachan/interpreter/assertion"

var O_Object = GetTypeObject(4)
var O_Never = GetTypeObject(5)

type TypeId int

type TypeKind int
const (
    TK_Placeholder TypeKind = iota
    TK_Object
    TK_Never
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
    __LowerBound  int
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
    __Exception     int
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
    __Extensible    bool
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
    __Extensible        bool
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

func __InitSpecialTypes (ctx *ObjectContext) {
    var T_Object = &TypeInfo {
        __Kind: TK_Object,
        __Name: "Object",
        __Initialized: true,
    }
    var T_Never = &TypeInfo {
        __Kind: TK_Never,
        __Name: "Never",
        __Initialized: true,
    }
    ctx.__RegisterType(T_Object)
    Assert (
        T_Object.__Id == UnwrapType(O_Object),
        "Type: invalid default special type initialization",
    )
    ctx.__RegisterType(T_Never)
    Assert (
        T_Never.__Id == UnwrapType(O_Never),
        "Type: invalid default special type initialization",
    )
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
    if U.__Kind == TK_Object {
        return True
    }
    switch T.__Kind {
    case TK_Never:
        return True
    case TK_Placeholder:
        if U.__Kind == TK_Placeholder {
            if T.__Id == U.__Id {
                return True
            } else {
                var T_as_Placeholder = (*T_Placeholder)(unsafe.Pointer(T))
                var T_upper *TypeInfo
                var T_upper_id int
                if T_as_Placeholder.__UpperBound != -1 {
                    T_upper = ctx.GetType(T_as_Placeholder.__UpperBound)
                    if T_upper.__Kind == TK_Placeholder {
                        // placeholder bound
                        T_upper_id = T_upper.__Id
                    } else {
                        // non-placeholder bound
                        T_upper_id = -1
                    }
                } else {
                    // placeholder bound (self)
                    T_upper_id = T.__Id
                }
                var U_as_Placeholder = (*T_Placeholder)(unsafe.Pointer(U))
                var U_lower *TypeInfo
                var U_lower_id int
                if U_as_Placeholder.__LowerBound != -1 {
                    U_lower = ctx.GetType(U_as_Placeholder.__LowerBound)
                    if U_lower.__Kind == TK_Placeholder {
                        // placeholder bound
                        U_lower_id = U_lower.__Id
                    } else {
                        // non-placeholder bound
                        U_lower_id = -1
                    }
                } else {
                    // placeholder bound (self)
                    U_lower_id = U.__Id
                }
                if T_upper_id != -1 && U_lower_id != -1 {
                    if T_upper_id == U_lower_id {
                        return True
                    } else {
                        return False
                    }
                } else if T_upper_id == -1 && U_lower_id == -1 {
                    return T_upper.IsSubTypeOf(U_lower, ctx)
                } else {
                    return False
                }
            }
        } else {
            var T_as_Placeholder = (*T_Placeholder)(unsafe.Pointer(T))
            if T_as_Placeholder.__UpperBound != -1 {
                var T_upper = ctx.GetType(T_as_Placeholder.__UpperBound)
                return T_upper.IsSubTypeOf(U, ctx)
            } else {
                return False
            }
        }
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

func (T *TypeInfo) CanHaveProperSubType(ctx *ObjectContext) bool {
    switch T.__Kind {
    case TK_Placeholder:
        var T_as_Placeholder =(*T_Placeholder)(unsafe.Pointer(T))
        if T_as_Placeholder.__LowerBound != -1 {
            var T_lower = ctx.GetType(T_as_Placeholder.__LowerBound)
            if T_lower.__Kind != TK_Placeholder {
                return T_lower.CanHaveProperSubType(ctx)
            } else {
                // unimplemented branch
                // [T < U, U < T] should be forbidden if implemented
                return true
            }
        } else {
            return true
        }
    case TK_Object:
        return true
    case TK_Never, TK_Singleton, TK_Function:
        return false
    case TK_Schema:
        return (*T_Schema)(unsafe.Pointer(T)).__Extensible
    case TK_Class:
        return (*T_Class)(unsafe.Pointer(T)).__Extensible
    default:  // Plain, Union, Trait, Interface
        return true
    }
}

func (U *T_Union) HasSubType(T *TypeInfo, ctx *ObjectContext) Triple {
    var super_exists = false
    var T_is_atomic = !(T.CanHaveProperSubType(ctx))
    for _, element := range U.__Elements {
        var E = ctx.GetType(element)
        if T.IsSubTypeOf(E, ctx) == True {
            if !super_exists {
                super_exists = true
            } else {
                // contained by more than 1 type
                return False
            }
        }
    }
    if super_exists {
        // there exists only 1 super type
        if T_is_atomic {
            return True
        } else {
            return Unknown
        }
    } else {
        if T_is_atomic {
            return False
        } else {
            return Unknown
        }
    }
}
