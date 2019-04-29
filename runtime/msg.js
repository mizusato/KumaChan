const INDENT = '    '

const MSG = {
    schema_invalid_default: f => `invalid default value for field ${f}`,
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
        `invalid template argument: ${name} is neither a type nor a primitive`
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
        + LF + 'both have method: ' + name
    ),
    method_missing: (name, C, I) => (
        'the class'
        + LF + INDENT + C
        + LF + 'does not implement'
        + LF + INDENT + I
        + LF + `(missing method ${name})`
    ),
    method_invalid: (name, C, I) => (
        'the class'
        + LF + INDENT + C
        + LF + 'does not implement'
        + LF + INDENT + I
        + LF + `(invalid method ${name})`
    ),
    interface_invalid: name => (
        `invalid interface definition: implemented blank method ${name}`
    ),
    exposing_non_instance: 'unable to expose non-instance object',
    not_exposing: C => `created instance does not expose instance of ${C}`,
    method_not_found: name => `method ${name}() does not exist`,
    format_invalid_key: key => `key '${key}' does not exist in given hash`,
    format_invalid_index: index => (
        `${'${'+(index+1)+'}'} (index ${index}) does not exist in given list`
    ),
    format_not_all_converted: (
        'not all arguments converted during formatting string'
    )
}
