'<include> types/types.js';
'<include> features/features.js';
'<include> functions.js';
'<include> operators.js';
'<include> constants.js';


let global_scope_data = {}
pour(global_scope_data, built_in_types)
pour(global_scope_data, built_in_functions)
pour(global_scope_data, built_in_constants)
let Global = new Scope(null, global_scope_data, true)
let Eval = new Scope(Global)


let global_helpers = {
    /* Core */
    c: call,
    o: get_operator,
    /* Features */
    fi: for_loop_i,
    fe: for_loop_e,
    fa: for_loop_a,
    g: (o, k, nf, f, r, c) => call(get_data, [o, k, nf], f, r, c),
    s: set_data,
    sl: (o, lo, hi, f, r, c) => call(get_slice, [o, lo, hi], f, r, c),
    ic: iterator_comprehension,
    lc: list_comprehension,
    /* Object Builders */
    cv: create_value,
    cc: inject_desc(create_class, 'create_class'),
    ci: inject_desc(create_interface, 'create_interface'),
    cs: inject_desc(create_schema, 'create_schema'),
    ns: inject_desc(new_struct, 'initialize_structure'),
    ct: inject_desc(f => $(x => call(f, [x])), 'create_simple_type'),
    ctt: inject_desc(f => new TypeTemplate(f), 'create_type_template'),
    cft: inject_desc(one_of, 'create_finite_set_type'),
    ce: inject_desc((n, ns) => new Enum(n, ns), 'create_enum'),
    cfs: inject_desc(create_fun_sig, 'create_function_signature'),
    /* Guards */
    rb: inject_desc(require_bool, 'require_boolean_value'),
    rp: inject_desc(require_promise, 'require_promise'),
    wf: inject_desc(when_expr_failed, 'when_expr_no_match'),
    /* Error Handling */
    ie: inject_desc(inject_ensure_args, 'inject_ensure_args'),
    ef: ensure_failed,
    tf: try_failed,
    enh: enter_handle_hook,
    exh: exit_handle_hook,
    pa: wrapped_panic,
    as: wrapped_assert,
    th: wrapped_throw,
    /* Types */
    a: Types.Any,
    b: Types.Bool,
    v: Types.Void,
    t: Types.Type,
    h: Types.Hash,
    i: Types.Instance,
    pm: Types.Promise,
    it: Types.Iterator,
    ait: Types.AsyncIterator,
    sid: Types.SliceIndexDefault
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
    w: (proto, replace, desc, raw) => wrap(replace || scope, proto, desc, raw),
    ins: inject_desc((m, c) => import_names(scope, m, c), 'import'),
    im: inject_desc(c => import_module(scope, c), 'import'),
    ia: inject_desc(m => import_all(scope, m), 'import'),
    __: global_helpers
})
