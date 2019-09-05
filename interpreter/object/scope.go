package object

type Scope struct {
    __Context         *Scope
    __Vals            ValObjectChunk
    __Refs            RefObjectChunk
    __MountOperator   *Function
    __PushOperator    *Function
    __Info            *ScopeInfo
}

type ScopeInfo struct {
    Variables  []VariableInfo
}

type VariableInfo struct {
    Depth      int
    Type       *TypeInfo
    IsMutable  bool
    IsRef      bool
    Offset     int
}

func NewScope (info *ScopeInfo, context *Scope) *Scope {
    var s = &Scope { __Info: info, __Context: context }
    // TODO: setFinalizer()
    return s
}

func (s *Scope) GetContext() *Scope {
    return s.__Context
}
