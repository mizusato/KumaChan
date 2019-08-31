package object

const MAX_ARGS = 8
const MAX_TEMPLATE_ARGS = 4

type Result struct {
    Type   ResultType
    Value  Object
}

type ResultType int
const (
    Return ResultType = iota
    Throw
)
