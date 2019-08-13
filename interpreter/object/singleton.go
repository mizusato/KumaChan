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

func NewSingleton (context *ObjectContext) Object {
    var id = context.__NextSingletonId
    var new_id = id + 1
    if new_id < id {
        panic("run out of singleton object id")
    }
    context.__NextSingletonId = new_id
    return Object {
        __Category: OC_Singleton,
        __Inline64: id,
    }
}
