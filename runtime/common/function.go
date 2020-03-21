package common

import (
	. "kumachan/error"
	"kumachan/loader"
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
	Reserved  Short
}

type FuncInfo struct {
	Name       loader.Symbol
	DeclPoint  ErrorPoint
	SourceMap  [] *node.Node
}

func (f *Function) ToValue(native_index []NativeFunction) Value {
	if f.IsNative {
		return NativeFunctionValue(native_index[f.NativeIndex])
	} else {
		return FunctionValue{
			Underlying:    f,
			ContextValues: make([]Value, 0, 0),
		}
	}
}
