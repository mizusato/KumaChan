function require_bool (value) {
    // if (require_bool(may_not_boolean_value)) { ... }
    ensure(is(value, Types.Bool), 'not_bool')
    return value
}


function require_promise (object) {
    // await should_be_promise_or_awaitable
    if (is(object, Types.Promise)) {
        return object
    } else {
        ensure(is(object, Types.Promiser), 'not_awaitable')
        return call_method(null, object, 'promise', [])
    }
}


function when_expr_failed () {
    ensure(false, 'when_expr_failed')
}
