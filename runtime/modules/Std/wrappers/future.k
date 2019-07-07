class Future<T> {
    init (raw: Promise) {
        let type_error = custom_error('TypeError', MSG.future_invalid_value())
        let wrapped = promise {
            await value = raw
            if value is not T {
                panic type_error
            }
            return value
        }
    }
    get_wrapped () -> Promise {
        return wrapped
    }
    operator prms (f) {
        return f -> get_wrapped()
    }
}
