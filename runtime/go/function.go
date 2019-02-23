const ARG_MAX = 10


type PassPolicy int
const (
    Immutable PassPolicy = iota
    Natural
    Dirty
)


/* underlying data of functions */


type ConsExpr struct {
    category string
    abstract string
}


type ParaExpr struct {
    name string
    constraint ConsExpr
    pass_policy PassPolicy
}


type ProtoExpr struct {
    affect EffectRange
    value ConsExpr
    parameters [ARG_MAX]ParaExpr
    quantity int
}


type FunData struct {
    proto ProtoExpr
    body []Instruction
    name string
    info string
    cache map[string]int  // identifier lookup cache: id -> depth
}


/* functional objects */


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


type FunctionObject struct {
    mod int
    fun int
    context *Scope
    prototype Prototype
    native_body func(*Scope) Object
}
func (f *FunctionObject) get_type() Type { return Function }
func (f *FunctionObject) create_scope_by_default(args *Arguments) (
    *Scope, *FunctionObject, error,
) {
    return f.create_scope(args, nil)
}


type OverloadObject struct {    
    functions []*FunctionObject
}
func (o *OverloadObject) get_type() Type { return Overload }
func (o *OverloadObject) create_scope_by_default(args *Arguments) (
    *Scope, *FunctionObject, error,
) {
    return o.create_scope(args, nil)
}


type FunctionEntity interface {
    // create_scope: (args, context) -> (scope, function, error)
    //   if context == nil, scope.context = function.context
    create_scope(*Arguments, *Scope) (*Scope, *FunctionObject, error)
}


type ArgBind struct {
    f FunctionEntity
    first_argument Object
}
func (a *ArgBind) get_type() Type { return Binding }
func (a *ArgBind) create_scope_by_default(args *Arguments) (
    *Scope, *FunctionObject, error,
) {
    if !args.is_full() {
        args.prepend(a.first_argument)  // dirty but efficient
                                        // keep the side-effect here in mind
        return a.f.create_scope(args, nil)
    } else  {
        return nil, nil, call_error(
            "too many arguments", "cannot prepend a fixed argument",
        )
    }
}


type CtxBind struct {
    f FunctionEntity
    context *Scope
}
func (c *CtxBind) get_type() Type { return Binding }
func (c *CtxBind) create_scope_by_default(args *Arguments) (
    *Scope, *FunctionObject, error,
) {
    return c.f.create_scope(args, c.context)
}


type CallableObject interface {
    create_scope_by_default(*Arguments) (*Scope, *FunctionObject, error)
}


func (f *FunctionObject) try_to_create_scope(args *Arguments, context *Scope) (
    *Scope,
) {
    if args.quantity == f.prototype.quantity {
        var parameters = f.prototype.parameters
        for i := 0; i < args.quantity; i++ {
            var parameter = parameters[i]
            var argument = args.values[i]
            var constraint = parameter.constraint
            if !constraint.check(argument) {
                return nil
            }
            if parameter.pass_policy == Dirty && is_solid(argument) {
                return nil
            }
        }
        var scope *Scope
        if context == nil {
            scope = CreateScope(f.context, f.prototype.affect)
        } else {
            scope = CreateScope(context, f.prototype.affect)
        }
        for i := 0; i < args.quantity; i++ {
            var parameter = parameters[i]
            var argument = args.values[i]
            if parameter.pass_policy == Immutable {
                scope.declare(parameter.name, ImRef(argument))
            } else {
                scope.declare(parameter.name, argument)
            }
        }
        return scope
    } else {
        return nil
    }
}


func (f *FunctionObject) create_scope(args *Arguments, context *Scope) (
    *Scope, *FunctionObject, error,
) {
    var scope *Scope
    if context == nil {
        scope = CreateScope(f.context, f.prototype.affect)
    } else {
        scope = CreateScope(context, f.prototype.affect)
    }
    if args.quantity == f.prototype.quantity {
        for i := 0; i < args.quantity; i++ {
            var arg = args.values[i]
            var parameter = f.prototype.parameters[i]
            var name = parameter.name
            var constraint = parameter.constraint
            if !constraint.check(arg) {
                return nil, nil, call_error("invalid argument", name)
            }
            if parameter.pass_policy == Dirty && is_solid(arg) {
                return nil, nil, call_error("solid argument", name)
            }
            if parameter.pass_policy == Immutable {
                scope.declare(name, ImRef(arg))
            } else {
                scope.declare(name, arg)
            }
        }
        return scope, f, nil
    } else {
        return nil, nil, call_error(
            "argument quantity error",
            printf(
                "%v arguments required but %v given",
                f.prototype.quantity,
                args.quantity,
            ),
        )
    }
}


func (o *OverloadObject) create_scope(args *Arguments, context *Scope) (
    *Scope, *FunctionObject, error,
) {
    var length = len(o.functions)
    for i := length-1; i >= 0; i-- {
        var scope = o.functions[i].try_to_create_scope(args, context)
        if scope != nil {
            return scope, o.functions[i], nil
        }
    }
    return nil, nil, call_error(
        "invalid call", "no matching function prototype",
    )
}


/** 
 *  TODO:
 *    - added(*FunctionObject)
 *    - concated(*OverloadObject)
 *    - find_method_for(*Object)
 */


type CallError struct {
    category string
    info string
}


func call_error(category string, info string) *CallError {
    return &CallError {
        category: category,
        info: info,
    }
}


func (e *CallError) Error() string {
    return e.category + ": " + e.info
}
