package object

import "runtime"
import ."kumachan/interpreter/assertion"

type Scope struct {
    __Context         *Scope
    __Vals            ValObjectChunk
    __Refs            RefObjectChunk
    __MountOperator   *Function
    __PushOperator    *Function
    __Recycled        bool
}

type ScopeInfo struct {
    Context    *ScopeInfo
    Variables  map[Identifier] VariableInfo
}

type VariableInfo struct {
    Type       *TypeInfo
    IsMutable  bool
    IsVal      bool
    Offset     int
}


func NewScope (v_count int, r_count int, context *Scope) *Scope {
    var s = &Scope {
        __Context: context,
        __Vals: GetValObjectChunk(v_count),
        __Refs: GetRefObjectChunk(r_count),
    }
    runtime.SetFinalizer(s, func (s *Scope) {
        s.Recycle()
    })
    return s
}

func (s *Scope) Recycle() {
    if !s.__Recycled {
        RecycleValObjectChunk(s.__Vals)
        RecycleRefObjectChunk(s.__Refs)
        *s = Scope { __Recycled: true }
    }
}

func (s *Scope) __EnsureNotRecycled() {
    Assert(!s.__Recycled, "Scope: invalid use of recycled scope")
}

func (s *Scope) GetContext() *Scope {
    s.__EnsureNotRecycled()
    return s.__Context
}

func (s *Scope) DeclareVal(object Object) {
    s.__EnsureNotRecycled()
    s.__Vals = append(s.__Vals, object)
}

func (s *Scope) DeclareRefByVal(object Object) {
    s.__EnsureNotRecycled()
    s.__Refs = append(s.__Refs, &object)
}

func (s *Scope) __DeclareRef(object_ref *Object) {
    s.__EnsureNotRecycled()
    s.__Refs = append(s.__Refs, object_ref)
}

func (s *Scope) GetValAtOffset(n int) Object {
    s.__EnsureNotRecycled()
    return s.__Vals[n]
}

func (s *Scope) GetValOfRefAtOffset(n int) Object {
    s.__EnsureNotRecycled()
    return *(s.__Refs[n])
}

func (s *Scope) SetValAtOffset(n int, object Object) {
    s.__EnsureNotRecycled()
    s.__Vals[n] = object
}

func (s *Scope) SetRefByValAtOffset(n int, object Object) {
    s.__EnsureNotRecycled()
    *(s.__Refs[n]) = object
}

func (s *Scope) CaptureValAtOffset(n int, closure *Scope) {
    s.__EnsureNotRecycled()
    closure.DeclareVal(s.__Vals[n])
}

func (s *Scope) CaptureRefAtOffset(n int, closure *Scope) {
    s.__EnsureNotRecycled()
    closure.__DeclareRef(s.__Refs[n])
}
