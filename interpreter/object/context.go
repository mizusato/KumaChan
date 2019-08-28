package object

import ."../assertion"

type TypeId int

type CallbackPriority int
const (
    Low  CallbackPriority = iota
    High
)

type Callback struct {
    IsGroup    bool
    Argc       int
    Argv       [MAX_ARGS]Object
    Callee     Object
    Callees    []Object
    Feedback   func()
}

type CallbackEnquer = func(Callback, CallbackPriority)

type ObjectContext struct {
    // TODO: add mutex
    __AtomicTypeNames      []string
    __EnqueCallback        CallbackEnquer
}

func NewObjectContext (enque CallbackEnquer) *ObjectContext {
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
