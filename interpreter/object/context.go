package object

import ."../assertion"

type TypeId int
type EnqueFunction = func(Object, []Object, func())

type ObjectContext struct {
    __AtomicTypeNames   []string
    __EnqueCallback     EnqueFunction
}

func NewObjectContext (enque EnqueFunction) *ObjectContext {
    return &ObjectContext {
        __AtomicTypeNames: []string {"Nil","Void","NotFound","Complete"},
        __EnqueCallback: enque,
    }
}

func (ctx *ObjectContext) __GetAtomicTypeId (name string) TypeId {
    var id = len(ctx.__AtomicTypeNames)
    Assert(id+1 > id, "ObjectContext: run out of atomic type id")
    ctx.__AtomicTypeNames = append(ctx.__AtomicTypeNames, name)
    return TypeId(id)
}

func (ctx *ObjectContext) GetAtomicTypeName (id TypeId) string {
    Assert (
        0 <= int(id) && int(id) < len(ctx.__AtomicTypeNames),
        "ObjectContext: unable to get the type name of an invalid id",
    )
    return ctx.__AtomicTypeNames[int(id)]
}
