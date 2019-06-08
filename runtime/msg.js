const INDENT = '    '

const MSG = {
    variable_not_declared: name => `variable ${name} not declared`,
    variable_not_found: name => `variable ${name} not found`,
    variable_declared: name => `variable ${name} already declared`,
    variable_invalid: name => `invalid value assigned to variable ${name}`,
    variable_fixed: name => `cannot reset fixed variable ${name}`,
    static_conflict: name => `static value conflict with argument ${name}`,
    arg_wrong_quantity: (r, g) => `${r} arguments required but ${g} given`,
    arg_invalid: name => `invalid argument ${name}`,
    arg_require_bool: name => `lazy argument ${name} requires a boolean value`,
    arg_invalid_inflate: name => (
        `invalid template argument ${name}: neither a type nor a primitive`
    ),
    retval_invalid: 'invalid return value',
    retval_invalid_inflate: 'return value of type template should be a type',
    non_callable: 'unable to call non-callable object',
    no_matching_function: available => (
        'invalid arguments: no matching function'
        + LF + LF + 'Available functions are: '
        + LF + LF + available
    ),
    superset_invalid: i => (
        `superset #${i} is invalid (should be Class or Interface)`
    ),
    method_conflict: (name, X1, X2) => (
        'method conflict:'
        + LF + INDENT + X1
        + LF + 'and'
        + LF + INDENT + X2
        + LF + `both have method: ${name}()`
    ),
    method_missing: (name, C, I) => (
        `The ${C}` + LF + 'does not implement'
        + LF + INDENT + I
        + LF + `(missing method ${name}())`
    ),
    method_invalid: (name, C, I) => (
        `The ${C}` + LF + 'does not implement'
        + LF + INDENT + I
        + LF + `(invalid method ${name}())`
    ),
    interface_invalid: name => (
        `invalid interface: blank method ${name}() should not be implemented`
    ),
    exposing_non_instance: 'unable to expose non-instance object',
    not_exposing: C => `created instance does not expose instance of ${C}`,
    method_not_found: name => `method ${name}() does not exist`,
    format_invalid_key: key => (
        `key '${key}' does not exist in given Hash or Structure`
    ),
    format_invalid_index: index => (
        `${'${'+(index+1)+'}'} (index ${index}) does not exist in given list`
    ),
    format_not_all_converted: (
        'not all arguments converted during formatting string'
    ),
    key_error: key => `key error: requested key '${key}' does not exist`,
    index_error: index => `index ${index} out of range`,
    not_bool: 'given expression did not evaluate to a boolean value',
    not_promise: 'expression to await did not evaluate to a Promise or Future',
    not_iterable: i => `comprehension argument #${i} is not iterable`,
    filter_not_bool: 'given filter function did not return a boolean value',
    invalid_range: (a, b) => `begin index ${a} is bigger than end index ${b}`,
    empty_list: 'invalid element access on empty list',
    invalid_slice: (a, b) => `invalid slice index pair (${a}, ${b})`,
    invalid_splice: a => `invalid splice amount ${a}`,
    invalid_struct_init_miss: k => (
        `invalid structure initialization: missing field '${k}'`
    ),
    invalid_struct_init_key: k => (
        `invalid structure initialization: invalid value for field '${k}'`
    ),
    invalid_struct_init_req: (
        'invalid structure initialization: requirement not satisfied'
    ),
    schema_invalid_default: f => `invalid default value for field '${f}'`,
    struct_key_error: k => `field '${k}' does not exist on the structure`,
    struct_key_invalid: k => (
        `given value for field '${k}' violated the schema`
    ),
    struct_req_violated: k => (
        `given value for field '${k}' violated the requirement of the schema`
    ),
    struct_inconsistent: k => (
        `inconsistency: value of field '${k}' became violating the schema`
    ),
    different_schema: (
        'cannot apply infix operator on structures of different schema'
    ),
    not_schema: 'cannot create new structure by a non-schema object',
    no_common_class: op => (
        (`unable to find a common class of both
        instances that defined operator ${op}`).replace(/\n[\t ]*/g, ' ')
    ),
    module_conflict: mod => `conflicting module name ${mod}`,
    missing_export: (mod, exp) => (
        `exported variable '${exp}' does not exist in module ${mod}`
    ),
    import_conflict: alias => (
        `invalid import: variable '${alias}' already declared`
    ),
    import_conflict_mod: (mod, alias) => (
        `cannot import module ${mod} as '${alias}': variable already declared`
    ),
    module_not_exist: mod => `module ${mod} does not exist`,
    not_repr: p => 'string format parameter ${' + p + '} is not representable',
    when_expr_failed: (
        'all conditions evaluated to false in this when expression'
    )
}
