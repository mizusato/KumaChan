package vm

import (
	. "kumachan/error"
	"kumachan/loader"
)

type Function struct {
	Code       [] Instruction
	BaseSize   FrameBaseSize
	Name       loader.Symbol
	Source     ErrorPoint
}

type FrameBaseSize struct {
	Context   Short
	Reserved  Short
}
