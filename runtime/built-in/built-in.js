'<include> types/types.js';
'<include> features/features.js';
'<include> functions/functions.js';
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
    [CALL]: call,
    [OPERATOR]: get_operator,
    /* Features */
    [FOR_LOOP_ITER]: for_loop_i,
    [FOR_LOOP_ENUM]: for_loop_e,
    [FOR_LOOP_VALUE]: for_loop_v,
    [FOR_LOOP_ASYNC]: for_loop_a,
    [GET]: (o, k, nf, f, r, c) => call(get_data, [o, k, nf], f, r, c),
    [SET]: set_data,
    [SLICE]: (o, lo, hi, f, r, c) => call(get_slice, [o, lo, hi], f, r, c),
    [ITER_COMP]: iterator_comprehension,
    [LIST_COMP]: list_comprehension,
    /* Object Builders */
    [C_SINGLETON]: create_value,
    [C_CLASS]: inject_desc(create_class, 'create_class'),
    [C_INTERFACE]: inject_desc(create_interface, 'create_interface'),
    [C_SCHEMA]: inject_desc(create_schema, 'create_schema'),
    [C_STRUCT]: inject_desc(new_struct, 'initialize_structure'),
    [C_TYPE]: inject_desc(f => $(f), 'create_simple_type'),
    [C_TEMPLATE]: inject_desc(f => new TypeTemplate(f), 'create_type_template'),
    [C_FINITE]: inject_desc(one_of, 'create_finite_set_type'),
    [C_ENUM]: inject_desc((n, ns) => new Enum(n, ns), 'create_enum'),
    [C_FUN_SIG]: inject_desc(create_fun_sig, 'create_function_signature'),
    [C_TREE_NODE]: inject_desc(inflate_tree_node, 'inflate_tree_node'),
    [C_OBSERVER]: create_observer,
    /* Guards */
    [REQ_BOOL]: inject_desc(require_bool, 'require_boolean_value'),
    [REQ_TYPE]: inject_desc(require_type, 'require_type_object'),
    [REQ_PROMISE]: inject_desc(require_promise, 'require_promise'),
    [WHEN_FAILED]: inject_desc(when_expr_failed, 'when_expr_failed'),
    [MATCH_FAILED]: inject_desc(match_expr_failed, 'match_expr_failed'),
    [SWITCH_FAILED]: inject_desc(switch_cmd_failed, 'switch_cmd_failed'),
    /* Error Handling */
    [INJECT_E_ARGS]: inject_desc(inject_ensure_args, 'inject_ensure_args'),
    [ENSURE_FAILED]: ensure_failed,
    [TRY_FAILED]: try_failed,
    [ENTER_H_HOOK]: enter_handle_hook,
    [EXIT_H_HOOK]: exit_handle_hook,
    [PANIC]: wrapped_panic,
    [ASSERT]: wrapped_assert,
    [THROW]: wrapped_throw,
    /* Types */
    [T_ANY]: Types.Any,
    [T_BOOL]: Types.Bool,
    [T_VOID]: Types.Void,
    [T_TYPE]: Types.Type,
    [T_HASH]: Types.Hash,
    [T_PROMISE]: Types.Promise,
    [T_INSTANCE]: Types.Instance,
    [T_ITERATOR]: Types.Iterator,
    [T_ASYNC_ITERATOR]: Types.AsyncIterator,
    [T_SLICE_INDEX_DEF]: Types.SliceIndexDefault,
    [T_PLACEHOLDER]: Types.TypePlaceholder
}

Object.freeze(global_helpers)


function bind_method_call (scope) {
    return (obj, name, args, file, row, col) => {
        return call_method(scope, obj, name, args, file, row, col)
    }
}

function bind_wrap (scope) {
    return (proto, replace, desc, raw) => {
        return wrap(replace || scope, proto, desc, raw)
    }
}

function bind_match_pattern (scope) {
    return (is_fixed, pattern, value) => {
        return match_pattern(scope, is_fixed, pattern, value)
    }
}

let get_helpers = scope => ({
    [L_METHOD_CALL]: bind_method_call(scope),
    [L_STATIC_SCOPE]: f => get_static(f, scope),
    [L_WRAP]: bind_wrap(scope),
    [L_VAR_LOOKUP]: inject_desc(scope.lookup.bind(scope), 'lookup_variable'),
    [L_VAR_DECL]: inject_desc(scope.declare.bind(scope), 'declare_variable'),
    [L_VAR_RESET]: inject_desc(scope.reset.bind(scope), 'reset_variable'),
    [L_ADD_FUN]: inject_desc(scope.add_function.bind(scope), 'add_function'),
    [L_OP_MOUNT]: inject_desc(scope.mount.bind(scope), 'call_mount_operator'),
    [L_OP_PUSH]: inject_desc(scope.push.bind(scope), 'call_push_operator'),
    [L_IMPORT_VAR]: inject_desc((m, c) => import_names(scope, m, c), 'import'),
    [L_IMPORT_MOD]: inject_desc(c => import_module(scope, c), 'import'),
    [L_IMPORT_ALL]: inject_desc(m => import_all(scope, m), 'import'),
    [L_MATCH]: inject_desc(bind_match_pattern(scope), 'match_pattern'),
    [L_GLOBAL_HELPERS]: global_helpers
})
