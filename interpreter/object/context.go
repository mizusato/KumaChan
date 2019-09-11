package object

import "sync"
import ."kumachan/interpreter/assertion"


type ObjectContext struct {
    __Mutex                sync.Mutex
    __IdPool               *IdPool
	__TypeInfoList         [] *TypeInfo
    __NativeClassList      [] NativeClass
}

func NewObjectContext () *ObjectContext {
    var ctx = &ObjectContext {
        __IdPool:             NewIdPool(),
        __TypeInfoList:       make([] *TypeInfo, 0),
        __NativeClassList:    make([] NativeClass, 0),
    }
    __InitDefaultSingletonTypes(ctx)
	__InitDefaultPlainTypes(ctx)
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

func (ctx *ObjectContext) __RegisterType(T *TypeInfo) {
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    var id = len(ctx.__TypeInfoList)
    Assert(id+1 > id, "ObjectContext: run out of type id")
    T.__Id = id
	ctx.__TypeInfoList = append(ctx.__TypeInfoList, T)
}

func (ctx *ObjectContext) GetType(id int) *TypeInfo {
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    Assert (
		0 <= id && id < len(ctx.__TypeInfoList),
		"ObjectContext: invalid type id",
	)
    return ctx.__TypeInfoList[id]
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
