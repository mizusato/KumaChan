class Future<T> is Promiser {
    init (raw: Promise) {
        static {
            let ES_Promise = es.new(__Promise)
        }
        let fatal_error = custom_error(MSG.future_invalid_value())
        let wrapped = ES_Promise << lambda(resolve, reject) {
            raw -> then -> lambda value -> {
                ensure type_ok { value is T }
                resolve(value)
                @handle error {
                    unless type_ok {
                        panic fatal_error
                    }
                }
            } -> catch -> lambda error -> {
                reject(error)
            }
        }
    }
    promise () -> Promise {
        return wrapped
    }
}
