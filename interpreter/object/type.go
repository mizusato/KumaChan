package object

type TypeKind int
const (
    TK_Uninitialized TypeKind = iota
    TK_Plain
    TK_Singleton
    TK_Schema
    TK_Class
    TK_Interface
    TK_Singnature
    TK_Compound
)

type TypeInfo struct {
    __Kind            TypeKind
    __Name            string
    __T_Plain         *T_Plain
    __T_Singleton     *T_Singleton
    __T_Schema        *T_Schema
    __T_Class         *T_Class
    __T_Interface     *T_Interface
    __T_Singnature    *T_Singnature
    __T_Compound      *T_Compound
}

type T_Plain struct {
    __Id         AtomicTypeId
    __Category   ObjectCategory
    __Checker    func(Object)bool
    __Parent     *TypeInfo
}

type T_Singleton struct {
    __Id         AtomicTypeId
}

type T_Schema struct {
    __Id            AtomicTypeId
    __Bases         [] *TypeInfo
    __Supers        [] *TypeInfo
    __Immutable     bool
    __Fields        map[Identifier] *TypeInfo
    __DefaultVals   map[Identifier] Object
    __Operators     map[CustomOperator] *Function
}

type T_Class struct {
    __Id                AtomicTypeId
    __BaseClasses       [] *TypeInfo
    __BaseInterfaces    [] *TypeInfo
    __SuperClasses      [] *TypeInfo
    __SuperInterfaces   [] *TypeInfo
    __Methods           map[Identifier] __MethodInfo
    __Operators         map[CustomOperator] *Function
}

type __MethodInfo struct {
    __From      *TypeInfo
    __Function  *Function
    __Offset    int
}

type T_Interface struct {
    __Id         AtomicTypeId
    // TODO
}

type T_Singnature struct {
    __Id         AtomicTypeId
    // TODO
}

type T_Compound struct {
    // TODO
}
