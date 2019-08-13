package object

type ObjectContext struct {
    __NextSingletonId   uint64
    __EnqueCallback     func(Object,[]Object)
}

func NewObjectContext (enque_callback func(Object,[]Object)) *ObjectContext {
    return &ObjectContext {
        __NextSingletonId: 4,
        __EnqueCallback: enque_callback,
    }
}
