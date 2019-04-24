let MSG = {
    schema_invalid_default: f => `invalid default value for field ${f}`,
    variable_not_found: name => `variable ${name} not found`,
    variable_declared: name => `variable ${name} already declared`,
    variable_not_declared: name => `variable ${name} not declared`,
    variable_cannot_reset: name => `variable ${name} is not re-assignable`,
    variable_immutable: name => `outer variable ${name} is immutable`,
    arg_wrong_quantity: (r, g) => `${r} arguments required but ${g} given`,
    arg_invalid: name => `invalid argument ${name}`,
    arg_immutable: name => `immutable value for dirty argument ${name}`,
    retval_invalid: 'invalid return value',
    non_callable: 'unable to call non-callable object',
    no_matching_function: 'invalid arguments: no matching function',
    method_conflict: (A1, name, A2) => (
        `exposed method conflict: ${A1} and ${A2} both have method ${name}`
    ),
    method_missing: (name, C, I) => (
        `class ${C} does not implement ${I} (missing method ${name})`
    ),
    method_invalid: (name, C, I) => (
        `class ${C} does not implement ${I} (invalid method ${name})`
    ),
    exposing_non_instance: 'unable to expose non-instance object',
    not_exposing: C => `created instance does not expose instance of ${C}`,
    method_not_found: name => `method ${name}() does not exist`,
    instance_immutable: M => (
        `unable to call dirty method ${M} on immutable instance`
    ),
    format_invalid_key: key => `key '${key}' does not exist in given hash`,
    format_invalid_index: index => (
        `${'${'+(index+1)+'}'} (index ${index}) does not exist in given list`
    ),
    format_not_all_converted: (
        'not all arguments converted during formatting string'
    )
}
