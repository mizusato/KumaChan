var Operators = {
    '+': fun (
        'local operator_plus (Number x, Number y) -> Number',
        (x, y) => x + y
    )
}


function call_operator (name) {
    assert(has(name, Operators))
    return (x,y) => call(Operators[name], [x,y])
}


var helpers = scope => ({
    c: call,
    m: (obj, name, args) => call_method(scope, obj, name, args),
    o: call_operator,
    is: is,
    id: var_lookup(scope),
    dl: var_declare(scope),
    rt: var_assign(scope),
    v: Void
})
