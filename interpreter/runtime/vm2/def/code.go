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
type RawCodeBranch struct {
	content   [] Instruction
	external  ExternalIndexMapping
	info      FunctionEntityInfo
	branches  [] RawCodeBranch
}
func CreateFunctionEntity(trunk RawCodeBranch, static AddrSpace) *FunctionEntity {
	var frame_size = uint(0)
	var f = createFunctionEntity(0, &frame_size, trunk, static)
	f.Code.frameSize = LocalSize(frame_size)
	return f
}
func createFunctionEntity(offset uint, fs *uint, this RawCodeBranch, static AddrSpace) *FunctionEntity {
	var required_fs = offset + uint(len(this.content))
	if required_fs >= MaxFrameValues { panic("frame too big") }
	if required_fs > *fs {
		*fs = required_fs
	}
	var branch_entities = make([] *FunctionEntity, len(this.branches))
	for i, b := range this.branches {
		branch_entities[i] = createFunctionEntity(required_fs, fs, b, static)
	}
	return &FunctionEntity {
		Code: Code {
			instOffset: LocalSize(offset),
			instList:   this.content,
			instExtMap: this.external,
			branches:   branch_entities,
			frameSize:  0,
			static:     static,
			stages:     analyzeStages(offset, this.content),
		},
		FunctionEntityInfo: this.info,
	}
}
func analyzeStages(offset uint, instList ([] Instruction)) ([] Stage) {
	panic("not implemented") // TODO
}

