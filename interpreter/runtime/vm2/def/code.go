package def


type Code struct {
	dstOffset  LocalSize
	instList   [] Instruction
	extIdxMap  ExternalIndexMapping
	stages     [] Stage
	branches   [] *FunctionEntity
	frameSize  LocalSize // len(instList) + len(/* instList of branches */)
	static     AddrSpace
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
	return code.static[s]
}
func (code *Code) Stages() ([] Stage) {
	return code.stages
}
func (code *Code) ChooseBranch(ptr ExternalIndexMapPointer, vec ShortIndexVector) uint {
	return code.extIdxMap.ChooseBranch(ptr, vec)
}
func (code *Code) BranchFuncValue(index uint) UsualFuncValue {
	return &ValFunc {
		Entity:  code.branches[index],
		Context: nil,
	}
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

type FunctionEntity struct {
	Code Code
	FunctionEntityInfo
}
type FunctionEntityInfo struct {
	ContextLength  LocalSize
}
type BranchData struct {
	InstList   [] Instruction
	ExtIdxMap  ExternalIndexMapping
	Stages     [] Stage
	Branches   [] BranchData
	Info       FunctionEntityInfo
}
func CreateFunctionEntity(trunk BranchData, static AddrSpace) *FunctionEntity {
	var frame_size = uint(0)
	var f = createFunctionEntity(0, &frame_size, trunk, static)
	f.Code.frameSize = LocalSize(frame_size)
	return f
}
func createFunctionEntity(offset uint, fs *uint, this BranchData, static AddrSpace) *FunctionEntity {
	var required_fs = offset + uint(len(this.InstList))
	if required_fs >= MaxFrameValues { panic("frame too big") }
	if required_fs > *fs {
		*fs = required_fs
	}
	var branch_entities = make([] *FunctionEntity, len(this.Branches))
	for i, b := range this.Branches {
		branch_entities[i] = createFunctionEntity(required_fs, fs, b, static)
	}
	return &FunctionEntity {
		Code: Code {
			dstOffset: LocalSize(offset),
			instList:  this.InstList,
			extIdxMap: this.ExtIdxMap,
			stages:    this.Stages,
			branches:  branch_entities,
			frameSize: 0,
			static:    static,
		},
		FunctionEntityInfo: this.Info,
	}
}

