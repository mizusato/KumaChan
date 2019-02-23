const ARG_STACK_SIZE = 100000
const CALL_STACK_SIZE = 100000


type Machine struct {
    modules []Module
    call_stack CallStack
    arg_stack ArgStack
    global_scope *Scope
    tmp Object
}


type Module struct {
    functions []FunData
    constants ConstTable
}


type ConstTable struct {
    int_values []IntObject
    num_values []NumberObject
    str_values []StringObject
}


type CallFrame struct {
    mod int
    fun int
    ptr int
    scope *Scope
    override Object
    value_constraint AbstractObject
}


type CallStack struct {
    frames [CALL_STACK_SIZE]CallFrame
    current int
}


type ArgFrame struct {
    f CallableObject
    args Arguments
}


type ArgStack struct {
    frames [ARG_STACK_SIZE]ArgFrame
    current int
}


func (stack *CallStack) init() {
    stack.current = -1
}


func (stack *CallStack) push() {
    stack.current++
}


func (stack *CallStack) pop() {
    stack.frames[stack.current] = CallFrame{}
    stack.current--
}


func (stack *CallStack) top() *CallFrame {
    return &(stack.frames[stack.current])
}


func (stack *ArgStack) init() {
    stack.current = -1
}


func (stack *ArgStack) push() {
    stack.current++
}


func (stack *ArgStack) pop() {
    var top = &(stack.frames[stack.current])
    top.f = nil
    top.args.empty()
    stack.current--
}


func (stack *ArgStack) top() *ArgFrame {
    return &(stack.frames[stack.current])
}


func CreateMachine(modules []Module) *Machine {
    if len(modules) == 0 { panic("no module available") }
    if len(modules[0].functions) == 0 { panic("entry module has no function") }
    // var entry *FunData = &modules[0].functions[0]
    var machine = &Machine {
        modules: modules,
        global_scope: generate_global_scope(),
    }
    machine.call_stack.init()
    machine.arg_stack.init()
    machine.call_stack.push()
    frame := machine.call_stack.top()
    frame.mod = 0
    frame.fun = 0
    frame.ptr = 0
    frame.scope = CreateScope(machine.global_scope, Local)
    return machine
}


func run(machine *Machine) {
    var current_frame *CallFrame = machine.call_stack.top()
    var mod = current_frame.mod
    var fun = current_frame.fun
    var ptr = current_frame.ptr
    var scope = current_frame.scope
    var module *Module = &(machine.modules[current_frame.mod])
    var function *FunData = &(module.functions[current_frame.fun])
    var instructions []Instruction = function.body
    var length = len(function.body)
    
    for ptr < length {
        inst := instructions[ptr]
        op, t, addr := inst.parse()
        switch op {
        case Load:
            switch t {
            case IntConst:
                machine.tmp = module.constants.int_values[addr]
            case NumConst:
                machine.tmp = module.constants.num_values[addr]
            case StrConst:
                machine.tmp = module.constants.str_values[addr]
            /*
            case BinConst:
                TODO()
            */
            case BoolVal:
                if addr == 0 {
                    machine.tmp = BoolObject(false)
                } else {
                    machine.tmp = BoolObject(true)
                }
            case VarLookup:
                identifier := module.constants.str_values[addr]
                value := scope.lookup(identifier)
                if value != nil {
                    machine.tmp = value
                } else {
                    panic(printf("variable %v not found", identifier))
                }
            default:
                panic(printf("invalid instruction %x", inst))
            }
        case Store:
        case Args:
            module.arg_stack.push()
        case Call:
        case Invoke:
            //module.arg_stack.top().args
        case Ret:
        }
    }
    
}
