package base

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
	POP     // [ ___ ]: Discard current value
	/* Data Transfer */
	GLOBAL  // [   index  ]: Load a global value (constant or function)
	LOAD    // [ _, offset]: Load a value from (frame base) + offset
	STORE   // [ _, offset]: Store current value to (frame base) + offset
	/* EnumValue Operations */
	ENUM    // [index, ___ ]: Create a value of a EnumValue
	JIF     // [index, dest]: Jump if index matches the current value
	JMP     // [ ____, dest]: Jump unconditionally
	BR      // [index, ___ ]: Branch functional reference on a EnumValue
	BRB     // [index, ___ ]: Branch functional reference on a branch reference
	BRF     // [index, ___ ]: Branch functional reference on a field reference
	/* TupleValue Operations */
	TUP     // [size,  _ ]: Create a value of a product type
	GET     // [index, _ ]: Extract the value of a field
	POPGET  // [index, _ ]: Replace current value into the value of a field
	SET     // [index, _ ]: Perform a functional update on a field
	FR      // [index, _ ]: Field functional reference on a TupleValue
	FRF     // [index, _ ]: Field functional reference on a field reference
	/* Function Type Operations */
	CTX     // [rec, _ ]: Use the current value as the context of a closure
	CALL    // [ __, _ ]: Call a (native)function (pop func, pop arg, push ret)
	/* Array Operations */
	ARRAY   // [ infoIndex ]: Create an empty array
	APPEND  // [ _________ ]: Append an element to the created array
	/* Multi-Switch Operations */
	MS      // [ _________ ]: Start multi-switch
	MSI     // [index, ___ ]: Append branch element index
	MSD     // [ ____, ___ ]: Append branch element index (default)
	MSJ     // [ ____, dest]: Jump if branch indexes matches the current value
)

func (inst Instruction) GetGlobalIndex() uint {
	return (uint(inst.Arg0) << LongSize) + uint(inst.Arg1)
}

func GlobalIndex(i uint) (Short, Long) {
	return Short(i >> LongSize), Long(i & ((1 << LongSize) - 1))
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

func (inst Instruction) String() string {
	switch inst.OpCode {
	case NOP:
		return "NOP"
	case NIL:
		return "NIL"
	case POP:
		return "POP"
	case GLOBAL:
		return fmt.Sprintf("GLOBAL %d", inst.GetGlobalIndex())
	case LOAD:
		return fmt.Sprintf("LOAD %d", inst.GetOffset())
	case STORE:
		return fmt.Sprintf("STORE %d", inst.GetOffset())
	case ENUM:
		return fmt.Sprintf("ENUM %d", inst.GetShortIndexOrSize())
	case JIF:
		return fmt.Sprintf("JIF %d %d",
			inst.GetShortIndexOrSize(), inst.GetDestAddr())
	case JMP:
		return fmt.Sprintf("JMP %d", inst.GetDestAddr())
	case BR:
		return fmt.Sprintf("BR %d", inst.GetShortIndexOrSize())
	case BRB:
		return fmt.Sprintf("BRB %d", inst.GetShortIndexOrSize())
	case BRF:
		return fmt.Sprintf("BRF %d", inst.GetShortIndexOrSize())
	case TUP:
		return fmt.Sprintf("TUP %d", inst.GetShortIndexOrSize())
	case GET:
		return fmt.Sprintf("GET %d", inst.GetShortIndexOrSize())
	case POPGET:
		return fmt.Sprintf("POPGET %d", inst.GetShortIndexOrSize())
	case SET:
		return fmt.Sprintf("SET %d", inst.GetShortIndexOrSize())
	case FR:
		return fmt.Sprintf("FR %d", inst.GetShortIndexOrSize())
	case FRF:
		return fmt.Sprintf("FRF %d", inst.GetShortIndexOrSize())
	case CTX:
		if inst.Arg0 != 0 {
			return "CTX REC"
		} else {
			return "CTX"
		}
	case CALL:
		return "CALL"
	case ARRAY:
		return fmt.Sprintf("ARRAY %d", inst.GetGlobalIndex())
	case APPEND:
		return "APPEND"
	case MS:
		return "MS"
	case MSI:
		return fmt.Sprintf("MSI %d", inst.GetShortIndexOrSize())
	case MSD:
		return "MSD"
	case MSJ:
		return fmt.Sprintf("MSJ %d", inst.GetDestAddr())
	default:
		panic("unknown instruction kind")
	}
}
