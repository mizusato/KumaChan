package object

import "unsafe"
import ."kumachan/interpreter/assertion"

type Triple int
const (
    False Triple = iota
    True
    Unknown
)

const __TypeCannotCompare = "Type: cannot deeply compare uninitialized types"


func (T *TypeInfo) Equals(U *TypeInfo) bool {
	return T.__Id == U.__Id
}

func (T *TypeInfo) IsIsomorphicTo(U *TypeInfo, ctx *ObjectContext) Triple {
    Assert(T.__Initialized, __TypeCannotCompare)
    Assert(U.__Initialized, __TypeCannotCompare)
    if T.__Id == U.__Id {
        // Fast Path
        return True
    } else {
        // Check if T ⊂ U and U ⊂ T
        var p = T.IsSubTypeOf(U, ctx)
        var q = U.IsSubTypeOf(T, ctx)
        if p == True && q == True {
            return True
        } else if p == Unknown || q == Unknown {
            return Unknown
        } else if p == False || q == False {
            return False
        } else {
            panic("impossible branch")
        }
    }
}

func (T *TypeInfo) IsSubTypeOf(U *TypeInfo, ctx *ObjectContext) Triple {
    Assert(T.__Initialized, __TypeCannotCompare)
    Assert(U.__Initialized, __TypeCannotCompare)
    /* Check if T ⊂ U */
    // (1) T = ∅  or  U = Ω
    if T.__Kind == TK_Never || U.__Kind == TK_Object {
        return True
    }
    // (2) Dealing with Placeholders
    if T.__Kind == TK_Placeholder && U.__Kind != TK_Placeholder {
        // T is Placeholder, U is NOT
        var T_as_Placeholder = (*T_Placeholder)(unsafe.Pointer(T))
        if T_as_Placeholder.__UpperBound != -1 {
            var T_upper = ctx.GetType(T_as_Placeholder.__UpperBound)
            return T_upper.IsSubTypeOf(U, ctx)
        } else {
            // True situation already handled at (1)
            return False
        }
    } else if T.__Kind != TK_Placeholder && U.__Kind == TK_Placeholder {
        // U is Placeholder, T is NOT
        var U_as_Placeholder = (*T_Placeholder)(unsafe.Pointer(U))
        if U_as_Placeholder.__LowerBound != -1 {
            var U_lower = ctx.GetType(U_as_Placeholder.__LowerBound)
            return T.IsSubTypeOf(U_lower, ctx)
        } else {
            // True situation already handled at (1)
            return False
        }
    } else if T.__Kind == TK_Placeholder && U.__Kind == TK_Placeholder {
        // T and U are both placeholder
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
    }
    // (3) Ordinary Cases
    switch T.__Kind {
    case TK_Object:
        // True situation already handled at (1)
        return False
    case TK_Never:
        panic("handled branch")
    case TK_Placeholder:
        panic("handled branch")
    case TK_Singleton:
        if U.__Kind == TK_Plain {
            var U_as_Plain = (*T_Plain)(unsafe.Pointer(U))
            if U_as_Plain.ContainsEntireCategory(OC_Type) {
                return True
            } else {
                return False
            }
        } else if U.__Kind == TK_Singleton {
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
            var T_as_Plain = (*T_Plain)(unsafe.Pointer(T))
            var U_as_Plain = (*T_Plain)(unsafe.Pointer(U))
            if T_as_Plain.__Category == U_as_Plain.__Category {
                if U_as_Plain.__Parent == -1 {
                    return True
                } else {
                    var current = T_as_Plain.__Parent
                    for current != -1 {
                        var C = ctx.GetType(current)
                        Assert(C.__Kind == TK_Plain, "Type: bad parent type")
                        if current == U.__Id {
                            return True
                        }
                        current = (*T_Plain)(unsafe.Pointer(C)).__Parent
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
            var U_as_Plain = (*T_Plain)(unsafe.Pointer(U))
            if U_as_Plain.ContainsEntireCategory(OC_Function) {
                return True
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
    case TK_Union:
        var T_as_Union = (*T_Union)(unsafe.Pointer(T))
        if U.__Kind == TK_Union {
            if T.__Id == U.__Id {
                return True
            } else {
                var U_as_Union = (*T_Union)(unsafe.Pointer(U))
                var ok = __DoSortedIntSlicesHaveContainingRelationship (
                    T_as_Union.__Elements,
                    U_as_Union.__Elements,
                )
                if ok {
                    return True
                }
                // else: unimplemented, fallthrough to rough check
            }
        }
        // Rough Check
        var all_contained = true
        for _, element := range T_as_Union.__Elements {
            var E = ctx.GetType(element)
            if E.IsSubTypeOf(U, ctx) != True {
                all_contained = false
                break
            }
        }
        if all_contained {
            return True
        } else {
            return Unknown
        }
    case TK_Trait:
        var T_as_Trait = (*T_Trait)(unsafe.Pointer(T))
        if U.__Kind == TK_Plain {
            var U_as_Plain = (*T_Plain)(unsafe.Pointer(U))
            if U_as_Plain.ContainsEntireCategory(OC_Instance) {
                return True
            } else {
                return False
            }
        } else if U.__Kind == TK_Trait {
            if T.__Id == U.__Id {
                return True
            } else {
                var U_as_Trait = (*T_Trait)(unsafe.Pointer(U))
                var ok = __DoSortedIntSlicesHaveContainingRelationship (
                    U_as_Trait.__Constraints,
                    T_as_Trait.__Constraints,
                )
                if ok {
                    return True
                }
                // else: unimplemented, fallthrough to rough check
            }
        }
        // Rough Check
        var some_contained = false
        for _, constraint := range T_as_Trait.__Constraints {
            var C = ctx.GetType(constraint)
            if C.IsSubTypeOf(U, ctx) == True {
                some_contained = true
                break
            }
        }
        if some_contained {
            return True
        } else {
            return Unknown
        }
    case TK_Schema:
        var T_as_Schema = (*T_Schema)(unsafe.Pointer(T))
        if U.__Kind == TK_Plain {
            var U_as_Plain = (*T_Plain)(unsafe.Pointer(U))
            if U_as_Plain.ContainsEntireCategory(OC_Struct) {
                return True
            } else {
                return False
            }
        } else if U.__Kind == TK_Schema {
            for _, super := range T_as_Schema.__Supers {
                if super == U.__Id {
                    return True
                }
            }
            return False
        } else {
            return False
        }
    case TK_Class:
        var T_as_Class = (*T_Class)(unsafe.Pointer(T))
        if U.__Kind == TK_Plain {
            var U_as_Plain = (*T_Plain)(unsafe.Pointer(U))
            if U_as_Plain.ContainsEntireCategory(OC_Instance) {
                return True
            } else {
                return False
            }
        } else if U.__Kind == TK_Class {
            for _, super := range T_as_Class.__SuperClasses {
                if super == U.__Id {
                    return True
                }
            }
            return False
        } else if U.__Kind == TK_Interface {
            for _, super := range T_as_Class.__SuperInterfaces {
                if super == U.__Id {
                    return True
                }
            }
            return False
        } else {
            return False
        }
    case TK_Interface:
        return False
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

func (P *T_Plain) ContainsEntireCategory(oc ObjectCategory) bool {
    return P.__Category == oc && P.__Parent == -1
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
