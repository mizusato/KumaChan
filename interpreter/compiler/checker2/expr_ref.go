package checker2

import (
	"kumachan/interpreter/compiler/checker2/checked"
)


type Ref interface { implRef() }

func (FuncRefs) implRef() {}
type FuncRefs struct {
	Functions  [] *Function
}

func (LocalRef) implRef() {}
type LocalRef struct {
	Binding  *checked.LocalBinding
}

func (LocalRefWithFuncRefs) implRef() {}
type LocalRefWithFuncRefs struct {
	LocalRef  LocalRef
	FuncRefs  FuncRefs
}


