package def


type Code struct {
	InsSeq  [] Instruction
	Offset  LocalSize
	Length  LocalSize  // len(InsSeq) + len(/* InsSeq of branches */)
	ExtMap  ExternalIndexMapping
	Static  AddrSpace
	Stages  [] [] Flow
}
func (code *Code) FrameRequiredSize() LocalSize {
	return code.Length
}
func (code *Code) GetSizeInsValue(i LocalAddr) LocalSize {
	return code.InsSeq[i].ToSize()
}

type AddrSpace ([] Value)

type Flow struct {
	Start  LocalAddr
	End    LocalAddr
}

type FunctionEntity struct {
	Code Code
	FunctionEntityInfo
}
type FunctionEntityInfo struct {
	ContextLength  LocalSize
}


