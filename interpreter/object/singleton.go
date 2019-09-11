package object

var Nil = Object {
    __Category: OC_Type,
    __Inline: 0,
}

var Void = Object {
    __Category: OC_Type,
    __Inline: 1,
}

var NotFound = Object {
    __Category: OC_Type,
    __Inline: 2,
}

var Complete = Object {
    __Category: OC_Type,
    __Inline: 3,
}

func __InitDefaultSingletonTypes (context *ObjectContext) {
    NewSingleton(context, "Nil")
    NewSingleton(context, "Void")
    NewSingleton(context, "NotFound")
    NewSingleton(context, "Complete")
}

func NewSingleton (context *ObjectContext, name string) Object {
    var T = & TypeInfo {
        __Kind: TK_Singleton,
        __Name: name,
    }
    context.__RegisterType(T)
    return Object {
        __Category: OC_Type,
        __Inline: uint64(T.__Id),
    }
}
