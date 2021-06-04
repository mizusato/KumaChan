package def


type FunctionEntity struct {
	Code Code
	FunctionEntityInfo
}
type FunctionInfo struct {
	// TODO
}
type FunctionEntityInfo struct {
	FunctionInfo
	ContextLength  LocalSize
}

type FunctionSeed interface { FunctionSeed() }

func (impl FunctionSeedLibraryNative) FunctionSeed() {}
type FunctionSeedLibraryNative struct {
	Id    string
	Info  FunctionInfo
}

func (impl FunctionSeedGeneratedNative) FunctionSeed() {}
type FunctionSeedGeneratedNative struct {
	Data  GeneratedNativeFunctionSeed
	Info  FunctionInfo
}
type GeneratedNativeFunctionSeed interface { GeneratedNativeFunctionSeed() }

func (impl FunctionSeedUsual) FunctionSeed() {}
type FunctionSeedUsual struct {
	Trunk   BranchData
	Static  AddrSpace
}
type BranchData struct {
	InstList   [] Instruction
	ExtIdxMap  ExternalIndexMapping
	Stages     [] Stage
	Branches   [] BranchData
	Info       FunctionEntityInfo
}
func CreateFunctionEntity(seed FunctionSeedUsual) *FunctionEntity {
	var frame_size = uint(0)
	var f = createFunctionEntity(0, &frame_size, seed.Trunk, seed.Static)
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

