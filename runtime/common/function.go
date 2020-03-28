package common

import (
	. "kumachan/error"
	"kumachan/transformer/node"
)

type Function struct {
	IsNative     bool
	NativeIndex  int
	Code         [] Instruction
	BaseSize     FrameBaseSize
	Info         FuncInfo
}

type FrameBaseSize struct {
	Context   Short
	Reserved  Long
}

type FuncInfo struct {
	Name       string
	DeclPoint  ErrorPoint
	SourceMap  [] *node.Node
}

func (f *Function) ToValue(native_registry func(int)Value) Value {
	if f.IsNative {
		return native_registry(f.NativeIndex)
	} else {
		return FunctionValue {
			Underlying:    f,
			ContextValues: make([]Value, 0, 0),
		}
	}
}
