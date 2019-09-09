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
    TK_TypeTemplate
    TK_FunctionTemplate
    TK_Compound
)

type TypeInfo struct {
    __Kind            TypeKind
    __Name            string
}

type T_Plain struct {
    __TypeInfo   TypeInfo
    __Id         AtomicTypeId
    __Category   ObjectCategory
    __Checker    func(Object)bool
    __Parent     *TypeInfo
}

type T_Singleton struct {
    __TypeInfo   TypeInfo
    __Id         AtomicTypeId
}

type T_Schema struct {
    __TypeInfo      TypeInfo
    __Id            AtomicTypeId
    __Bases         [] *TypeInfo
    __Supers        [] *TypeInfo
    __Immutable     bool
    __Fields        map[Identifier] *TypeInfo
    __DefaultVals   map[Identifier] Object
    __Operators     map[CustomOperator] *Function
}

type T_Class struct {
    __TypeInfo          TypeInfo
    __Id                AtomicTypeId
    __BaseClasses       [] *TypeInfo
    __BaseInterfaces    [] *TypeInfo
    __SuperClasses      [] *TypeInfo
    __SuperInterfaces   [] *TypeInfo
    __Methods           map[Identifier] __MethodInfo
    __Operators         map[CustomOperator] *Function
}

type __MethodInfo struct {
    __TypeInfo  TypeInfo
    __From      *TypeInfo
    __Function  *Function
    __Offset    int
}

type T_Interface struct {
    __TypeInfo   TypeInfo
    __Id         AtomicTypeId
    // TODO
}

type T_Singnature struct {
    __TypeInfo   TypeInfo
    __Id         AtomicTypeId
    // TODO
}

type T_TypeTemplate struct {
    __TypeInfo   TypeInfo
    __Id         AtomicTypeId
    // TODO
}

type T_FunctionTemplate struct {
    __TypeInfo   TypeInfo
    __Id         AtomicTypeId
    // TODO
}

type T_Compound struct {
    __TypeInfo  TypeInfo
    // TODO
}
