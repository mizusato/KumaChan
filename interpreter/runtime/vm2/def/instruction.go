package def

import (
	"unsafe"
	"fmt"
)


type ShortIndex uint8
const ShortIndexSize = unsafe.Sizeof(ShortIndex(0))

type ExternalIndexMapPointer uint16
const ExternalIndexMapPointerSize = unsafe.Sizeof(ExternalIndexMapPointer(0))

type SelectArgIndexVector uint32
const SelectArgIndexVectorSize = unsafe.Sizeof(SelectArgIndexVector(0))

type LocalAddr uint16
const LocalAddrSize = unsafe.Sizeof(LocalAddr(0))

const MaxSelectArgLength = (SelectArgIndexVectorSize / ShortIndexSize)
const MaxEnumCases = (1 << ShortIndexSize)
const MaxTupleElements = (1 << ShortIndexSize)
const MaxClosureContexts = (1 << ShortIndexSize)
const MaxBranchExpressions = (1 << ExternalIndexMapPointerSize)
const MaxFrameValues = (1 << LocalAddrSize)
const MaxStaticValues = (1 << LocalAddrSize)

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

type ExternalIndexMapping struct {
	SwitchMapping  ([] [] uint)
	SelectMapping  ([] map[SelectArgIndexVector] uint)
}
func (e ExternalIndexMapping) SwitchChooseBranch (
	ptr  ExternalIndexMapPointer,
	idx  ShortIndex,
) uint {
	var m = e.SwitchMapping
	if !(uint(ptr) < uint(len(m))) { panic("invalid ExtIdx") }
	var branches = m[ptr]
	if !(uint(idx) < uint(len(branches))) { panic("invalid branch index") }
	return branches[idx]
}
func (e ExternalIndexMapping) SelectChooseBranch (
	ptr  ExternalIndexMapPointer,
	vec  SelectArgIndexVector,
) uint {
	var m = e.SelectMapping
	if !(uint(ptr) < uint(len(m))) { panic("invalid ExtIdx") }
	var branch, found = m[ptr][vec]
	if !(found) { panic("branch not found") }
	return branch
}

type OpCode uint8
const (
	SIZE    OpCode = iota  // whole instruction as an unsigned number
	STATIC  // [ ____ ___ SRC! ]: Copy the value from a static address
	FRAME   // [ ____ ___ SRC  ]: Copy the value from a stack frame address
	ENUM    // [ OBJ  IDX ____ ]: Create a enum value
	SWITCH  // [ OBJ  EXT SRC* ]: Choose a branch by a enum value
	SELECT  // [ OBJ* EXT SRC* ]: Choose a branch by a tuple of enum values
	BR      // [ OBJ  IDX ____ ]: Make a functional branch ref on a enum value
	BRB     // [ OBJ  IDX ____ ]: Make a functional branch ref on a branch ref
	BRF     // [ OBJ  IDX ____ ]: Make a functional branch ref on a field ref
	TUPLE   // [ OBJ* ___ ____ ]: Create a tuple value
	GET     // [ OBJ  IDX ____ ]: Get the value of a field on a tuple value
	SET     // [ OBJ  IDX SRC  ]: Perform a functional update on a tuple value
	FR      // [ OBJ  IDX ____ ]: Make a functional field ref on a tuple value
	FRF     // [ OBJ  IDX ____ ]: Make a functional field ref on a field ref
	LSV     // [ OBJ* ___ ____ ]: Create a variant list
	LSC     // [ OBJ* IDX ____ ]: Create a compact list according to type info
	MPS     // [ OBJ* ___ SRC* ]: Create a map with string key
	MPI     // [ OBJ* ___ SRC* ]: Create a map with integer key
	CL      // [ OBJ  ___ SRC* ]: Create a closure
	CLR     // [ OBJ  ___ SRC* ]: Create a self-referential closure
	CALL    // [ OBJ  ___ SRC  ]: Call a function
)

func (inst *Instruction) ToSize() uint {
	if inst.OpCode != SIZE { panic("invalid operation") }
	return uint(*(*uint64)(unsafe.Pointer(inst)))
}

func (inst Instruction) String() string {
	switch inst.OpCode {
	case SIZE:
		return fmt.Sprintf("SIZE == %d", inst.ToSize())
	case STATIC:
		return fmt.Sprintf("STATIC (%d)", inst.Src)
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
	case BRB:
		return fmt.Sprintf("BRB %d [%d]", inst.Obj, inst.Idx)
	case BRF:
		return fmt.Sprintf("BRF %d [%d]", inst.Obj, inst.Idx)
	case TUPLE:
		return fmt.Sprintf("Tuple %d*", inst.Obj)
	case GET:
		return fmt.Sprintf("GET %d [%d]", inst.Obj, inst.Idx)
	case SET:
		return fmt.Sprintf("SET %d [%d] %d", inst.Obj, inst.Idx, inst.Src)
	case FR:
		return fmt.Sprintf("FR %d [%d]", inst.Obj, inst.Idx)
	case FRF:
		return fmt.Sprintf("FRF %d [%d]", inst.Obj, inst.Idx)
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

