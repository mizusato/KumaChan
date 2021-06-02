package def


type Code struct {
	InsSeq  [] Instruction
	Offset  LocalSize
	Length  LocalSize  // len(InsSeq) + len(/* InsSeq of branches */)
	ExtMap  ExternalIndexMapping
	Static  AddrSpace
	Stages  [] Stage
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

type Stage ([] Flow)
func (stage Stage) TheOnlyFlow() Flow {
	if len(stage) != 1 { panic("invalid operation") }
	return stage[0]
}
func (stage Stage) ForEachFlow(f func(Flow)) {
	for _, flow := range stage {
		f(flow)
	}
}

type FunctionEntity struct {
	Code Code
	FunctionEntityInfo
}
type FunctionEntityInfo struct {
	ContextLength  LocalSize
}


