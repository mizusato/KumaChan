package compiler

import . "kumachan/error"


type Error struct {
	Point     ErrorPoint
	Concrete  ConcreteError
}

type ConcreteError interface { CompilerError() }

func (impl E_NativeFunctionNotFound) CompilerError() {}
type E_NativeFunctionNotFound struct {
	Name  string
}

func (impl E_NativeConstantNotFound) CompilerError() {}
type E_NativeConstantNotFound struct {
	Name  string
}

func (impl E_CircularConstantDependency) CompilerError() {}
type E_CircularConstantDependency struct {
	Constants  [] string
}
