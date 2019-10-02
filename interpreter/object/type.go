package object

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
    TK_NativeClass
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
    __Items       [] FunctionTypeItem
}

type FunctionTypeItem struct {
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

type T_NativeClass struct {
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
