const ARG_MAX = 10


type PassPolicy int
const (
    Immutable PassPolicy = iota
    Natural
    Dirty
)


/* underlying data of functions */


type ParaExpr struct {
    name string
    constraint string
    pass_policy PassPolicy
}


type ProtoExpr struct {
    affect EffectRange
    value string
    parameters [ARG_MAX]ParaExpr
    quantity int
}


type FunData struct {
    proto ProtoExpr
    body []Instruction
    name string
    info string
}


/* callable objects (except class) */


type Parameter struct {
    name string
    constraint AbstractObject
    pass_policy PassPolicy
}
type Prototype struct {
    affect EffectRange
    value AbstractObject
    parameters [ARG_MAX]Parameter
    quantity int
}
type Arguments struct {
    values [ARG_MAX]Object
    quantity int
}


func (args *Arguments) empty() {
    for i := 0; i < args.quantity; i++ {
        args.values[i] = nil
    }
    args.quantity = 0
}


func (args *Arguments) is_full() bool {
    return args.quantity == ARG_MAX
}


func (args *Arguments) prepend(first_arg Object) {
    if !args.is_full() {
        for i := args.quantity; i > 0; i-- {
            args.values[i] = args.values[i-1]
        }
        args.values[0] = first_arg
        args.quantity += 1
    } else {
        panic("unable to prepend argument: quantity limit exceeded")
    }
}


func (args *Arguments) append(new_arg Object) {
    if !args.is_full() {
        args.values[args.quantity] = new_arg
        args.quantity += 1
    } else {
        panic("unable to append argument: quantity limit exceeded")
    }
}


// FunctionEntity = *FunctionObject | *OverloadObject
type FunctionEntity interface {
    __is_entity() {}
}


// CallableObject = FunctionEntity | *ArgBind | *CtxBind | *ClassObject
type CallableObject interface {
    __is_callable() {}
}


type FunctionObject struct {
    mod int
    fun int
    context *Scope
    prototype Prototype
}
func (f *FunctionObject) get_type() Type { return Function }
func (f *FunctionObject) __is_entity() {}
func (f *FunctionObject) __is_callable() {}


type OverloadObject struct {    
    functions []*FunctionObject
}
func (o *OverloadObject) get_type() Type { return Overload }
func (o *OverloadObject) __is_entity() {}
func (o *OverloadObject) __is_callable() {}


type ArgBind struct {
    f FunctionEntity
    first_argument Object
}
func (a *ArgBind) get_type() Type { return Binding }
func (a *ArgBind) __is_callable() {}


type CtxBind struct {
    f FunctionEntity
    context *Scope
}
func (c *CtxBind) get_type() Type { return Binding }
func (c *CtxBind) __is_callable() {}


func (f *FunctionObject) create_scope(args *Arguments, context *Scope) (*Scope) {
    var scope *Scope
    if context == nil {
        scope = CreateScope(f.context, f.prototype.affect)
    } else {
        scope = CreateScope(context, f.prototype.affect)
    }
    for i := 0; i < args.quantity; i++ {
        var arg = args.values[i]
        var parameter = f.prototype.parameters[i]
        var name = parameter.name
        if parameter.pass_policy == Immutable {
            scope.declare(name, ImRef(arg))
        } else {
            scope.declare(name, arg)
        }
    }
    return scope, f
}


func (o *OverloadObject) added(f *FunctionObject) *OverloadObject {
    return &OverloadObject {
        functions: append(o.functions, f),
    }
}


func (o1 *OverloadObject) concated(o2 *OverloadObject) *OverloadObject {
    return &OverloadObject {
        functions: append(o1.functions, o2.functions...),
    }
}
