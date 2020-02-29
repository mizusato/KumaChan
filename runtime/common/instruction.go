package common


type Short = uint8
type Long = uint16
const ShortSize = 8
const LongSize = 16
const SumMaxBranches = 1 << ShortSize
const ProductMaxSize = 1 << ShortSize
const FunCodeMaxLength = 1 << LongSize
const RegistryMaxSize = 1 << (ShortSize + LongSize)

type Instruction struct {
	OpCode  OpType
	Arg0    Short
	Arg1    Long
}

type OpType Short
const (
	NOP  OpType  =  iota
	/* Data Transfer */
	GLOBAL  // [ RegIndex ]: Load a global value (constant or function)
	LOAD    // [Offset, __]: Load a value from BaseAddr + Offset
	STORE   // [Offset, __]: Store current value to BaseAddr + Offset
	/* Sum Type Operations */
	SUM     // [Index, ________]: Create a value of a sum type
	JIF     // [Index, JumpAddr]: Jump if Index matches the current value
	JMP     // [_____, JumpAddr]: Jump unconditionally
	/* Product Type Operations */
	PROD    // [Size,  __]: Create a value of a product type
	GET     // [Index, __]: Extract the value of a field
	SET     // [Index, __]: Perform a functional update on a field
	/* Function Type Operations */
	CTX     // [Rec, _]: Use the current value as the context of a closure
	CALL    // [___, _]: Call a (native)function (pop func, pop arg, push ret)
)

func (inst Instruction) GetRegIndex() int {
	return (int(inst.Arg0) << LongSize) + int(inst.Arg1)
}

func (inst Instruction) GetOffset() int {
	return int(inst.Arg0)
}

func (inst Instruction) GetJumpAddr() int {
	return int(inst.Arg1)
}

func (inst Instruction) GetIndexOrSize() int {
	return int(inst.Arg0)
}

func (inst Instruction) GetShortIndexOrSize() Short {
	return inst.Arg0
}
