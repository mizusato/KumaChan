'<include> built-in/types.js';
'<include> built-in/exceptions.js';
'<include> built-in/features.js';
'<include> built-in/functions.js';
'<include> built-in/operators.js';
'<include> built-in/constants.js';


let global_scope_data = {}
let built_in_additional = {
    es: {
        undefined: undefined,
        null: null,
        Symbol: ES.Symbol
    }
}
pour(global_scope_data, built_in_types)
pour(global_scope_data, built_in_functions)
pour(global_scope_data, built_in_constants)
pour(global_scope_data, built_in_additional)
let Global = new Scope(null, global_scope_data, true)

let Eval = new Scope(Global)
let default_scopes = { Global, Eval }
Object.freeze(default_scopes)


let global_helpers = {
    a: Types.Any,
    c: call,
    o: get_operator,
    g: (o, k, nf, f, r, c) => call(get_data, [o, k, nf], f, r, c),
    s: set_data,
    cc: create_class,
    ci: create_interface,
    ef: ensure_failed,
    tf: try_failed,
    ie: inject_desc(inject_ensure_args, 'inject_ensure_args'),
    pa: wrapped_panic,
    as: wrapped_assert,
    th: wrapped_throw,
    gp: get_call_stack_pointer,
    rs: restore_call_stack,
    cs: clear_call_stack,
    f: for_loop,
    rb: inject_desc(require_bool, 'require_boolean_value'),
    it: Types.Iterator,
    v: Void
}

Object.freeze(global_helpers)


function bind_method_call (scope) {
    return (obj, name, args, file, row, col) => {
        return call_method(scope, obj, name, args, file, row, col)
    }
}

let get_helpers = scope => ({
    m: bind_method_call(scope),
    id: inject_desc(scope.lookup.bind(scope), 'lookup_variable'),
    dl: inject_desc(scope.declare.bind(scope), 'declare_variable'),
    rt: inject_desc(scope.reset.bind(scope), 'reset_variable'),
    df: inject_desc(scope.define_function.bind(scope), 'define_function'),
    gs: f => get_static(f, scope),
    w: (proto, vals, desc, raw) => wrap(scope, proto, vals, desc, raw),
    __: global_helpers
})
