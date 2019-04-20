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


/**
 *  Parameter & Function Prototype
 */

let Parameter = struct({
    name: Type.String,
    pass_policy: PassPolicy,
    constraint: Type.Abstract
})

let ParameterList = list_of(Parameter)

let Prototype = Ins(struct({
    affect: EffectRange,
    value: Type.Abstract,
    parameters: ParameterList
}), $( proto => no_repeat(map(proto.parameters, p => p.name)) ))

let PassFlag = { natural: '*', dirty: '&', immutable: '' }
let FlagValue = { '*': 'natural', '&': 'dirty', '': 'immutable' }

function parse_decl (string) {
    let match = string.match(/([^ ]+) ([^\( ]+) *\(([^\)]*)\) -> (.+)/)
    let [_, affect, name, params_str, value_str] = match
    let has_p = params_str.trim().length > 0
    let parameters = has_p? (list(map(params_str.split(','), para_str => {
        para_str = para_str.trim()
        let match = para_str.match(/([^ ]+) (\*|\&)?(.+)/)
        let [_, type_str, policy_str, name] = match
        let constraint = Global.lookup(type_str)
        policy_str = policy_str || ''
        let pass_policy = FlagValue[policy_str]
        return { name, constraint, pass_policy }
    }))): []
    let value = Global.lookup(value_str)
    let proto = { affect, parameters, value }
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
        assert(is(data, Type.Container.Hash))
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
        assert(typeof variable == 'string')
        return this.assignable.has(variable)
    }
    has (variable) {
        assert(typeof variable == 'string')
        return has(variable, this.data)
    }
    declare (variable, initial_value, is_assignable = false) {
        assert(typeof variable == 'string')
        assert(!this.has(variable))
        this.data[variable] = initial_value
        if (is_assignable) {
            this.assignable.add(variable)
        }
    }
    try_to_declare (variable, initial_value, is_assignable = false) {
        assert(typeof variable == 'string')
        if (!this.has(variable)) {
            this.declare(variable, initial_value, is_assignable)
        }
    }
    assign (variable, new_value) {
        assert(typeof variable == 'string')
        assert(this.has(variable))
        assert(this.assignable.has(variable))
        this.data[variable] = new_value
    }
    force_declare (variable, initial_value) {
        assert(typeof variable == 'string')
        if (this.has(variable)) {
            this.assign(variable, initial_value)
        } else {
            this.declare(variable, initial_value, true)
        }
    }
    unset (variable) {
        assert(typeof variable == 'string')
        assert(this.has(variable))
        delete this.data[variable]
    }
    lookup (variable) {
        assert(typeof variable == 'string')
        let info = this.find(variable)
        if (info == NotFound) {
            return NotFound
        } else {
            return info.is_mutable? info.object: Im(info.object)
        }
    }
    find (variable) {
        assert(typeof variable == 'string')
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
 *  Scope Operation Functions with Error Producer
 */

let name_err = new ErrorProducer(NameError)
let assign_err = new ErrorProducer(AssignError)
let access_err = new ErrorProducer(AccessError)

function var_lookup(scope) {
    assert(scope instanceof Scope)
    return function lookup (name) {
        let value = scope.lookup(name)
        name_err.assert(value != NotFound, MSG.variable_not_found(name))
        return value
    }
}

function var_declare(scope) {
    assert(scope instanceof Scope)
    return function declare (name, initial_value, is_assignable = false) {
        name_err.assert(!scope.has(name), MSG.variable_declared(name))
        scope.declare(name, initial_value, is_assignable)
    }
}

function var_assign(scope) {
    assert(scope instanceof Scope)
    return function assign (name, new_value) {
        let info = scope.find(name)
        name_err.assert(info != NotFound, MSG.variable_not_declared(name))
        assign_err.assert(info.is_assignable, MSG.variable_const(name))
        access_err.assert(info.is_mutable, MSG.variable_immutable(name))
        info.scope.assign(name, new_value)
        return Void
    }
}


function scope_kit (scope) {
    return {
        id: var_lookup(scope),
        reset: var_assign(scope),
        declare: var_declare(scope)
    }
}


/**
 *  Function Wrapper
 */

 function wrap (context, proto, vals, desc, raw) {
     assert(context instanceof Scope)
     assert(is(proto, Prototype))
     assert(is(vals, Uni(Type.Null, Type.Container.Hash)))
     assert(is(raw, ES.Function))
     assert(is(desc, Type.String))
     let err = new ErrorProducer(CallError, desc)
     let invoke = (args, use_ctx = null, check = true) => {
         // check arguments
         if (check) {
             let result = check_args(args, proto, true)
             if (result != 'OK') {
                 err.throw(result)
             }
         }
         // generate scope
         let scope = new Scope(
             (use_ctx !== null)? use_ctx: context,
             proto.affect
         )
         inject_args(args, proto, scope)
         if (vals != null) {
             list(mapkv(vals, (k, v) => scope.declare(k, v)))
         }
         // call the raw function
         let value = raw(scope)
         // check the returned value
         err.assert(is(value, proto.value), MSG.retval_invalid)
         return value
     }
     // wrap function
     let wrapped = give_arity(
         ((...args) => invoke(args, null)),
         proto.parameters.length
     )
     wrapped[WrapperInfo] = Object.freeze({
         context, invoke, proto, vals, desc, raw
     })
     return wrapped
 }

function check_args (args, proto, get_err_msg = false) {
    // IMPORTANT: return string, "OK" = valid
    let r = proto.parameters.length
    let g = args.length
    // check if argument quantity correct
    if (r != g) {
        return get_err_msg? MSG.arg_wrong_quantity(r, g): 'NG'
    }
    // check constraints
    for (let i=0; i<proto.parameters.length; i++) {
        let parameter = proto.parameters[i]
        let arg = args[i]
        let name = parameter.name
        // check if the argument matches constraint
        if( !is(arg, parameter.constraint) ) {
            return get_err_msg? MSG.arg_invalid(name): 'NG'
        }
        // cannot pass immutable object as dirty argument
        let is_dirty = parameter.pass_policy == 'dirty'
        let is_immutable = IsIm(arg)
        if (is_dirty && is_immutable) {
            return get_err_msg? MSG.arg_immutable(name): 'NG'
        }
    }
    return 'OK'
}

function inject_args (args, proto, scope) {
    for (let i=0; i<proto.parameters.length; i++) {
        let parameter = proto.parameters[i]
        let arg = args[i]
        // apply default values of schema
        if (parameter.constraint instanceof Schema) {
            assert(is(arg, Type.Container.Hash))
            let schema = parameter.constraint
            if (schema.defaults != null) {
                let defaults_applied = mapval(arg, x => x)
                for (let key of Object.keys(schema.defaults)) {
                    if (!has(key, arg)) {
                        defaults_applied[key] = schema.defaults[key]
                    }
                }
                arg = defaults_applied
            }
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
    assert(is(f, Type.Function.Wrapped))
    f = cancel_binding(f)
    let info = f[WrapperInfo]
    let g = give_arity(
        ((...args) => info.invoke(args, context)),
        info.proto? info.proto.parameters.length: 0
    )
    let invoke = function (args, use_context = null) {
        assert(use_context === null)
        return info.invoke(args, context)
    }
    g[WrapperInfo] = { original: f, invoke: invoke }
    return g
}

function cancel_binding (f) {
    assert(is(f, Type.Function.Wrapped))
    return f[WrapperInfo].original || f
}

function call (f, args) {
    if (is(f, Type.Function.Wrapped)) {
        // TODO: add frame to call stack (add info for debugging)
        return f[WrapperInfo].invoke(args)
        // TODO: remove frame from call stack
    } else if (is(f, Type.Abstract.Class)) {
        return call(f.create, args)
    } else if (typeof f == 'function') {
        return Function.prototype.apply.call(f, null, args)
    } else {
        (new ErrorProducer(CallError)).throw('calling non-callable object')
    }
}

function fun (decl_string, body) {
    let parsed = parse_decl(decl_string)
    return wrap(Global, parsed.proto, null, parsed.name, (scope, expose) => {
        return body.apply(
            null,
            list(cat([scope, expose], map(
                parsed.proto.parameters,
                p => scope.lookup(p.name)
            )))
        )
    })
}


/**
 *  Overload Tools
 */

let SoleList = list_of(Type.Function.Wrapped.Sole)

function overload (functions, desc = '') {
    assert(is(functions, SoleList))
    assert(is(desc, Type.String))
    let only1 = (functions.length == 1)
    let invoke = !only1? ((args, caller_scope, use_context = null) => {
        for (let f of rev(functions)) {
            let info = f[WrapperInfo]
            if (check_args(args, info.proto, caller_scope) === 'OK') {
                return info.invoke(args, caller_scope, use_context)
            }
        }
        let err = new ErrorProducer(CallError, desc)
        err.throw(MSG.no_matching_function)
    }): functions[0][WrapperInfo].invoke
    let o = ((...args) => invoke(args, null))
    functions = Object.freeze(functions)
    o[WrapperInfo] = Object.freeze({ functions, invoke, desc })
    return o
}

function overload_added (f, o) {
    assert(is(o, Type.Function.Wrapped.Overload))
    return overload([...o[WrapperInfo].functions, f])
}

function overload_concated (o2, o1) {
    assert(is(o2, Type.Function.Wrapped.Overload))
    assert(is(o1, Type.Function.Wrapped.Overload))
    return overload(list(
        cat(o1[WrapperInfo].functions, o2[WrapperInfo].functions)
    ))
}
