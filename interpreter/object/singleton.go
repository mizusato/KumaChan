package object

import ."../assertion"

var Nil = Object {
    __Category: OC_Singleton,
    __Inline: 0,
}

var Void = Object {
    __Category: OC_Singleton,
    __Inline: 1,
}

var NotFound = Object {
    __Category: OC_Singleton,
    __Inline: 2,
}

var Complete = Object {
    __Category: OC_Singleton,
    __Inline: 3,
}

func __InitDefaultSingletonTypes (context *ObjectContext) {
    context.__RegisterSingleton("Nil")
    context.__RegisterSingleton("Void")
    context.__RegisterSingleton("NotFound")
    context.__RegisterSingleton("Complete")
}

func NewSingleton (context *ObjectContext, name string) Object {
    return context.__RegisterSingleton(name)
}

func GetSingletonTypeInfo (context *ObjectContext, object Object) *TypeInfo {
    Assert (
        object.__Category == OC_Singleton,
        "invalid usage of GetSingletonTypeInfo()",
    )
    return context.__GetSingletonTypeInfo(AtomicTypeId(object.__Inline))
}
