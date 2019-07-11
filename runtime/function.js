/**
 *  Function & Scope
 *
 *  In order to provide runtime type check for functions
 *    and change the behaviour of variable declaration,
 *    it is necessary to wrap ordinary JavaScript functions
 *    into Wrapped Functions and define our own scope objects.
 *  A Wrapped Function is also a valid JavaScript function,
 *    but a [WrapperInfo] object is added to it.
 */
const WrapperInfo = Symbol('WrapperInfo')
const BareFuncDesc = Symbol('BareFuncDesc')

let Wrapped = Ins(ES.Function, $(x => typeof x[WrapperInfo] == 'object'))
let Bare = Ins(ES.Function, $(x => typeof x[WrapperInfo] != 'object'))
let ArgError = new Map([
    [1, 'arg_wrong_quantity'],
    [2, 'arg_invalid']
])

pour(Types, {
    ES_Function: Bare,
    Wrapped: Wrapped,
    Function: Ins(Wrapped, $(f => has('context', f[WrapperInfo]))),
    Overload: Ins(Wrapped, $(f => has('functions', f[WrapperInfo]))),
    Binding: Ins(Wrapped, $(f => has('pointer', f[WrapperInfo])))
})

function call_by_js (f, args) {
    assert(is(f, Types.Wrapped))
    assert(is(args, Types.List))
    return f.apply(null, args)
}

function inject_desc (f, desc) {
    assert(is(f, Bare))
    // inject description for bare JavaScript functions
    f[BareFuncDesc] = `Runtime.${desc}`
    // the injected description text will be shown in call stack backtrace
    return f
}

inject_desc(call_by_js, 'call_by_js')


/**
 *  Definition of Parameter & Function Prototype
 */
let Parameter = format({
    name: Types.String,
    type: Type
})

let Prototype = format({
    value_type: Type,
    parameters: TypedList.of(Parameter)
}, proto => no_repeat(map(proto.parameters, p => p.name)) )


/**
 *  Scope Object
 *
 *  There are two kinds of variables can be declared in a scope,
 *    1. fixed variable: defined by `let NAME [:TYPE] = VALUE`
 *    2. non-fixed variable: defined by `var NAME [:TYPE] = VALUE`
 *  The value of fixed variable cannot be changed to another value,
 *    but it does NOT mean it's a constant, because the inner structure of
 *    the value actually can be changed, for example, the command
 *    `let hash = { 'a': 0 }; set hash['a'] = 1` is legal, but the command
 *    `let num = 0; reset num = 1` is illegal.
 */
class Scope {
    constructor (context, data = {}, read_only = false) {
        assert(context === null || context instanceof Scope)
        assert(is(data, Types.Hash))
        // upper scope
        this.context = context
        // variables in this scope
        this.data = copy(data)
        // types of non-fixed variables
        this.types_of_non_fixed = {}
        // is it read-only? (e.g. the global scope is read-only)
        this.read_only = read_only
        // bound operator functions
        this.operators = {}
        if (read_only) {
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
        assert(!this.read_only)
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
    add_function (name, f) {
        // overload functions with a common name
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
    define_mount (f) {
        assert(is(f, ES.Function))
        this.operators.mount = f
    }
    mount (i) {
        let s = find(this.iter_scope_chain(), s => s.operators.mount != null)
        ensure(s !== NotFound, 'invalid_mount')
        assert(is(s.operators.mount, ES.Function))
        return s.operators.mount
    }
    reset (variable, new_value) {
        assert(is(variable, Types.String))
        let scope = this
        while (scope !== null) {
            if (scope.has(variable)) {
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
    try_to_declare (variable, initial_value, is_fixed, type) {
        assert(!this.read_only)
        assert(is(variable, Types.String))
        if (!this.has(variable)) {
            this.declare(variable, initial_value, is_fixed, type)
        }
    }
    iter_scope_chain () {
        return iterate (
            this,
            scope => scope.context,
            scope => scope === null
        )
    }
    find (variable) {
        assert(is(variable, Types.String))
        return find(map(this.iter_scope_chain(), (scope, depth) => {
            if (scope.has(variable)) {
                let value = scope.data[variable]
                // got a fixed variable
                if (scope.is_fixed(variable)) {
                    return value
                }
                // got a non-fixed variable, check consistency
                let ok = is(value, scope.types_of_non_fixed[variable])
                ensure(ok, 'variable_inconsistent', variable)
                return value
            } else {
                return NotFound
            }
        }), object => object !== NotFound)
    }
    find_function (variable) {
        // used by `call_method()` in `oo.js`
        assert(is(variable, Types.String))
        let result = find(map(this.iter_scope_chain(), (scope, depth) => {
            if (scope.has(variable)) {
                let value = scope.data[variable]
                // not function, lookup upper scope
                if (!is(value, ES.Function)) {
                    return { value: NotFound }
                }
                // got a fixed variable
                if (scope.is_fixed(variable)) {
                    return { value, ok: true }
                }
                // got a non-fixed variable, check consistency
                let ok = is(value, scope.types_of_non_fixed[variable])
                return { value, ok }
            } else {
                return { value: NotFound }
            }
        }), result => result.value !== NotFound)
        return (result !== NotFound)? result: { value: NotFound }
    }
}

/* shorthand */
let new_scope = context => new Scope(context)


/**
 *  Wraps a bare function
 *
 *  @param context Scope
 *  @param proto Prototype
 *  @param desc string
 *  @param raw Bare
 *  @return Function
 */
function wrap (context, proto, desc, raw) {
    assert(context === null || context instanceof Scope)
    assert(is(proto, Prototype))
    assert(is(raw, Bare))
    assert(is(desc, Types.String))
    let invoke = (args, use_ctx = null, check = true) => {
        // arguments may have been checked by overload
        if (check) {
            let result = check_args(args, proto)
            ensure (
                result.ok, ArgError.get(result.err),
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
    let wrapped = give_arity (
        (...args) => call(wrapped, args, '<JS>'),
        proto.parameters.length
    )
    foreach(proto.parameters, p => Object.freeze(p))
    Object.freeze(proto.parameters)
    Object.freeze(proto)
    let info = { context, invoke, proto, desc, raw }
    Object.freeze(info)
    wrapped[WrapperInfo] = info
    Object.freeze(wrapped)
    return wrapped
}


/**
 *  Creates a scope for static variables
 *
 *  @param f function
 *  @param context Scope
 *  @return Scope
 */
function get_static (f, context) {
    assert(is(f, Bare))
    let scope = new Scope(context)
    f(scope) // execute the static block
    return scope
}


/**
 *  Checks if arguments are valid according to a function prototype
 *
 *  @param args array
 *  @param proto Prototype
 *  @return { ok: boolean, err: key of `ArgError`, info: any }
 */
function check_args (args, proto) {
    // no type assertion here, because only called by `wrap()` and `overload()`
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


/**
 *  Injects arguments to specified scope
 *
 *  @param args array
 *  @param proto Prototype
 *  @param scope Scope
 */
function inject_args (args, proto, scope) {
    // no type assertion here, because only called by `wrap()`
    let arity = proto.parameters.length
    for (let i=0; i<arity; i++) {
        let parameter = proto.parameters[i]
        scope.declare(parameter.name, args[i], false, parameter.type)
    }
}


/**
 *  Overloads some wrapped functions into a single wrapped function
 *
 *  @param functions Function[]
 *  @param name string
 *  @return Overload
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
        // this special handling makes error message more clear
        invoke = (args, use_ctx = null) => {
            let info = functions[0][WrapperInfo]
            push_call(2, info.desc)
            try {
                return info.invoke(args, use_ctx)
            } catch (e) {
                throw e
            } finally {
                pop_call()
            }
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
                try {
                    return info.invoke(args, use_ctx)
                } catch (e) {
                    throw e
                } finally {
                    pop_call()
                }
            } else {
                let n = i
                let available = join(map(count(n), i => {
                    let r = result_list[i]
                    return (
                        info_list[i].desc
                        + '  (' + get_msg (
                            ArgError.get(r.err),
                            [r.info, args.length]
                        ) + ')'
                    )
                }), LF)
                ensure(false, 'no_matching_function', available)
            }
        }
    }
    let o = (...args) => call(o, args, '<JS>')
    let info = Object.freeze({ functions, invoke, name, desc })
    Object.freeze(info)
    o[WrapperInfo] = info
    Object.freeze(o)
    return o
}


/**
 *  Creates a new overload with a new function added to the old one
 *
 *  @param f Function
 *  @param o Overload
 *  @param name string
 *  @return Overload
 */
function overload_added (f, o, name) {
    assert(is(o, Types.Overload))
    return overload([...o[WrapperInfo].functions, f], name)
}


/**
 *  Creates a new overload from `o1` concatenated with `o2`
 *
 *  @param o2 Overload
 *  @param o1 Overload
 *  @param name string
 *  @return Overload
 */
function overload_concated (o2, o1, name) {
    assert(is(o2, Types.Overload))
    assert(is(o1, Types.Overload))
    return overload(list(
        cat(o1[WrapperInfo].functions, o2[WrapperInfo].functions)
    ), name)
}


/**
 *  Creates a function binding pointing to `f` with replaced context
 *
 *  @param f Wrapped
 *  @param context Scope
 *  @return Binding
 */
function bind_context (f, context) {
    assert(is(f, Types.Wrapped))
    assert(context instanceof Scope)
    f = cancel_binding(f)
    let f_invoke = f[WrapperInfo].invoke
    let desc = f[WrapperInfo].desc
    let invoke = function (args, use_ctx = null) {
        assert(use_ctx === null)
        return f_invoke(args, context)
    }
    let binding = (...args) => call(binding, args, '<JS>')
    let info = { invoke, desc, pointer: f, bound: context }
    Object.freeze(info)
    binding[WrapperInfo] = info
    Object.freeze(binding)
    return binding
}


/**
 *  Creates a new function as `f` embraced into another context
 *
 *  @param f Wrapped
 *  @param context Scope
 *  @return Wrapped
 */
function embrace_in_context (f, context) {
    assert(is(f, Types.Wrapped))
    assert(context instanceof Scope)
    f = cancel_binding(f)
    if (is(f, Types.Overload)) {
        return overload(f[WrapperInfo].functions.map(F => {
            F = F[WrapperInfo]
            return wrap(context, F.proto, F.desc, F.raw)
        }), f[WrapperInfo].name)
    } else if(is(f, Types.Function)) {
        let F = f[WrapperInfo]
        return wrap(context, F.proto, F.desc, F.raw)
    } else {
        assert(false)
    }
}


/**
 *  Deref `f` if it is a binding
 *
 *  @param f Wrapped
 *  @return Wrapped
 */
function cancel_binding (f) {
    assert(is(f, Types.Wrapped))
    return f[WrapperInfo].pointer || f
}


/**
 *  Calls a callable object, put debug info onto the call stack
 *
 *  @param f Callable
 *  @param args array
 *  @param file string
 *  @param row integer
 *  @param col integer
 */
function call (f, args, file = null, row = -1, col = -1) {
    assert(is(args, Types.List))
    if (is(f, Types.TypeTemplate)) {
        f = f.inflate
    } else if (is(f, Types.Class)) {
        f = f.create
    } else if (is(f, Types.Schema)) {
        f = f.create_struct_from_another
    }
    let call_type = file? CALL_FROM_SCRIPT: CALL_FROM_BUILT_IN
    if (is(f, Types.Wrapped)) {
        let info = f[WrapperInfo]
        push_call(call_type, info.desc, file, row, col)
        try {
            return info.invoke(args)
        } catch (e) {
            throw e
        } finally {
            pop_call()
        }
    } else if (is(f, ES.Function)) {
        let desc = f[BareFuncDesc] || get_summary(f.toString())
        push_call(call_type, desc, file, row, col)
        try {
            return f.apply(null, args)
        } catch (e) {
            throw e
        } finally {
            pop_call()
        }
    } else {
        ensure(false, 'non_callable', `${file} (row ${row}, column ${col})`)
    }
}


/**
 *  Function Declaration Parser for Built-in Functions
 *
 *  @param string function FUNC_NAME (NAME1: TYPE1, NAME2: TYPE2, ...) -> TYPE
 *  @return { name: String, proto: Prototype }
 */
function parse_decl (string) {
    function parse_template_arg (arg_str) {
        if (arg_str == 'true') { return true }
        if (arg_str == 'false') { return false }
        if (!Number.isNaN(Number(arg_str))) { return Number(arg_str) }
        if (arg_str.match(/^'[^']*'$/) != null) { return arg_str.slice(1, -1) }
        assert(has(arg_str, Types))
        return Types[arg_str]
    }
    function parse_type (type_str) {
        // Note: Type names are resolved by `Types[TYPE_NAME]`
        match = type_str.match(/([^<]+)<(.+)>/)
        if (match != null) {
            let template_str = match[1]
            let args_str = match[2]
            let arg_strs = args_str.split(',').map(a => a.trim())
            let args = arg_strs.map(parse_template_arg)
            assert(has(template_str, Types))
            // note: template.inflate() already bound, apply(null, ...) is OK
            return Types[template_str].inflate.apply(null, args)
        } else {
            assert(has(type_str, Types))
            return Types[type_str]
        }
    }
    let match = string.match(/function +([^( ]+) +\(([^)]*)\) *-> *(.+)/)
    let [_, name, params_str, value_str] = match
    let has_p = params_str.trim().length > 0
    let parameters = has_p? (list(map(params_str.split(','), para_str => {
        para_str = para_str.trim()
        let match = para_str.match(/([^ :]+) *: *(.+)/)
        let [_, name, type_str] = match
        let type = parse_type(type_str)
        let parameter = { name, type }
        Object.freeze(parameter)
        return parameter
    }))): []
    Object.freeze(parameters)
    let value_type = parse_type(value_str)
    let proto = { parameters, value_type }
    Object.freeze(proto)
    assert(is(proto, Prototype))
    return { name, proto }
}


/**
 *  Built-in Function Creator
 *
 *  @param decl_string string
 *  @param body function (ARG1, ARG, ...) => RETURN_VALUE
 *  @return Function
 */
function fun (decl_string, body) {
    let parsed = parse_decl(decl_string)
    assert(is(body, ES.Function))
    assert(!is(body, Types.Wrapped))
    let desc = decl_string.replace(/^function */, '')
    return wrap(null, parsed.proto, desc, (scope, expose) => {
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
 *  Built-in Overload Creator
 *
 *  @param name string
 *  @param args array of (string | function)
 *  @return Overload
 */
function f (name, ...args) {
    // see usage at `built-in/functions.js`
    assert(args.length % 2 == 0)
    let functions = list(map(
        count(args.length/2),
        i => fun(args[i*2], args[i*2+1])
    ))
    return overload(functions, name)
}
