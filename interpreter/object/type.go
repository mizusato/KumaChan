package object

import ."kumachan/interpreter/assertion"

type TypeId int

type TypeKind int
const (
    TK_Singleton TypeKind = iota
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
    __TypeInfo      TypeInfo
    __MethodTypes   [] int
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
