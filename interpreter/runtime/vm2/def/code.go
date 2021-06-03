package def


type Code struct {
	instOffset  LocalSize
	instList    [] Instruction
	instExtMap  ExternalIndexMapping
	branches    [] *FunctionEntity
	frameSize   LocalSize // len(instList) + len(/* instList of branches */)
	static      AddrSpace
	stages      [] Stage
}
func (code *Code) FrameSize() LocalSize {
	return code.frameSize
}
func (code *Code) Inst(i LocalAddr) Instruction {
	return code.instList[i]
}
func (code *Code) InstCount() LocalSize {
	return LocalSize(len(code.instList))
}
func (code *Code) InstDstAddr(i LocalAddr) LocalAddr {
	return (i + code.instOffset)
}
func (code *Code) Static(s LocalAddr) Value {
	return code.static[s]
}
func (code *Code) Stages() ([] Stage) {
	return code.stages
}
func (code *Code) ChooseBranch(ptr ExternalIndexMapPointer, vec ShortIndexVector) uint {
	return code.instExtMap.ChooseBranch(ptr, vec)
}
func (code *Code) BranchFuncValue(index uint) UsualFuncValue {
	return &ValFunc {
		Entity:  code.branches[index],
		Context: nil,
	}
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

func CreateFunctionEntity (
	trunk     [] Instruction,
	info      FunctionEntityInfo,
	branches  [] [] Instruction,
	b_info    [] FunctionEntityInfo,
	static    AddrSpace,
) *FunctionEntity {
	panic("not implemented") // TODO
}

func analyzeStages(instList ([] Instruction)) ([] Stage) {
	panic("not implemented") // TODO
}

