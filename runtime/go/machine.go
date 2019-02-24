const ARG_STACK_SIZE = 100000
const CALL_STACK_SIZE = 100000


type RuntimeError struct {
    msg string
}


func (e *RuntimeError) Error() string {
    return e.msg
}


func create_error(msg string) *RuntimeError {
    return &RuntimeError { msg: msg }
}


type ErrorProducer struct {
    assert func(bool, string)
    when func(error)
    throw func(string)
}


func err_producer(machine *Machine) ErrorProducer {
    var print_msg = func (msg string) {
        // TODO: print call stack
        panic(printf("%v\n%v", msg, machine))
    }
    return ErrorProducer {
        assert: func(val bool, msg string) {
            if !val {
                print_msg(msg)
            }
        },
        when: func(err error) {
            if err != nil {
                print_msg(err.Error())
            }
        },
        throw: func(msg string) {
            print_msg(msg)
        },
    }
}


type Machine struct {
    modules []Module
    call_stack CallStack
    arg_stack ArgStack
    global_scope *Scope
    tmp Object  // vitural register
}


type Module struct {
    functions []FunData
    constants ConstTable
}


type ConstTable struct {
    int_values []IntegerObject
    num_values []NumberObject
    str_values []StringObject
}


func (module *Module) get_int(addr Address) IntegerObject {
    if addr < len(module.constants.int_values) {
        return module.constants.int_values[addr]
    } else {
        return IntegerObject(0)
    }
}


func (module *Module) get_num(addr Address) NumberObject {
    if addr < len(module.constants.num_values) {
        return module.constants.num_values[addr]
    } else {
        return NumberObject(0.0)
    }
}


func (module *Module) get_str(addr Address) StringObject {
    if addr < len(module.constants.str_values) {
        return module.constants.str_values[addr]
    } else {
        return StringObject("")
    }
}


func (module *Module) get_fun(addr Address) *FunData {
    if add < len(module.functions) {
        return &(module.functions[addr])
    } else {
        return nil
    }
}


type CallType int
const (
    Ordinary CallType = iota
    ArgChecker
    ValChecker
)


type CallFrame struct {
    mod int
    fun int
    ptr int
    scope *Scope
    override Object
    call_type CallType
    value Object
    value_constraint AbstractObject
}


func (c *CallFrame) get_fun_data(machine *Machine) *FunData {
    if 0 <= c.mod && c.mod < len(machine.modules) {
        functions := machine.modules[c.mod].functions
        if 0 <= c.fun && c.fun < len(functions) {
            return &(functions[c.fun])
        } else {
            return nil
        }
    } else {
        return nil
    }
}


type CallStack struct {
    frames [CALL_STACK_SIZE]CallFrame
    current int
}


type ArgFrame struct {
    f CallableObject
    args Arguments
    call_type CallType
    emplace_fisished bool
    checked_arg_quantity int
    remaining_fun_quantity int
}


type ArgStack struct {
    frames [ARG_STACK_SIZE]ArgFrame
    current int
}


func (stack *CallStack) push() {
    stack.current++
}


func (stack *CallStack) pop() {
    stack.frames[stack.current] = CallFrame{}
    if stack.current > 0 {
        stack.current--
    }
}


func (stack *CallStack) top() *CallFrame {
    return &(stack.frames[stack.current])
}


func (stack *ArgStack) push() {
    stack.current++
}


func (stack *ArgStack) pop() {
    stack.frames[stack.current] = ArgFrame{}
    if stack.current > 0 {
        stack.current--
    }
}


func (stack *ArgStack) top() *ArgFrame {
    return &(stack.frames[stack.current])
}


func (stack *ArgStack) get_args() *Arguments {
    return &(stack.top().args)
}


func (stack *ArgStack) set_callee(f CallableObject) {
    top := stack.top()
    top.f = callable
}


func CreateMachine(modules []Module) *Machine {
    if len(modules) == 0 { panic("no module available") }
    if len(modules[0].functions) == 0 { panic("entry module has no function") }
    // var entry *FunData = &modules[0].functions[0]
    var machine = &Machine {
        modules: modules,
        global_scope: generate_global_scope(),
    }
    frame := machine.call_stack.top()
    frame.mod = 0
    frame.fun = 0
    frame.ptr = 0
    frame.scope = CreateScope(machine.global_scope, Local)
    return machine
}


func run(machine *Machine) {
    var fatal = err_producer(machine)
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
        // TODO
        op, t, addr := inst.parse()
        switch op {
        case Load:
            switch t {
            case IntAddr:
                machine.tmp = module.get_int(addr)
            case NumAddr:
                machine.tmp = module.get_num(addr)
            case StrAddr:
                machine.tmp = module.get_str(addr)
            // case BinConst: TODO
            case BoolVal:
                machine.tmp = BoolObject(addr != 0)
            case FunAddr:
                fun_data := module.get_fun(addr)
                fatal.assert(fun_data != nil, "invalid function address")
                args := machine.arg_stack.get_args()
                required := fun_data.proto.quantity + 1
                given := args.quantity
                fatal.assert(required == given, printf(
                    "%v constraints required but %v given",
                    required, given,
                ))
                prototype := Prototype{}
                prototype.affect = fun_data.proto.affect
                prototype.quantity = fun_data.proto.quantity
                for i := 0; i < prototype.quantity; i++ {
                    para_expr := fun_data.proto.parameters[i]
                    prototype.parameters[i].name = para_expr.name
                    prototype.parameters[i].pass_policy = para_expr.pass_policy
                    arg_cons_expr := para_expr.constraint
                    arg := args.values[i]
                    constraint, ok := arg.(AbstractObject)
                    fatal.assert(ok, printf(
                        "invalid argument constraint %v", arg_cons_expr,
                    ))
                    prototype.parameters[i].constraint = constraint
                }
                retval_cons_expr := fun_data.proto.value
                arg := args.values[fun_data.proto.quantity]
                constraint, ok := arg.(AbstractObject)
                fatal.assert(ok, printf(
                    "invalid return value constraint %v", retval_cons_expr
                ))
                prototype.value = constraint
                module.tmp = &(FunctionObject {
                    mod: mod,
                    fun: addr,
                    context: scope,
                    prototype: prototype
                })
            case VarLookup:
                identifier := module.get_str(addr)
                value := scope.lookup(identifier)
                fatal.assert(value != nil, printf(
                    "variable %v not found", identifier,
                ))
                machine.tmp = value
            default:
                fatal.throw(printf("invalid instruction %x", inst))
            }
        case Store:
            switch t {
            case ArgNext:
                args := machine.arg_stack.get_args()
                fatal.assert(
                    !args.is_full(),
                    "argument quantity limit exceeded",
                )
                args.append(machine.tmp)
            case Callee:
                f, ok := machine.tmp.(CallableObject)
                fatal.assert(ok, "invalid callee object")
                machine.arg_stack.set_callee(f)
            case VarDeclare:
                identifier := module.get_str(addr)
                fatal.assert(!scope.has(identifier), printf(
                    "variable %v already declared", identifier
                ))
                scope.declare(identifier, machine.tmp)
            case VarAssign:
                identifier := module.get_str(addr)
                err := scope.assign(identifier, machine.tmp)
                fatal.assert(err != "", err)
            default:
                fatal.throw(printf("invalid instruction %x", inst))
            }
        case Push:
            machine.arg_stack.push()
        case Call:
            top := machine.arg_stack.top()
            top.emplace_finished = true
        case Invoke:
            args := machine.arg_stack.get_args()
            f := InternalFunction(addr)
            switch f {
            case f_list:
                machine.tmp = CreateList()
            case f_hash:
                machine.tmp = CreateHash()
            case f_element:
                fatal.assert(
                    args.quantity == 2,
                    "element(): wrong argument quantity"
                )
                list, ok := args.values[0].(*ListObject)
                element := args.values[1]
                fatal.assert(ok, "element(): invalid argument")
                list.append(element)
                machine.tmp = list
            case f_pair:
                fatal.assert(
                    args.quantity == 3,
                    "pair(): wrong argument quantity"
                )
                hash, ok0 := args.values[0].(*HashObject)
                key, ok1 := args.values[1].(StringObject)
                value := args.values[2]
                fatal.assert(ok0 && ok1, "pair(): invalid argument")
                hash.set(string(key), value)
                machine.tmp = hash
            default:
                fatal.throw(printf("invalid instruction %x", inst))
            }
            machine.args_stack.pop()
        case Ret:
            callee_frame := machine.call_stack.top()
            switch callee_frame.call_type {
            case Ordinary:
                cur := callee_frame
                cur.value = machine.tmp
                concept := cur.value_constraint.get_concept()
                switch c := concept.(type) {
                case ScriptConcept:
                    machine.args_stack.push()
                    top := machine.args_stack.top()
                    top.call_type = ValChecker
                    top.args.append(cur.value)
                    top.emplace_finished = true
                    top.checked_arg_quantity = 1
                    top.callee = c.checker
                case NativeConcept:
                    fatal.assert(c.checker(cur.value), "invalid return value")
                    machine.tmp = cur.value
                    machine.call_stack.pop()
                default:
                    fatal.throw("cannot extract checker from unknown concept")
                }
                continue
            case ValChecker:
                v, ok := machine.tmp.(BoolObject)
                fatal.assert(ok, "non-boolean value returned by checker")
                machine.call_stack.pop()
                fatal.assert(v, "invalid return value")
                cur := machine.call_stack.top()
                machine.tmp = cur.value
                machine.call_stack.pop()
                continue
            case ArgChecker:
                v, ok := machine.tmp.(BoolObject)
                fatal.assert(ok, "non-boolean value returned by checker")
                arg_frame := machine.arg_stack.top()
                machine.call_stack.pop()
                if bool(v) {
                    arg_frame.checked_arg_quantity++
                } else {
                    n := arg_frame.checked_arg_quantity
                    cur := machine.call_stack.top()
                    fun_data := cur.get_fun_data()
                    if fun_data == nil {
                        fatal.throw(printf("invalid argument #%v", n))
                    } else {
                        name := fun_data.proto.parameters[n].name
                        fatal.throw(printf("invalid argument %v", name))
                    }
                }
                continue
            }
        default:
            fatal.throw(printf("invalid instruction %x", inst))            
        }
        ptr++
    }
}
