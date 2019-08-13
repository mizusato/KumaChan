package object

type Scope struct {
    __Bitmap           uint64
    __Context          *Scope
    __Immutable        bool
    __InlineValsCount  uint8
    __InlineRefsCount  uint8
    __InlineVals       [VAL_INLINE_MAX] ValEntry
    __InlineRefs       [REF_INLINE_MAX] RefEntry
    __Vals             map[Identifier] Variable
    __Refs             map[Identifier] *Variable
    __MountOperator    Object
    __PushOperator     Object
}

const VAL_INLINE_MAX = 8
const REF_INLINE_MAX = 4

type Identifier uint64

type Variable struct {
    __Fixed   bool
    __Type    Object
    __Value   Object
}

type ValEntry struct {
    __Name  Identifier
    __Val   Variable
}

type RefEntry struct {
    __Name  Identifier
    __Ref   *Variable
}
