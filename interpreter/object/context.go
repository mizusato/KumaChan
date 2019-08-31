package object

import "sync"
import ."../assertion"

type AtomicTypeId uint64

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
    __Mutex                sync.Mutex
    __IdPool               *IdPool
    __NextAtomicTypeId     AtomicTypeId
    __SingletonTypeInfo    map[AtomicTypeId]*TypeInfo
    __NativeClassList      []NativeClass
    __EnqueCallback        CallbackEnquer
}

func NewObjectContext (enque CallbackEnquer) *ObjectContext {
    var ctx = &ObjectContext {
        __IdPool:             NewIdPool(),
        __SingletonTypeInfo:  make(map[AtomicTypeId]*TypeInfo),
        __NativeClassList:    make([]NativeClass, 0),
        __EnqueCallback:      enque,
    }
    __InitDefaultSingletonTypes(ctx)
    return ctx
}

func (ctx *ObjectContext) GetId(name string) Identifier {
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    return ctx.__IdPool.GetId(name)
}

func (ctx *ObjectContext) GetName(id Identifier) string {
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    return ctx.__IdPool.GetString(id)
}

func (ctx *ObjectContext) __GetNewAtomicTypeId() AtomicTypeId {
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    var id = ctx.__NextAtomicTypeId
    Assert(id+1 > id, "ObjectContext: run out of atomic type id")
    ctx.__NextAtomicTypeId = id + 1
    return id
}

func (ctx *ObjectContext) __RegisterSingleton(name string) Object {
    var id = ctx.__GetNewAtomicTypeId()
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    var t = &TypeInfo {
        __Kind: TK_Singleton,
        __Name: name,
        T_Singleton: T_Singleton {
            __Id: id,
        },
    }
    ctx.__SingletonTypeInfo[id] = t
    return Object {
        __Category: OC_Singleton,
        __Inline: uint64(id),
    }
}

func (ctx *ObjectContext) __GetSingletonTypeInfo(id AtomicTypeId) *TypeInfo {
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    var t, exists = ctx.__SingletonTypeInfo[id]
    Assert(exists, "ObjectContext: invalid singleton type id")
    return t
}

func (ctx *ObjectContext) __RegisterNativeClass (
    name      string,
    methods   NativeClassMethodList,
) *NativeClass {
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    var id = len(ctx.__NativeClassList)
    Assert(id+1 > id, "ObjectContext: run out of native class id")
    ctx.__NativeClassList = append(ctx.__NativeClassList, NativeClass {
        __Name: name,
        __Id: NativeClassId(id),
        __Methods: methods,
    })
    return &ctx.__NativeClassList[id]
}