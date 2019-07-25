promise {
    function delayed_value (x: Any, error: Bool) -> Promise {
        return create_promise << lambda(resolve, reject) {
            let e = custom_error('rejected')
            set_timeout(50) -> subscribe -> lambda {
                if error {
                    reject(e)
                } else {
                    resolve(x)
                }
            }
        }
    }
    await x = create_promise << lambda(resolve, reject) {
        delayed_value(1, false) -> then -> lambda(x) {
            return delayed_value(2+x, false)
        } -> then -> lambda(x) {
            assert x == 3
            return delayed_value(0, true)
        } -> catch -> lambda(error) {
            resolve(5)
        }
    }
    await y = create_promise << lambda(resolve, reject) {
        delayed_value(7, false) -> finally -> lambda(value) {
            resolve(value)
        }
    }
    await z = create_promise << lambda(resolve, reject) {
        delayed_value(0, true) -> finally -> lambda(error) {
            assert error is Error
            resolve(11)
        }
    }
    assert x*y*z == 5*7*11
}
