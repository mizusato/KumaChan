package object

const MAX_ARGS = 8
const MAX_TEMPLATE_ARGS = 5

type Function struct {
    __FunInfo    int
    __Context    *Scope
    __Template   int
    __TypeArgs   [MAX_TEMPLATE_ARGS] int
    __PausePos   int
}

type FunctionInfo struct {
    __IsNative    bool
    __Native      func([MAX_ARGS]Object) Object
    __ByteCode    int
    __ScopeInfo   ScopeInfo
}

