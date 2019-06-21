'<include> types.js';
'<include> exception.js';
'<include> features.js';
'<include> functions.js';
'<include> operators.js';
'<include> constants.js';
'<include> es.js';


let global_scope_data = {}
pour(global_scope_data, built_in_types)
pour(global_scope_data, built_in_functions)
pour(global_scope_data, built_in_constants)
pour(global_scope_data, built_in_es_compatible)
let Global = new Scope(null, global_scope_data, true)
let Eval = new Scope(Global)


let global_helpers = {
    c: call,
    o: get_operator,
    g: (o, k, nf, f, r, c) => call(get_data, [o, k, nf], f, r, c),
    s: set_data,
    cc: inject_desc(create_class, 'create_class'),
    ci: inject_desc(create_interface, 'create_interface'),
    cs: inject_desc(create_schema, 'create_schema'),
    ns: inject_desc(new_structure, 'initialize_structure'),
    ct: inject_desc(f => $(x => call(f, [x])), 'create_simple_type'),
    ctt: inject_desc(f => new TypeTemplate(f), 'create_type_template'),
    cf: one_of,  // create Finite
    ce: inject_desc((n, ns) => new Enum(n, ns), 'create_enum'),
    ef: ensure_failed,
    tf: try_failed,
    ie: inject_desc(inject_ensure_args, 'inject_ensure_args'),
    pa: wrapped_panic,
    as: wrapped_assert,
    th: wrapped_throw,
    c2f: convert_to_fatal,
    gp: get_call_stack_pointer,
    rs: restore_call_stack,
    ccs: clear_call_stack,
    f: for_loop,
    rb: inject_desc(require_bool, 'require_boolean_value'),
    a: Types.Any,
    b: Types.Bool,
    t: Type,
    it: Types.Iterator,
    pm: Types.Promise,
    i: Types.Instance,
    sid: Types.SliceIndexDefault,
    cv: create_value,
    sl: (o, lo, hi, f, r, c) => call(get_slice, [o, lo, hi], f, r, c),
    rp: inject_desc(require_promise, 'require_promise'),
    aw: async_e_wrap,
    ic: iterator_comprehension,
    lc: list_comprehension,
    wf: when_expr_failed,
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
    ins: inject_desc((m, c) => import_names(scope, m, c), 'import'),
    im: inject_desc(c => import_module(scope, c), 'import'),
    ia: inject_desc(m => import_all(scope, m), 'import'),
    __: global_helpers
})
