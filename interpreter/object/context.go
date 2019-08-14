package object

type EnqueFunction = func(Object, []Object, func())

type ObjectContext struct {
    __NextSingletonId   uint64
    __EnqueCallback     EnqueFunction
}

func NewObjectContext (enque EnqueFunction) *ObjectContext {
    return &ObjectContext {
        __NextSingletonId: 4,
        __EnqueCallback: enque,
    }
}
