class Future<T> is Promiser {
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
    promise () -> Promise {
        return wrapped
    }
}
