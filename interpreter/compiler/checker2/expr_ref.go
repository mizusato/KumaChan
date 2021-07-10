package checker2

import (
	"kumachan/interpreter/compiler/checker2/checked"
)


type Ref interface { implRef() }

func (TypeRef) implRef() {}
type TypeRef struct {
	TypeDef  TypeDef
}

func (FuncRefs) implRef() {}
type FuncRefs struct {
	Functions  [] *Function
}

func (LocalRef) implRef() {}
type LocalRef struct {
	Binding  *checked.LocalBinding
}

func (RefWithLocalRef) implRef() {}
type RefWithLocalRef struct {
	Ref       Ref
	LocalRef  LocalRef
}


