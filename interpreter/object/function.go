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
    Panic
)

type Function struct {
    __InfoId     int
    __Context    *Scope
    __Type       int
    __TypeArgs   [MAX_TEMPLATE_ARGS] int
}

type FunctionInfo struct {
    __IsNative       bool
    __Native         func([MAX_ARGS]Object) Result
    __ScopeInfo      ScopeInfo
}
