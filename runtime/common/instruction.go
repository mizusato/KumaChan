package common


type Short = uint8
type Long = uint16
const ShortSize = 8
const LongSize = 16
const SumMaxBranches = 1 << ShortSize
const ProductMaxSize = 1 << ShortSize
const ClosureMaxSize = 1 << ShortSize
const FunCodeMaxLength = 1 << LongSize
const LocalSlotMaxSize = 1 << LongSize
const GlobalSlotMaxSize = 1 << (ShortSize + LongSize)

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
	LOAD    // [__, Offset]: Load a value from BaseAddr + Offset
	STORE   // [__, Offset]: Store current value to BaseAddr + Offset
	/* Sum Type Operations */
	SUM     // [Index,    ___  ]: Create a value of a sum type
	JIF     // [Index, DestAddr]: Jump if Index matches the current value
	JMP     // [____,  DestAddr]: Jump unconditionally
	/* Product Type Operations */
	PROD    // [Size,  _ ]: Create a value of a product type
	GET     // [Index, _ ]: Extract the value of a field
	SET     // [Index, _ ]: Perform a functional update on a field
	/* Function Type Operations */
	CTX     // [Rec, _ ]: Use the current value as the context of a closure
	CALL    // [___, _ ]: Call a (native)function (pop func, pop arg, push ret)
)

func (inst Instruction) GetRegIndex() uint {
	return (uint(inst.Arg0) << LongSize) + uint(inst.Arg1)
}

func RegIndex(i uint) (Short, Long) {
	return Short(i & ((1 << LongSize) - 1)), Long(i >> LongSize)
}

func (inst Instruction) GetOffset() uint {
	return uint(inst.Arg1)
}

func (inst Instruction) GetDestAddr() uint {
	return uint(inst.Arg1)
}

func (inst Instruction) GetIndexOrSize() uint {
	return uint(inst.Arg0)
}

func (inst Instruction) GetShortIndexOrSize() Short {
	return inst.Arg0
}
