function require_bool (value) {
    // if (require_bool(may_not_boolean_value)) { ... }
    ensure(is(value, Types.Bool), 'not_bool')
    return value
}


function require_promise (object) {
    // await should_be_awaitable
    ensure(is(object, Types.Awaitable), 'not_awaitable')
    return prms(object)
}


function when_expr_failed () {
    ensure(false, 'when_expr_failed')
}
