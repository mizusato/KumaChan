/**
 *  Function & Scope
 */


/**
 *  Effect Range of Function
 *
 *  The effect range of a function determines the range of scope chain
 *  that can be affected by the function, which indicates
 *  the magnitude of side-effect.
 *
 *  |  value  | Local Scope | Upper Scope | Other Scope |
 *  |---------|-------------|-------------|-------------|
 *  |  global | full-access | full-access | full-access |
 *  |  upper  | full-access | full-access |  read-only  |
 *  |  local  | full-access |  read-only  |  read-only  |
 *
 */

let EffectRange = one_of('local', 'upper', 'global')


/**
 *  Pass Policy of Parameter
 *
 *  The pass policy of a parameter determines how the function
 *    process the corresponding argument.
 *  If pass policy is set to immutable, the function will not be able to
 *    modify the argument. (e.g. add element to list)
 *
 *  |   value   | Immutable Argument |  Mutable Argument  |
 *  |-----------|--------------------|--------------------|
 *  | immutable |    direct pass     | treat as immutable |
 *  |  natural  |    direct pass     |    direct pass     |
 *  |   dirty   |     forbidden      |    direct pass     |
 *
 */

let PassPolicy = one_of('immutable', 'natural', 'dirty')


let Wrapped = $(x => (
    (typeof x == 'function')
    && typeof x[WrapperInfo] == 'object'
))

let FunctionTypes = {
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
    type: Type,
    pass_policy: PassPolicy,
}, null, $( p => assert(Object.isFrozen(p)) ))

let Prototype = struct({
    affect: EffectRange,
    value: Type,
    parameters: Types.TypedList.of(Parameter)
}, null, $(
    proto => (
        assert(Object.isFrozen(proto))
        && no_repeat(map(proto.parameters, p => p.name))
    )
))

let PassFlag = { natural: '*', dirty: '&', immutable: '' }
let FlagValue = { '*': 'natural', '&': 'dirty', '': 'immutable' }

function parse_decl (string) {
    let match = string.match(/([^ ]+) ([^( ]+) *\(([^)]*)\) -> (.+)/)
    let [_, affect, name, params_str, value_str] = match
    let has_p = params_str.trim().length > 0
    let parameters = has_p? (list(map(params_str.split(','), para_str => {
        para_str = para_str.trim()
        let match = para_str.match(/([^ ]+) (\*|&)?(.+)/)
        let [_, type_str, policy_str, name] = match
        assert(has(type_str, Types))
        let type = Types[type_str]
        policy_str = policy_str || ''
        let pass_policy = FlagValue[policy_str]
        let parameter = { name, type, pass_policy }
        Object.freeze(parameter)
        return parameter
    }))): []
    Object.freeze(parameters)
    assert(has(value_str, Types))
    let value = Types[value_str]
    let proto = { affect, parameters, value }
    Object.freeze(proto)
    assert(is(proto, Prototype))
    return { name, proto }
}


/**
 *  Scope Object
 */

class Scope {
    constructor (context, affect = 'local', data = {}) {
        assert(context === null || context instanceof Scope)
        assert(is(affect, EffectRange))
        assert(is(data, Types.Hash))
        // <context> = upper scope
        this.context = context
        // <affect> = effect range of the corresponding function
        this.affect = affect
        // <data> = Hash { VariableName -> VariableValue }
        this.data = data
        // <assignable> = Set { Non-Constants }
        this.assignable = new Set()
        Object.freeze(this)
    }
    check_assignable (variable) {
        assert(is(variable, Types.String))
        return this.assignable.has(variable)
    }
    has (variable) {
        assert(is(variable, Types.String))
        return has(variable, this.data)
    }
    declare (variable, initial_value, is_assignable = false) {
        assert(is(variable, Types.String))
        assert(!this.has(variable))
        this.data[variable] = initial_value
        if (is_assignable) {
            this.assignable.add(variable)
        }
    }
    try_to_declare (variable, initial_value, is_assignable = false) {
        assert(is(variable, Types.String))
        if (!this.has(variable)) {
            this.declare(variable, initial_value, is_assignable)
        }
    }
    assign (variable, new_value) {
        assert(is(variable, Types.String))
        assert(this.has(variable))
        assert(this.assignable.has(variable))
        this.data[variable] = new_value
    }
    force_declare (variable, initial_value) {
        assert(is(variable, Types.String))
        if (this.has(variable)) {
            this.assign(variable, initial_value)
        } else {
            this.declare(variable, initial_value, true)
        }
    }
    unset (variable) {
        assert(is(variable, Types.String))
        assert(this.has(variable))
        delete this.data[variable]
    }
    lookup (variable) {
        assert(is(variable, Types.String))
        let info = this.find(variable)
        if (info == NotFound) {
            return NotFound
        } else {
            return info.is_mutable? info.object: Im(info.object)
        }
    }
    find (variable) {
        assert(is(variable, Types.String))
        let affect = this.affect
        let mutable_depth = 0
        if (affect == 'local') {
            mutable_depth = 0
        } else if (affect == 'upper') {
            mutable_depth = 1
            let upper = this.context
            while (upper != null && upper.affect == 'upper') {
                mutable_depth += 1
                upper = upper.context
            }
        } else if (affect == 'global') {
            mutable_depth = Infinity
        }
        let scope_chain = iterate(
            this, scope => scope.context, scope => scope == null
        )
        return find(map(scope_chain, (scope, depth) => {
            let object = (
                scope.has(variable)? scope.data[variable]: NotFound
            )
            let is_mutable = (
                depth <= mutable_depth && IsMut(object)
            )
            let is_assignable = scope.check_assignable(variable)
            return { scope, depth, object, is_mutable, is_assignable }
        }), info => info.object != NotFound)
    }
}


/**
 *  Scope Operation Functions with Error Handling
 */

function var_lookup(scope) {
    assert(scope instanceof Scope)
    return function lookup (name) {
        let value = scope.lookup(name)
        ensure(value != NotFound, 'variable_not_found', name)
        return value
    }
}

function var_declare(scope) {
    assert(scope instanceof Scope)
    return function declare (name, initial_value, is_assignable = false) {
        ensure(!scope.has(name), 'variable_declared', name)
        scope.declare(name, initial_value, is_assignable)
        return Void
    }
}

function var_assign(scope) {
    assert(scope instanceof Scope)
    return function assign (name, new_value) {
        let info = scope.find(name)
        ensure(info != NotFound, 'variable_not_declared', name)
        ensure(info.is_assignable, 'variable_cannot_reset', name)
        ensure(info.is_mutable, 'variable_immutable', name)
        info.scope.assign(name, new_value)
        return Void
    }
}


/**
 *  Function Wrapper
 */

 let arg_msg = new Map([
     [1, 'arg_wrong_quantity'],
     [2, 'arg_invalid'],
     [3, 'arg_immutable']
 ])

 function wrap (context, proto, vals, desc, raw) {
     assert(context instanceof Scope)
     assert(is(proto, Prototype))
     assert(vals === null || vals instanceof Scope)
     assert(is(raw, ES.Function))
     assert(is(desc, Types.String))
     let invoke = (args, use_ctx = null, check = true) => {
         // check arguments
         if (check) {
             let result = check_args(args, proto)
             ensure(result.ok, arg_msg[result.err], result.info, args.length)
         }
         // generate scope
         let scope = new Scope(use_ctx || context, proto.affect)
         inject_args(args, proto, scope)
         if (vals != null) {
             foreach(vals.data, (k, v) => scope.declare(k, v))
         }
         // call the raw function
         call_stack.push(desc)
         let value = raw(scope)
         ensure(is(value, proto.value), 'retval_invalid')
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
        // cannot pass immutable object as dirty argument
        let is_dirty = parameter.pass_policy == 'dirty'
        let is_immutable = IsIm(arg)
        if (is_dirty && is_immutable) {
            return { ok: false, err: 3, info: name }
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
            arg = parameter.type.patch(NoRef(arg))
        }
        // if pass policy is immutable, register the argument
        if (parameter.pass_policy == 'immutable') {
            arg = Im(arg)
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
    let invoke = function (args, use_context = null) {
        assert(use_context === null)
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
    }
    let call_type = file? 1: 3
    if (is(f, Wrapped)) {
        let info = f[WrapperInfo]
        push_call(call_type, info.desc, file, row, col)
        let value = info.invoke(args)
        pop_call()
        return value
    } else if (is(f, ES.Function)) {
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
    assert(is(body, Type.Function.Bare))
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
                let available = map(iterate(0, i => i+1, i => !(i < n)), i => {
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
        iterate(0, i => i+2, i => !(i < args.length)),
        i => fun(args[i], args[i+1])
    ))
    return overload(functions, name)
}
