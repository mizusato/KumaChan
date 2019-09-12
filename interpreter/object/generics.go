package object

type TypeExpr struct {
    __Kind TypeExprKind
}

type TypeExprKind int
const (
    TE_Final TypeExprKind = iota
    TE_Function
    TE_Inflation
)

type FinalTypeExpr struct {
    __TypeExpr  TypeExpr
    __Type      int
}

type FunctionTypeExpr struct {
    __TypeExpr     TypeExpr
    __Parameters   [] *TypeExpr
    __ReturnValue  *TypeExpr
}

type InflationTypeExpr struct {
    __TypeExpr    TypeExpr
    __Arguments   [] *TypeExpr
    __Template    *GenericType
}

type GenericType struct {
    __Kind         GenericTypeKind
    __Name         string
    __Parameters   [] GenericTypeParameter
}

type GenericTypeParameter struct {
    __Name            string
    __HasUpperBound   bool
    __UpperBound      int
}

type GenericTypeKind int
const (
    GT_Union GenericTypeKind = iota
    GT_Trait
    GT_Schema
    GT_Class
    GT_Interface
)

type GenericUnionType struct {
    __GenericType  GenericType
    __Elements     [] *TypeExpr
}

type GenericTraitType struct {
    __GenericType  GenericType
    __Elements     [] *TypeExpr
}

type GenericSchemaType struct {
    __GenericType     GenericType
    __BaseList        [] *TypeExpr
    __OwnFieldList    map[Identifier] GenericSchemaField
}

type GenericSchemaField struct {
    __Type           *TypeExpr
    __HasDefault     bool
    __DefaultValue   Object
}

type GenericClassType struct {
    __GenericType         GenericType
    __BaseClassList       [] *TypeExpr
    __BaseInterfaceList   [] *TypeExpr
    __OwnMethodList       map[Identifier] *GenericClassMethod
}

type GenericClassMethod struct {
    __Type       *TypeExpr
    __Function   *Function
}

type GenericInterfaceType struct {
    __GenericType  GenericType
    // TODO
}
