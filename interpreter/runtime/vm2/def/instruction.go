package def

import (
	"fmt"
	"unsafe"
)


type ShortIndex uint8
const ShortIndexSize = unsafe.Sizeof(ShortIndex(0))
type ShortSize = ShortIndex
const ShortSizeMax = (1 << (8 * ShortIndexSize))

type ExternalIndexMapPointer uint16
const ExternalIndexMapPointerSize = unsafe.Sizeof(ExternalIndexMapPointer(0))

type ShortIndexVector uint32
const ShortIndexVectorSize = unsafe.Sizeof(ShortIndexVector(0))
const MaxShortIndexVectorElements = uint(ShortIndexVectorSize / ShortIndexSize)
func CreateShortIndexVectorSingleElement(index ShortIndex) ShortIndexVector {
	return ShortIndexVector(index)
}
func CreateShortIndexVector(indexes ([] ShortIndex)) ShortIndexVector {
	if !(uint(len(indexes)) <= MaxShortIndexVectorElements) {
		panic("something went wrong")
	}
	var vec ShortIndexVector = 0
	for _, index := range indexes {
		vec = (vec << (ShortIndexSize * 8))
		vec = (vec | ShortIndexVector(index))
	}
	return vec
}

type LocalAddr uint16
type LocalSize = LocalAddr
const LocalAddrSize = unsafe.Sizeof(LocalAddr(0))

const MaxEnumCases = (1 << ShortIndexSize)
const MaxTupleElements = (1 << ShortIndexSize)
const MaxBranchExpressions = (1 << ExternalIndexMapPointerSize)
const MaxFrameValues = ((1 << LocalAddrSize) - 1)
const MaxInsSeqLength = ((1 << LocalAddrSize) - 1)
const MaxStaticValues = (1 << LocalAddrSize)
const MaxClosureContexts = (1 << LocalAddrSize)

type Instruction struct {
	OpCode  OpCode
	Idx     ShortIndex
	ExtIdx  ExternalIndexMapPointer
	Obj     LocalAddr
	Src     LocalAddr
}
const InstructionSize = unsafe.Sizeof(Instruction {})
var _ = (func() struct{} {
	if InstructionSize != 8 { panic("incorrect instruction size") }
	return struct{}{}
})()

type ExternalIndexMapping ([] ExternalIndexMap)
type ExternalIndexMap struct {
	HasDefault  bool
	Default     uint
	VectorMap   map[ShortIndexVector] uint
}
func (all ExternalIndexMapping) ChooseBranch (
	ptr  ExternalIndexMapPointer,
	vec  ShortIndexVector,
) uint {
	if !(uint(ptr) < uint(len(all))) { panic("invalid ExtIdx") }
	var m = &all[ptr]
	var branch, found = m.VectorMap[vec]
	if found {
		return branch
	} else {
		if m.HasDefault {
			return m.Default
		} else {
			panic("matching branch not found")
		}
	}
}

type OpCode uint8
const (
	SIZE    OpCode = iota  // whole instruction as an unsigned number
	ARG     // [ ____ ___ ____ ]: Get the argument
	STATIC  // [ ____ ___ PTR  ]: Copy a value from a static address
	CTX     // [ ____ ___ PTR  ]: Copy a value from a context address
	FRAME   // [ ____ ___ SRC  ]: Copy a value from a stack frame address
	ENUM    // [ OBJ  IDX ____ ]: Create a enum value
	SWITCH  // [ OBJ  EXT ____ ]: Choose a branch by a enum value
	SELECT  // [ OBJ* EXT ____ ]: Choose a branch by several enum values
	BR      // [ OBJ  IDX ____ ]: Make a functional branch ref on a enum value
	BRC     // [ OBJ  IDX ____ ]: Make a functional branch ref on a case ref
	BRP     // [ OBJ  IDX ____ ]: Make a functional branch ref on a proj ref
	TUPLE   // [ OBJ* ___ ____ ]: Create a tuple value
	GET     // [ OBJ  IDX ____ ]: Get the value of a field on a tuple value
	SET     // [ OBJ  IDX SRC  ]: Perform a functional update on a tuple value
	FR      // [ OBJ  IDX ____ ]: Make a functional field ref on a tuple value
	FRP     // [ OBJ  IDX ____ ]: Make a functional field ref on a proj ref
	LSV     // [ OBJ* ___ ____ ]: Create a variant list
	LSC     // [ OBJ* IDX ____ ]: Create a compact list according to type info
	MPS     // [ OBJ* ___ SRC* ]: Create a map with string key
	MPI     // [ OBJ* ___ SRC* ]: Create a map with integer key
	CL      // [ OBJ  ___ SRC* ]: Create a closure
	CLR     // [ OBJ  ___ SRC* ]: Create a self-referential closure
	CALL    // [ OBJ  ___ SRC  ]: Call a function
)

func (inst *Instruction) ToSize() LocalSize {
	if inst.OpCode != SIZE { panic("invalid operation") }
	return LocalSize(*(*uint64)(unsafe.Pointer(inst)))
}

func (inst Instruction) String() string {
	switch inst.OpCode {
	case SIZE:
		return fmt.Sprintf("SIZE == %d", inst.ToSize())
	case ARG:
		return "ARG"
	case STATIC:
		return fmt.Sprintf("STATIC (%d)", inst.Src)
	case CTX:
		return fmt.Sprintf("CTX (%d)", inst.Src)
	case FRAME:
		return fmt.Sprintf("FRAME %d", inst.Src)
	case ENUM:
		return fmt.Sprintf("ENUM %d [%d]", inst.Obj, inst.Idx)
	case SWITCH:
		return fmt.Sprintf("SWITCH %d [<%d>] %d*", inst.Obj, inst.ExtIdx, inst.Src)
	case SELECT:
		return fmt.Sprintf("SELECT %d* [<%d>] %d*", inst.Obj, inst.ExtIdx, inst.Src)
	case BR:
		return fmt.Sprintf("BR %d [%d]", inst.Obj, inst.Idx)
	case BRC:
		return fmt.Sprintf("BRC %d [%d]", inst.Obj, inst.Idx)
	case BRP:
		return fmt.Sprintf("BRP %d [%d]", inst.Obj, inst.Idx)
	case TUPLE:
		return fmt.Sprintf("Tuple %d*", inst.Obj)
	case GET:
		return fmt.Sprintf("GET %d [%d]", inst.Obj, inst.Idx)
	case SET:
		return fmt.Sprintf("SET %d [%d] %d", inst.Obj, inst.Idx, inst.Src)
	case FR:
		return fmt.Sprintf("FR %d [%d]", inst.Obj, inst.Idx)
	case FRP:
		return fmt.Sprintf("FRP %d [%d]", inst.Obj, inst.Idx)
	case LSV:
		return fmt.Sprintf("LSV %d*", inst.Obj)
	case LSC:
		return fmt.Sprintf("LSC %d* [%d]", inst.Obj, inst.Idx)
	case MPS:
		return fmt.Sprintf("MPS %d* %d*", inst.Obj, inst.Src)
	case MPI:
		return fmt.Sprintf("MPI %d* %d*", inst.Obj, inst.Src)
	case CL:
		return fmt.Sprintf("CL %d %d*", inst.Obj, inst.Src)
	case CLR:
		return fmt.Sprintf("CLR %d %d*", inst.Obj, inst.Src)
	case CALL:
		return fmt.Sprintf("CALL %d %d", inst.Obj, inst.Src)
	default:
		panic("impossible branch")
	}
}
