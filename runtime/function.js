/**
 *  Function & Scope
 */
const WrapperInfo = Symbol('WrapperInfo')
const BareFuncDesc = Symbol('BareFuncDesc')

function inject_desc (f, desc) {
    f[BareFuncDesc] = `Runtime.${desc}`
    return f
}

let Wrapped = Ins(ES.Function, $(x => typeof x[WrapperInfo] == 'object'))
let NotWrapped = Ins(ES.Function, $(x => typeof x[WrapperInfo] != 'object'))

let FunctionTypes = {
    ES_Function: NotWrapped,
    Wrapped: Wrapped,
    Function: Ins(Wrapped, $(f => has('context', f[WrapperInfo]))),
    Overload: Ins(Wrapped, $(f => has('functions', f[WrapperInfo]))),
    Binding: Ins(Wrapped, $(f => has('original', f[WrapperInfo])))
}

pour(Types, FunctionTypes)


/**
 *  Parameter & Function Prototype
 */

let Parameter = format({
    name: Types.String,
    type: Type
})

let Prototype = format({
    value_type: Type,
    parameters: TypedList.of(Parameter)
}, proto => no_repeat(map(proto.parameters, p => p.name)) )

function parse_decl (string) {
    function parse_template_arg (arg_str) {
        if (arg_str == 'true') { return true }
        if (arg_str == 'false') { return false }
        if (!Number.isNaN(Number(arg_str))) { return Number(arg_str) }
        if (arg_str.match(/^'[^']*'$/) != null) { return arg_str.slice(1, -1) }
        assert(has(arg_str, Types))
        return Types[arg_str]
    }
    let match = string.match(/function +([^( ]+) +\(([^)]*)\) *-> *(.+)/)
    let [_, name, params_str, value_str] = match
    let has_p = params_str.trim().length > 0
    let parameters = has_p? (list(map(params_str.split(','), para_str => {
        para_str = para_str.trim()
        let match = para_str.match(/([^ :]+) *: *(.+)/)
        let [_, name, type_str] = match
        let type = null
        match = type_str.match(/([^<]+)<(.+)>/)
        if (match != null) {
            let template_str = match[1]
            let args_str = match[2]
            let arg_strs = args_str.split(',').map(a => a.trim())
            let args = arg_strs.map(parse_template_arg)
            assert(has(template_str, Types))
            // note: template.inflate() already bound, apply(null, ...) is OK
            type = Types[template_str].inflate.apply(null, args)
        } else {
            assert(has(type_str, Types))
            type = Types[type_str]
        }
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
    constructor (context, data = {}, readonly = false) {
        assert(context === null || context instanceof Scope)
        assert(is(data, Types.Hash))
        this.context = context
        this.data = data
        this.cache = {}
        this.types_of_non_fixed = {}
        this.readonly = readonly
        if (readonly) {
            Object.freeze(this.data)
        }
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
        assert(!this.readonly)
        assert(is(variable, Types.String))
        assert(is(is_fixed, Types.Bool))
        assert(is(type, Type))
        ensure(!this.has(variable), 'variable_declared', variable)
        ensure(is(initial_value, type), 'variable_invalid', variable)
        this.data[variable] = initial_value
        if (!is_fixed) {
            this.types_of_non_fixed[variable] = type
        }
        return Void
    }
    define_function (name, f) {
        assert(is(name, Types.String))
        assert(is(f, Types.Function))
        let existing = this.find(name)
        if (existing !== NotFound) {
            if (is(existing, Types.Function)) {
                let g = overload([existing, f], name)
                if (this.has(name)) {
                    this.reset(name, g)
                } else {
                    this.declare(name, g, false)
                }
            } else if (is(existing, Types.Overload)) {
                let g = overload_added(f, existing, name)
                if (this.has(name)) {
                    this.reset(name, g)
                } else {
                    this.declare(name, g, false)
                }
            } else {
                this.declare(name, f, false)
            }
        } else {
            this.declare(name, f, false)
        }
        return Void
    }
    reset (variable, new_value) {
        assert(is(variable, Types.String))
        let scope = this
        while (scope !== null) {
            if (scope.has(variable)) {
                break
            } else if (has(variable, scope.cache)) {
                scope = scope.cache[variable]
                assert(scope.has(variable))
                break
            }
            scope = scope.context
        }
        ensure(scope !== null, 'variable_not_declared', variable)
        ensure(!scope.is_fixed(variable), 'variable_fixed', variable)
        let type_ok = is(new_value, scope.types_of_non_fixed[variable])
        ensure(type_ok, 'variable_invalid', variable)
        scope.data[variable] = new_value
        return Void
    }
    lookup (variable) {
        assert(is(variable, Types.String))
        let value = this.find(variable)
        ensure(value !== NotFound, 'variable_not_found', variable)
        return value
    }
    try_to_declare (variable, initial_value, is_fixed = true, type = null) {
        assert(!this.readonly)
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
            scope => scope === null
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
        }), object => object !== NotFound)
    }
}

function new_scope (context) {
    return new Scope(context)
}


/**
 *  Function Wrapper
 */

 let arg_msg = new Map([
     [1, 'arg_wrong_quantity'],
     [2, 'arg_invalid']
 ])

 function get_static (f, context) {
     assert(is(f, ES.Function))
      let scope = new Scope(context)
      f(scope)
      return scope
 }

 function wrap (context, proto, replace, desc, raw) {
     assert(context === null || context instanceof Scope)
     assert(is(proto, Prototype))
     assert(replace === null || replace instanceof Scope)
     assert(is(raw, ES.Function))
     assert(is(desc, Types.String))
     if (replace !== null) {
         context = replace
     }
     let invoke = (args, use_ctx = null, check = true) => {
         // check arguments
         if (check) {
             let result = check_args(args, proto)
             ensure (
                 result.ok, arg_msg.get(result.err),
                 result.info, args.length
             )
         }
         // generate scope
         let scope = new Scope(use_ctx || context)
         inject_args(args, proto, scope)
         let value = raw(scope)
         ensure(is(value, proto.value_type), 'retval_invalid')
         // return the value after type check
         return value
     }
     let arity = proto.parameters.length
     let wrapped = give_arity((...args) => {
         try {
             return invoke(args)
         } catch (error) {
             if (!(error instanceof RuntimeError)) {
                 clear_call_stack()
             }
             throw error
         }
     }, arity)
     foreach(proto.parameters, p => Object.freeze(p))
     Object.freeze(proto.parameters)
     Object.freeze(proto)
     let info = { context, invoke, proto, desc, raw }
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
        scope.declare(parameter.name, args[i], false, parameter.type)
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
    assert(is(args, Types.List))
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
    } else if (is(f, ES.Function)) {
        let desc = f[BareFuncDesc] || get_summary(f.toString())
        try {
            push_call(call_type, desc, file, row, col)
            let value = f.apply(null, args)
            pop_call()
            return value
        } catch (e) {
            if (!(e instanceof RuntimeError)) {
                clear_call_stack()
            }
            throw e
        }
    } else {
        push_call(call_type, '*** Non-Callable Object', file)
        ensure(false, 'non_callable')
    }
}

function fun (decl_string, body) {
    let parsed = parse_decl(decl_string)
    assert(is(body, ES.Function))
    assert(!is(body, Types.Wrapped))
    let desc = decl_string.replace(/^function */, '')
    return wrap(null, parsed.proto, null, desc, (scope, expose) => {
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
    assert(is(functions, TypedList.of(Types.Function)))
    assert(is(name, Types.String))
    functions = copy(functions)
    Object.freeze(functions)
    let desc = `${name}[${functions.length}]`
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
                result_list.push(check_args(args, info_list[i].proto))
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
                return value
            } else {
                let n = i
                let available = join(map(count(n), i => {
                    let r = result_list[i]
                    return (
                        info_list[i].desc
                        + '  (' + get_msg (
                            arg_msg.get(r.err),
                            [r.info, args.length]
                        ) + ')'
                    )
                }), LF)
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
