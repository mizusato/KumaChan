type Instruction uint32


type Operation uint32
const (
    Load Operation = iota
    Store
    Args
    Call
    Ret
    Invoke
)


type SourceType uint32
const (
    IntAddr SourceType = iota
    NumAddr
    StrAddr
    BinAddr
    IdAddr
    BoolVal
)


type DestType uint32
const (
    ArgNext DestType = iota
    FunPtr
    NewVar
    OldVar
)


type Address uint32
