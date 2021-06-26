package def


type Code struct {
	dstOffset  LocalSize
	instList   [] Instruction
	extMap     ExternalBranchMapping
	stages     [] Stage
	branches   [] *FunctionEntity
	closures   [] *FunctionEntity
	frameSize  LocalSize // len(instList) + len(/* instList of branches */)
	static     [] *Value
	genStatic  *GeneratedStaticData
}
func (code *Code) FrameSize() LocalSize {
	if code.frameSize == 0 { panic("something went wrong") }
	return code.frameSize
}
func (code *Code) Inst(i LocalAddr) Instruction {
	return code.instList[i]
}
func (code *Code) InstCount() LocalSize {
	return LocalSize(len(code.instList))
}
func (code *Code) InstDstAddr(i LocalAddr) LocalAddr {
	return (i + code.dstOffset)
}
func (code *Code) Static(s LocalAddr) Value {
	return *(code.static[s])
}
func (code *Code) Stages() ([] Stage) {
	return code.stages
}
func (code *Code) ChooseBranch(ptr ExternalBranchMapIndex, vec ShortIndexVector) uint {
	return code.extMap.ChooseBranch(ptr, vec)
}
func (code *Code) BranchFuncValue(index uint) UsualFuncValue {
	return &ValFunc {
		Entity:  code.branches[index],
		Context: nil,
	}
}
func (code *Code) ClosureEntity(index LocalAddr) *FunctionEntity {
	return code.closures[index]
}
func (code *Code) DispatchTable(index LocalAddr) *DispatchTable {
	return code.genStatic.DispatchTables[index]
}
func (code *Code) InterfaceTransformPath(index LocalAddr) *InterfaceTransformPath {
	return code.genStatic.TransformPaths[index]
}

type AddrSpace ([] Value)

type Flow struct {
	SimpleFlow
	NestedFlow
}
type SimpleFlow struct {
	Start   LocalAddr
	End     LocalAddr
}
type NestedFlow struct {
	Stages  [] Stage
}
func (flow Flow) Simple() bool {
	return (flow.Stages == nil)
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

