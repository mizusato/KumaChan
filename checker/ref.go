package checker

import "kumachan/loader"


func (impl RefConst) ExprVal() {}
type RefConst struct {
	Name  loader.Symbol
}

func (impl RefFunction) ExprVal() {}
type RefFunction struct {
	Name   string
	Index  uint
}

func (impl RefLocal) ExprVal() {}
type RefLocal struct {
	Name  string
}

func (impl NativeFunction) ExprVal() {}
type NativeFunction struct {
	Function  interface {}
}
