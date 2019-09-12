package object

import ."kumachan/interpreter/assertion"

type TypeId int

type TypeKind int
const (
    TK_Singleton TypeKind = iota
    TK_Plain
    TK_Function
    TK_Schema
    TK_Class
    TK_Interface
    TK_Trait
    TK_Union
)

type TypeInfo struct {
    __Kind            TypeKind
    __Id              int
    __Name            string
    __IsInitialized   bool
}

type T_Plain struct {
    __TypeInfo   TypeInfo
    __Category   ObjectCategory
    __Checker    func(Object)bool
    __Parent     int
}

type T_Schema struct {
    __TypeInfo      TypeInfo
    __Bases         [] int
    __Supers        [] int
    __Immutable     bool
    __Fields        map[Identifier] int
    __DefaultVals   map[Identifier] Object
}

type T_Class struct {
    __TypeInfo          TypeInfo
    __BaseClasses       [] int
    __BaseInterfaces    [] int
    __SuperClasses      [] int
    __SuperInterfaces   [] int
    __Methods           map[Identifier] __MethodInfo
}

type __MethodInfo struct {
    __From      int
    __Offset    int
    __Function  *Function
}

type T_Interface struct {
    __TypeInfo   TypeInfo
    // TODO
}

type T_Singnature struct {
    __TypeInfo   TypeInfo
    // TODO
}

type T_Compound struct {
    __TypeInfo  TypeInfo
    // TODO
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
