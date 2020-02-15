package common

import (
	. "kumachan/error"
	"kumachan/loader"
)

type Function struct {
	Code        [] Instruction
	BaseSize    FrameBaseSize
	SourceInfo  SourceInfo
}

type FrameBaseSize struct {
	Context   Short
	Reserved  Short
}

type SourceInfo struct {
	Name       loader.Symbol
	DeclPoint  ErrorPoint
	CodeMap    [] ErrorPoint
}
