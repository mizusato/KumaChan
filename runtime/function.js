/**
 *  Function & Scope
 */
const WrapperInfo = Symbol('WrapperInfo')

let Wrapped = $(x => (
    (typeof x == 'function')
    && typeof x[WrapperInfo] == 'object'
))

let FunctionTypes = {
    ES_Function: ES.Function,
    Wrapped: Wrapped,
    Function: Ins(Wrapped, $(f => has('context', f[WrapperInfo]))),
    Overload: Ins(Wrapped, $(f => has('functions', f[WrapperInfo]))),
    Binding: Ins(Wrapped, $(f => has('original', f[WrapperInfo])))
}

pour(Types, FunctionTypes)


/**
 *  Parameter & Function Prototype
 */

let Parameter = struct({
    name: Types.String,
    type: Type
}, null, $( p => assert(Object.isFrozen(p)) ))

let Prototype = struct({
    value_type: Type,
    parameters: Types.TypedList.of(Parameter)
}, null, $(
    proto => (
        assert(Object.isFrozen(proto))
        && no_repeat(map(proto.parameters, p => p.name))
    )
))

function parse_decl (string) {
    let match = string.match(/function +([^( ]+) +\(([^)]*)\) *-> *(.+)/)
    let [_, name, params_str, value_str] = match
    let has_p = params_str.trim().length > 0
    let parameters = has_p? (list(map(params_str.split(','), para_str => {
        para_str = para_str.trim()
        let match = para_str.match(/([^ :]+) *: *(.+)/)
        let [_, name, type_str] = match
        assert(has(type_str, Types))
        let type = Types[type_str]
        let parameter = { name, type }
        Object.freeze(parameter)
        return parameter
    }))): []
    Object.freeze(parameters)
    assert(has(value_str, Types))
    let value_type = Types[value_str]
    let proto = { parameters, value_type }
    Object.freeze(proto)
    assert(is(proto, Prototype))
    return { name, proto }
}


/**
 *  Scope Object
 */

class Scope {
    constructor (context, data = {}) {
        assert(context === null || context instanceof Scope)
        assert(is(data, Types.Hash))
        this.context = context
        this.data = data
        this.cache = {}
        this.types_of_non_fixed = {}
        Object.freeze(this)
    }
    is_fixed (variable) {
        assert(is(variable, Types.String))
        return !has(variable, this.types_of_non_fixed)
    }
    has (variable) {
        assert(is(variable, Types.String))
        return has(variable, this.data)
    }
    declare (variable, initial_value, is_fixed = true, type = Any) {
        assert(is(variable, Types.String))
        assert(is(is_fixed, Types.Bool))
        assert(is(type, Type))
        ensure(!this.has(name), 'variable_declared', name)
        ensure(is(initial_value, type), 'variable_invalid', name)
        this.data[variable] = initial_value
        if (!is_fixed) {
            this.types_of_non_fixed[variable] = type
        }
    }
    reset (variable, new_value) {
        assert(is(variable, Types.String))
        ensure(this.has(variable), 'variable_not_declared', variable)
        ensure(!this.is_fixed(variable), 'variable_fixed', variable)
        let type_ok = is(new_value, this.types_of_non_fixed[variable])
        ensure(type_ok, 'variable_invalid', variable)
        this.data[variable] = new_value
    }
    lookup (variable) {
        assert(is(variable, Types.String))
        let value = this.find(variable)
        ensure(value != NotFound, 'variable_not_found', variable)
        return value
    }
    try_to_declare (variable, initial_value, is_fixed = true, type = null) {
        assert(is(variable, Types.String))
        if (!this.has(variable)) {
            this.declare(variable, initial_value, is_fixed, type)
        }
    }
    find (variable) {
        assert(is(variable, Types.String))
        let scope_chain = iterate (
            this,
            scope => scope.context,
            scope => scope == null
        )
        return find(map(scope_chain, (scope, depth) => {
            if (has(variable, scope.cache)) {
                let cached_scope = scope.cache[variable]
                assert(cached_scope.has(variable))
                if (depth > 2) {
                    this.cache[variable] = cached_scope
                }
                return cached_scope.data[variable]
            }
            if (scope.has(variable)) {
                if (depth > 2) {
                    this.cache[variable] = scope
                }
                return scope.data[variable]
            } else {
                return NotFound
            }
        }), object => object != NotFound)
    }
}


/**
 *  Function Wrapper
 */

 let arg_msg = new Map([
     [1, 'arg_wrong_quantity'],
     [2, 'arg_invalid']
 ])

 function wrap (context, proto, vals, desc, raw) {
     assert(context instanceof Scope)
     assert(is(proto, Prototype))
     assert(vals === null || vals instanceof Scope)
     assert(is(raw, Types.ES_Function))
     assert(is(desc, Types.String))
     let invoke = (args, use_ctx = null, check = true) => {
         // check arguments
         if (check) {
             let result = check_args(args, proto)
             ensure(result.ok, arg_msg[result.err], result.info, args.length)
         }
         // generate scope
         let scope = new Scope(use_ctx || context)
         inject_args(args, proto, scope)
         if (vals != null) {
             foreach(vals.data, (k, v) => scope.declare(k, v))
         }
         // call the raw function
         call_stack.push(desc)
         let value = raw(scope)
         ensure(is(value, proto.value_type), 'retval_invalid')
         call_stack.pop()
         // return the value after type check
         return value
     }
     let arity = proto.parameters.length
     let wrapped = give_arity((...args) => invoke(args), arity)
     let info = { context, invoke, proto, vals, desc, raw }
     Object.freeze(info)
     wrapped[WrapperInfo] = info
     Object.freeze(wrapped)
     return wrapped
 }

function check_args (args, proto) {
    let arity = proto.parameters.length
    // check if argument quantity correct
    if (arity != args.length) {
        return { ok: false, err: 1, info: arity }
    }
    // check types
    for (let i=0; i<arity; i++) {
        let parameter = proto.parameters[i]
        let arg = args[i]
        let name = parameter.name
        // check if the argument is of required type
        if( !is(arg, parameter.type) ) {
            return { ok: false, err: 2, info: name }
        }
    }
    return { ok: true }
}

function inject_args (args, proto, scope) {
    let arity = proto.parameters.length
    for (let i=0; i<arity; i++) {
        let parameter = proto.parameters[i]
        let arg = args[i]
        // apply default values of schema
        if (is(parameter.type, Schema)) {
            arg = parameter.type.patch(arg)
        }
        // inject argument to scope
        scope.declare(parameter.name, arg)
    }
}

function bind_context (f, context) {
    assert(is(f, Wrapped))
    assert(context instanceof Scope)
    f = cancel_binding(f)
    let f_invoke = f[WrapperInfo].invoke
    let invoke = function (args, use_ctx = null) {
        assert(use_ctx === null)
        return f_invoke(args, context)
    }
    let binding = ((...args) => invoke(args))
    let info = { original: f, invoke: invoke }
    Object.freeze(info)
    binding[WrapperInfo] = info
    Object.freeze(binding)
    return binding
}

function cancel_binding (f) {
    assert(is(f, Wrapped))
    return f[WrapperInfo].original || f
}

function call (f, args, file = null, row = -1, col = -1) {
    if (is(f, Types.Class)) {
        f = f.create
    } else if (is(f, Types.TypeTemplate)) {
        f = f.inflate
    }
    let call_type = file? 1: 3
    if (is(f, Wrapped)) {
        let info = f[WrapperInfo]
        push_call(call_type, info.desc, file, row, col)
        let value = info.invoke(args)
        pop_call()
        return value
    } else if (is(f, Types.ES_Function)) {
        push_call(call_type, get_summary(f.toString()))
        let value = f.apply(args)
        pop_call()
        return value
    } else {
        push_call(call_type, '*** Non-Callable Object', file)
        ensure(false, 'non_callable')
    }
}

function fun (decl_string, body) {
    let parsed = parse_decl(decl_string)
    assert(is(body, Types.ES_Function))
    assert(!is(body, Types.Wrapped))
    return wrap(Global, parsed.proto, null, decl_string, (scope, expose) => {
        return body.apply(
            null,
            list(cat(
                map(parsed.proto.parameters, p => scope.lookup(p.name)),
                [scope, expose]
            ))
        )
    })
}


/**
 *  Overload Tools
 */

function overload (functions, name) {
    assert(is(functions, Types.TypedList.of(Types.Function)))
    assert(is(name, Types.String))
    functions = copy(functions)
    Object.freeze(functions)
    let desc = 'overload: ' + name
    let only1 = (functions.length == 1)
    let invoke = null
    if (only1) {
        invoke = (args, use_ctx = null) => {
            let info = functions[0][WrapperInfo]
            push_call(2, info.desc)
            let value = info.invoke(args, use_ctx)
            pop_call()
            return value
        }
    } else {
        invoke = (args, use_ctx = null) => {
            let info_list = []
            let result_list = []
            let ok = false
            let info = null
            let i = 0
            for (let f of rev(functions)) {
                info_list.push(f[WrapperInfo])
                result_list.push(check_args(args, info.proto))
                if (result_list[i].ok) {
                    info = info_list[i]
                    ok = true
                    break
                }
                i += 1
            }
            if (ok) {
                push_call(2, info.desc)
                let value = info.invoke(args, use_ctx)
                pop_call()
            } else {
                let n = i
                let available = map(count(n), i => {
                    let r = result_list[i]
                    return (
                        info_list[i].desc
                        + LF + INDENT + get_msg (
                            arg_msg[r.err], r.info. args.length
                        )
                    )
                })
                ensure(false, 'no_matching_function', available)
            }
        }
    }
    let o = ((...args) => invoke(args, null))
    let info = Object.freeze({ functions, invoke, desc })
    Object.freeze(info)
    o[WrapperInfo] = info
    Object.freeze(o)
    return o
}

function overload_added (f, o, name) {
    assert(is(o, Types.Overload))
    return overload([...o[WrapperInfo].functions, f], name)
}

function overload_concated (o2, o1, name) {
    assert(is(o2, Types.Overload))
    assert(is(o1, Types.Overload))
    return overload(list(
        cat(o1[WrapperInfo].functions, o2[WrapperInfo].functions)
    ), name)
}

function f (name, ...args) {
    assert(args.length % 2 == 0)
    let functions = list(map(
        count(args.length/2),
        i => fun(args[i*2], args[i*2+1])
    ))
    return overload(functions, name)
}
