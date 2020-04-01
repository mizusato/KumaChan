package common

import "fmt"


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
const ArrayMaxSize = 1 << (ShortSize + LongSize)

type Instruction struct {
	OpCode  OpType
	Arg0    Short
	Arg1    Long
}

type OpType Short
const (
	NOP  OpType  =  iota
	NIL     // [ ___ ]: Load a nil value
	/* Data Transfer */
	GLOBAL  // [   index   ]: Load a global value (constant or function)
	LOAD    // [ _, offset ]: Load a value from (frame base) + offset
	STORE   // [ _, offset ]: Store current value to (frame base) + offset
	/* Sum Type Operations */
	SUM     // [index,  ____ ]: Create a value of a sum type
	JIF     // [index,  dest ]: Jump if Index matches the current value
	JMP     // [narrow, dest ]: Jump unconditionally
	/* Product Type Operations */
	PROD    // [size,  _ ]: Create a value of a product type
	GET     // [index, _ ]: Extract the value of a field
	SET     // [index, _ ]: Perform a functional update on a field
	/* Function Type Operations */
	CTX     // [rec, _ ]: Use the current value as the context of a closure
	CALL    // [___, _ ]: Call a (native)function (pop func, pop arg, push ret)
	/* Array Operations */
	ARRAY   // [ size ]: Create an empty array
	APPEND  // [ ____ ]: Append an element to the created array
)

func (inst Instruction) GetGlobalIndex() uint {
	return (uint(inst.Arg0) << LongSize) + uint(inst.Arg1)
}
func (inst Instruction) GetArraySize() uint {
	return inst.GetGlobalIndex()
}

func GlobalIndex(i uint) (Short, Long) {
	return Short(i >> LongSize), Long(i & ((1 << LongSize) - 1))
}
func ArraySize(n uint) (Short, Long) {
	return GlobalIndex(n)
}

func (inst Instruction) GetOffset() uint {
	return uint(inst.Arg1)
}

func (inst Instruction) GetDestAddr() uint {
	return uint(inst.Arg1)
}

func (inst Instruction) GetShortIndexOrSize() uint {
	return uint(inst.Arg0)
}

func (inst Instruction) GetRawShortIndexOrSize() Short {
	return inst.Arg0
}

func (inst Instruction) String() string {
	switch inst.OpCode {
	case NOP:
		return "NOP"
	case NIL:
		return "NIL"
	case GLOBAL:
		return fmt.Sprintf("GLOBAL %d", inst.GetGlobalIndex())
	case LOAD:
		return fmt.Sprintf("LOAD %d", inst.GetOffset())
	case STORE:
		return fmt.Sprintf("STORE %d", inst.GetOffset())
	case SUM:
		return fmt.Sprintf("SUM %d", inst.GetShortIndexOrSize())
	case JIF:
		return fmt.Sprintf("JIF %d %d",
			inst.GetShortIndexOrSize(), inst.GetDestAddr())
	case JMP:
		if inst.Arg0 != 0 {
			return fmt.Sprintf("JMP NARROW %d", inst.GetDestAddr())
		} else {
			return fmt.Sprintf("JMP %d", inst.GetDestAddr())
		}
	case PROD:
		return fmt.Sprintf("PROD %d", inst.GetShortIndexOrSize())
	case GET:
		return fmt.Sprintf("GET %d", inst.GetShortIndexOrSize())
	case SET:
		return fmt.Sprintf("SET %d", inst.GetShortIndexOrSize())
	case CTX:
		if inst.Arg0 != 0 {
			return "CTX REC"
		} else {
			return "CTX"
		}
	case CALL:
		return "CALL"
	case ARRAY:
		return fmt.Sprintf("ARRAY %d", inst.GetArraySize())
	case APPEND:
		return "APPEND"
	default:
		panic("unknown instruction kind")
	}
}
