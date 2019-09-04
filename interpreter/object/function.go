package object

const MAX_ARGS = 8
const MAX_TEMPLATE_ARGS = 5

type Result struct {
    Kind   ResultKind
    Value  Object
}

type ResultKind int
const (
    Return ResultKind = iota
    Throw
)

type Function struct {
    __Kind              FunctionKind
    __NativeFunction    NativeFunction
    __UserlandFunction  UserlandFunction
}

type NativeFunction struct {
    __Native   func([MAX_ARGS]Object) Result
}

type UserlandFunction struct {
    __BodyId   int
    // TODO
}

type FunctionKind int
const (
    Native FunctionKind = iota
    Userland
)
