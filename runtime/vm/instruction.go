package vm


type Short = uint8
type Long = uint16

type Instruction struct {
	OpCode  OpType
	Arg0    Short
	Arg1    Long
}

type OpType Short
const (
	NOP  OpType  =  iota
	FUNC    // [Base, FunID]: Create a function value (closure)
	CTX     // [_, _]: Add the current value to the context of created closure
	PUSH    // [Base, ValId]: Push a top-level value (constant or function)
	POP     // [_, _]: Discard the current value
	LOAD    // [Offset, _]: Load a value from the reserved area of current frame
	STORE   // [Offset, _]: Store current value to the reserved area
	CALL    // [_, _]: Call a function
	NATIVE  // [Base, NativeId]: Call a native function
	STRUCT  // [Size, _]: Create a struct (value of a tuple/bundle type)
	FILL    // [_, _]: Fill the current value into created struct
	GET     // [Index, _]: Extract a struct field as new current value
	SET     // [Index, _]: Immutably update a struct field with current value
	MATCH   // [_, _]: Match on the current value
	JIF     // [Index, Addr]: Jump to Addr if Index matches the value
	JMP     // [_, Addr]: Jump to Addr unconditionally
)
