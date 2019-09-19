package object

import "fmt"
import "sync"
import "strconv"
import "strings"
import ."kumachan/interpreter/assertion"


type ObjectContext struct {
    __Mutex                sync.Mutex
    __IdPool               *IdPool
    __NativeClassList      [] NativeClass
	__TypeInfoList         [] *TypeInfo
    __GenericTypeList      [] *GenericType
    __InflatedTypes        map[string] int
}

func NewObjectContext () *ObjectContext {
    var ctx = &ObjectContext {
        __IdPool:              NewIdPool(),
        __NativeClassList:     make([] NativeClass, 0),
        __TypeInfoList:        make([] *TypeInfo, 0),
        __GenericTypeList:     make([] *GenericType, 0),
        __InflatedTypes:       make(map[string] int),
    }
    __InitDefaultSingletonTypes(ctx)
    __InitSpecialTypes(ctx)
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
    Assert(T != nil, "ObjectContext: invalid TypeInfo: nil")
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

func (ctx *ObjectContext) GetTypeName(id int) string {
    return ctx.GetType(id).__Name
}

func (ctx *ObjectContext) __RegisterGenericType(G *GenericType) {
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    Assert(G != nil, "ObjectContext: invalid generic type: nil")
    var id = len(ctx.__GenericTypeList)
    Assert(id+1 > id, "ObjectContext: run out generic type id")
    G.__Id = id
    ctx.__GenericTypeList = append(ctx.__GenericTypeList, G)
}

func (ctx *ObjectContext) GetGenericType(id int) *GenericType {
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    Assert (
		0 <= id && id < len(ctx.__GenericTypeList),
		"ObjectContext: invalid type id",
	)
    return ctx.__GenericTypeList[id]
}

func (ctx *ObjectContext) __RegisterInflatedType (
    T *TypeInfo, g int, args []int,
) {
    ctx.__RegisterType(T)
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    var key = GetInflatedTypeKey(g, args)
    var _, exists = ctx.__InflatedTypes[key]
    Assert(!exists, "ObjectContext: duplicate registration of an inflated type")
    ctx.__InflatedTypes[key] = T.__Id
}

func (ctx *ObjectContext) GetInflatedType(g int, args []int) (int, bool) {
    ctx.__Mutex.Lock()
    defer ctx.__Mutex.Unlock()
    var key = GetInflatedTypeKey(g, args)
    var id, exists = ctx.__InflatedTypes[key]
    if exists {
        return id, true
    } else {
        return -1, false
    }
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

func GetInflatedTypeKey(g int, args []int) string {
    var arg_id_list = make([]string, len(args))
    for i, arg := range args {
        arg_id_list[i] = strconv.Itoa(arg)
    }
    return fmt.Sprintf("%v[%v]", g, strings.Join(arg_id_list,","))
}
