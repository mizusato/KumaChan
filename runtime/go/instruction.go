type Instruction uint32


/**
 *  Instruction: 32 bits
 *  
 *  ------------------------------------
 *  | 04 | 04 |           24           |
 *  ------------------------------------
 *    op  type          address
 */


func (inst Instruction) parse() (Operation, AddrType, Address) {
    return Operation(inst >> 28),
           AddrType(inst << 4 >> 28),
           Address(inst << 8 >> 8)    
}


type Operation uint32
const (
    Load Operation = iota
    Store
    Args
    Call
    Invoke
    Ret
)


type AddrType uint32
const (
    /* load */
    IntConst AddrType = iota
    NumConst
    StrConst
    BinConst
    BoolVal
    VarLookup
    /* store */
    ArgNext
    FunId
    VarDeclare
    VarAssign
)


type Address uint32
