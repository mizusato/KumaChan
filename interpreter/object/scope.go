package object

type Scope struct {
    __Context          *Scope
    __Immutable        bool
    __InlineValsCount  uint8
    __InlineRefsCount  uint8
    __InlineVals       [VAL_INLINE_MAX] __ValEntry
    __InlineRefs       [REF_INLINE_MAX] __RefEntry
    __Vals             map[Identifier] __Variable
    __Refs             map[Identifier] *__Variable
    __MountOperator    Object
    __PushOperator     Object
}

const VAL_INLINE_MAX = 8
const REF_INLINE_MAX = 4

type __Variable struct {
    __IsFixed       bool    // defined by 'let' or 'var'
    __Value         Object  // current value of variable
    __NonFixedType  Object  // if defined by 'var', a type should be stored
}

type __ValEntry struct {
    __Name  Identifier
    __Val   __Variable
}

type __RefEntry struct {
    __Name  Identifier
    __Ref   *__Variable
}
