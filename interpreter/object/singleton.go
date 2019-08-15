package object

var Nil = Object {
    __Category: OC_Singleton,
    __Inline64: 0,
}

var Void = Object {
    __Category: OC_Singleton,
    __Inline64: 1,
}

var NotFound = Object {
    __Category: OC_Singleton,
    __Inline64: 2,
}

var Complete = Object {
    __Category: OC_Singleton,
    __Inline64: 3,
}

func NewSingleton (context *ObjectContext, name string) Object {
    var id = context.__GetAtomicTypeId(name)
    return Object {
        __Category: OC_Singleton,
        __Inline64: uint64(id),
    }
}
