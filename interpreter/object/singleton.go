package object

import ."kumachan/interpreter/assertion"

const __SingletonInvalidInit = "Singleton: invalid default type initialization"

var Nil = GetTypeObject(0)
var Void = GetTypeObject(1)
var NotFound = GetTypeObject(2)
var Complete = GetTypeObject(3)

func __InitDefaultSingletonTypes (context *ObjectContext) {
    var Nil_ = NewSingleton(context, "Nil")
    var Void_ = NewSingleton(context, "Void")
    var NotFound_ = NewSingleton(context, "NotFound")
    var Complete_ = NewSingleton(context, "Complete")
    Assert(Nil_ == Nil, __SingletonInvalidInit)
    Assert(Void_ == Void, __SingletonInvalidInit)
    Assert(NotFound_ == NotFound, __SingletonInvalidInit)
    Assert(Complete_ == Complete, __SingletonInvalidInit)
}

func NewSingleton (context *ObjectContext, name string) Object {
    var T = & TypeInfo {
        __Kind: TK_Singleton,
        __Name: name,
    }
    context.__RegisterType(T)
    return GetTypeObject(T.__Id)
}
