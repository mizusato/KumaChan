package object

const SCOPE_INLINE_MAX = 16

type Scope struct {
    __Context          *Scope
    __InlineCount      int
    __InlineVals       [SCOPE_INLINE_MAX] __VariableEntry
    __Vals             map[Identifier] __Variable
    __Refs             map[Identifier] *__Variable
    __MountOperator    *Function
    __PushOperator     *Function
}

type __Variable struct {
    __Value   Object
    __Type    *TypeInfo
}

type __VariableEntry struct {
    __Name    Identifier
    __Value   __Variable
}
