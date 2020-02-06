package vm


type Instruction struct {
	OpCode  OpType
	Arg0    uint8
	Arg1    uint16
}

type OpType uint8
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
	SET     // [Index, _]: Stage update of a struct field
	COMMIT  // [_, _]: Commit the staged update of struct fields
	MATCH   // [_, _]: Match on the current value
	JIF     // [Index, Addr]: Jump to Addr if Index matches the value
	JMP     // [_, Addr]: Jump to Addr unconditionally
)
