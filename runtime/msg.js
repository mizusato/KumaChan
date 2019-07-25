/**
 *  Error Message Templates
 */


const INDENT = '    '

const MSG = {
    invalid_code_point: index => `invalid code point 0x${index}`,
    variable_not_declared: name => `variable ${name} not declared`,
    variable_not_found: name => `variable ${name} not found`,
    variable_declared: name => `variable ${name} already declared`,
    variable_invalid: name => `invalid value assigned to variable ${name}`,
    variable_fixed: name => `cannot reset fixed variable ${name}`,
    variable_inconsistent: (name, p) => (
        `the value of variable ${name} violated its type constraint`
        + (p? `, called through UFCS at ${p}`: '')
    ),
    static_conflict: name => `static value conflict with argument ${name}`,
    arg_wrong_quantity: (r, g) => `${r} arguments required but ${g} given`,
    arg_invalid: name => `invalid argument ${name}`,
    arg_require_bool: name => `lazy argument ${name} requires a boolean value`,
    arg_invalid_inflate: name => (
        `invalid template argument ${name}: neither a type nor a primitive`
    ),
    retval_invalid: 'invalid return value',
    retval_invalid_inflate: 'return value of type template should be a type',
    non_callable: p => `unable to call non-callable object at ${p}`,
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
    operator_conflict: (op, C1, C2) => (
        `operator ${op} defined in ${C1}`
        + LF + `conflicts with the operator ${op} defined in`
        + LF + INDENT + C2
    ),
    interface_invalid: name => (
        `invalid interface: blank method ${name}() should not be implemented`
    ),
    pf_conflict: name => (
        `unable to declare private function ${name}: conflict with arguments`
    ),
    self_conflict: (
        `unable to declare 'self' reference: variable name already used`
    ),
    invalid_push: 'push operator is only available inside observer',
    invalid_mount: 'mount operator is only available inside initializer',
    mounting_non_instance: 'unable to mount non-instance object',
    mounting_undeclared: C => (
        'unable to mount instance of undeclared base class:'
        + LF + INDENT + C
    ),
    mounting_mounted: 'an instance of the same base class is already mounted',
    not_mounting: C => `created instance did not mount an instance of ${C}`,
    method_not_found: (name, p) => (
        `method ${name}() does not exist, called at ${p}`
    ),
    creator_returned_invalid: 'the creator returned an invalid instance',
    format_invalid_key: key => (
        `key '${key}' does not exist in given Hash or Struct`
    ),
    format_invalid_index: index => (
        `${'${'+(index+1)+'}'} (index ${index}) does not exist in given list`
    ),
    format_not_all_converted: (
        'not all arguments converted during formatting string'
    ),
    get_from_nil: 'unable to perform get operation on Nil without nil flag',
    key_error: key => `key error: requested key '${key}' does not exist`,
    index_error: index => `index ${index} out of range`,
    not_bool: 'given expression did not evaluate to a boolean value',
    not_awaitable: 'expression to await did not evaluate to a awaitable value',
    not_iterable: i => `comprehension argument #${i} is not iterable`,
    element_not_iterable: 'flat(): non-iterable element appeared in sequence',
    filter_not_bool: 'given filter function did not return a boolean value',
    cond_not_bool: 'given condition function did not return a boolean value',
    element_not_string: 'non-string element was found in the iterable object',
    invalid_range: (a, b) => (
        `invalid range: begin index ${a} is bigger than end index ${b}`
    ),
    empty_list: 'invalid element access on empty list',
    invalid_slice: 'invalid slice: lower bound is bigger than higher bound',
    slice_index_error: index => `slice index ${index} out of range`,
    invalid_splice: a => `invalid splice amount ${a}`,
    invalid_struct_init_miss: k => (
        `invalid structure initialization: missing field '${k}'`
    ),
    invalid_struct_init_key: k => (
        `invalid structure initialization: invalid value for field '${k}'`
    ),
    schema_invalid_field: f => `constraint given for field ${f} is not a type`,
    schema_invalid_default: f => `invalid default value for field '${f}'`,
    schema_field_conflict: f => `field '${f}' was defined twice or more`,
    struct_field_missing: k => `field '${k}' does not exist on the struct`,
    struct_field_invalid: k => (
        `given value for field '${k}' violated the schema of the struct`
    ),
    struct_inconsistent: k => (
        `the value of field '${k}' became violating the schema`
    ),
    struct_nil_flag: 'cannot use nil flag on struct field',
    different_schema: (
        'cannot apply infix operator on structures of different schema'
    ),
    not_schema: 'cannot create new structure by a non-schema object',
    bad_entry_list: (
        'bad entry list: the number of keys != the number of values'
    ),
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
    when_expr_failed: 'all conditions evaluated to false in when expression',
    switch_cmd_failed: 'all conditions evaluated to false in switch command',
    signature_invalid_arg: i => (
        `invalid function signature: non-type object given for parameter #${i}`
    ),
    signature_invalid_ret: (
        'invalid function signature: non-type object given for return value'
    ),
    invalid_cast: 'invalid type cast: object was cast to non-desired type',
    invalid_copy: (
        'bad copy: new object must be an instance of the class of old one'
    ),
    did_not_copy: 'bad copy: new object is the same as old one',
    replace_not_string: (
        'bad string replace: given function returned a non-string value'
    ),
    regexp_invalid: 'invalid regular expression',
    sqrt_of_negative: (
        'unable to find the real square root of a negative number'
    ),
    log_of_non_positive: (
        'unable to find the real logarithm of a non-positive number'
    ),
    asin_out_of_domain: (
        'given argument exceeded the domain of arcsin function'
    ),
    acos_out_of_domain: (
        'given argument exceeded the domain of arccos function'
    ),
    match_non_getter: (
        'cannot perform (sub-)pattern matching on a non-GeneralGetter object'
    ),
    match_nil: (
        'cannot perform (sub-)pattren matching on Nil without nil flag'
    ),
    push_observer_closed: 'invalid push: observer already closed',
    push_no_error_handler: (
        'invalid push: no error handler registered on current subscriber'
    ),
    invalid_unsub: 'this observer cannot be unsubscribed',
    redundant_unsub: 'subscription already cancelled'
}
