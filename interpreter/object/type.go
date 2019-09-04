package object


type TypeKind int
const (
    TK_Category TypeKind = iota
    TK_PlainSubSet
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
    __T_Category      T_Category
    __T_PlainSubSet   T_PlainSubSet
    __T_Singleton     T_Singleton
    __T_Schema        T_Schema
    __T_Class         T_Class
    __T_Interface     T_Interface
    __T_Singnature    T_Singnature
    __T_Compound      T_Compound
}


type T_Category struct {
    __Id         AtomicTypeId
    __Category   ObjectCategory
}

type T_PlainSubSet struct {
    __Id         AtomicTypeId
    __Category   ObjectCategory
    __Checker    func(Object)bool
    __Parent     *TypeInfo
}

type T_Singleton struct {
    __Id         AtomicTypeId
}

type T_Schema struct {
    __Id         AtomicTypeId
    // TODO
}

type T_Class struct {
    __Id         AtomicTypeId
    // TODO
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
