const ARG_MAX = 10


type Constraint struct {
    category string
    abstract string
}


type ParaExpr struct {
    name string
    constraint Constraint
}


type ProtoExpr struct {
    affect EffectRange
    value Constraint
    parameters [ARG_MAX]ParaExpr
    quantity int
}


type FunData struct {
    proto ProtoExpr
    body []Instruction
    name string
    info string
}


type Parameter struct {
    name string
    constraint AbstractObject
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


type FunctionObject struct {
    module_index int
    data_index int
    context *Scope
    prototype Prototype
}


func (f *FunctionObject) get_type() Type { return Function }


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


func (f *FunctionObject) try_to_create_scope(args *Arguments) *Scope {
    if args.quantity == f.prototype.quantity {
        var parameters = f.prototype.parameters
        for i := 0; i < args.quantity; i++ {
            var constraint = parameters[i].constraint
            if !constraint.checker(args.values[i]) {
                return nil
            }
        }
        var scope = CreateScope(f.context, f.prototype.affect)
        for i := 0; i < args.quantity; i++ {
            scope.declare(parameters[i].name, args.values[i])
        }
        return scope
    } else {
        return nil
    }
}


func (f *FunctionObject) create_scope(args *Arguments) (*Scope, error) {
    var scope = CreateScope(f.context, f.prototype.affect)
    if args.quantity == f.prototype.quantity {
        for i := 0; i < args.quantity; i++ {
            var arg = args.values[i]
            var parameter = f.prototype.parameters[i]
            var name = parameter.name
            var constraint = parameter.constraint
            if constraint.checker(arg) {
                scope.declare(name, arg)
            } else {
                return nil, call_error("invalid argument", name)
            }
        }
        return scope, nil
    } else {
        return nil, call_error(
            "argument quantity error",
            printf(
                "%v arguments required but %v given",
                f.prototype.quantity,
                args.quantity,
            ),
        )
    }
}


type OverloadObject struct {    
    functions []*FunctionObject
}


func (o *OverloadObject) get_type() Type { return Overload }


/** 
 *  TODO:
 *    - added(*FunctionObject)
 *    - concated(*OverloadObject)
 *    - find_method_for(*Object)
 */


func (o *OverloadObject) create_scope(args *Arguments) (*Scope, error) {
    var length = len(o.functions)
    for i := length-1; i >= 0; i-- {
        var scope = o.functions[i].try_to_create_scope(args)
        if scope != nil {
            return scope, nil
        }
    }
    return nil, call_error("invalid call", "no matching function prototype")
}


type FunctionalObject interface {
    create_scope(*Arguments) (*Scope, error)
    // TODO: (*Scope, *FunctionObject, error)
}
