function try_to_get_promise (value) {
    if (is(value, Types.Awaitable)) {
        return prms(value)
    } else {
        return value
    }
}


function try_to_forward_promise (value, resolve, reject) {
    if (is(value, Types.Awaitable)) {
        let p = prms(value)
        p.then(x => resolve(x))
        p.catch(e => reject(e))
    } else {
        resolve(value)
    }
}


pour(built_in_functions, {
    postpone: fun (
        'function postpone (time: Size) -> Promise',
            time => new Promise(resolve => {
                setTimeout(() => resolve(Nil), time)
            })
    ),
    set_timeout: fun (
        'function set_timeout (time: Size) -> Observer',
            time => observer(push => {
                let t = setTimeout(() => {
                    push(Nil)
                    push(Complete)
                }, time)
                return () => {
                    clearTimeout(t)
                }
            })
    ),
    set_interval: fun (
        'function set_interval (time: Size) -> Observer',
            time => observer(push => {
                let i = setInterval(() => {
                    push(Nil)
                }, time)
                return () => {
                    clearInterval(i)
                }
            })
    ),
    create_promise: fun (
        'function create_promise (f: Arity<2>) -> Promise',
            f => {
                return new Promise((resolve, reject) => {
                    let wrapped_resolve = fun (
                        'function resolve (value: Any) -> Void',
                            value => {
                                resolve(value)
                                return Void
                            }
                    )
                    let wrapped_reject = fun (
                        'function reject (error: Error) -> Void',
                            error => {
                                reject(error)
                                return Void
                            }
                    )
                    call(f, [wrapped_resolve, wrapped_reject])
                })
            }
    ),
    then: fun (
        'function then (a: Awaitable, f: Arity<1>) -> Promise',
            (a, f) => {
                let p = prms(a)
                return p.then(value => {
                    return try_to_get_promise(call(f, [value]))
                })
            },
    ),
    catch: fun (
        'function catch (a: Awaitable, f: Arity<1>) -> Promise',
            (a, f) => {
                let p = prms(a)
                return p.catch(error => {
                    if (is_fatal(error)) {
                        throw error
                    } else {
                        return try_to_get_promise(call(f, [error]))
                    }
                })
            }
    ),
    finally: fun (
        'function finally (a: Awaitable, f: Arity<1>) -> Promise',
            (a, f) => {
                let p = prms(a)
                return new Promise((resolve, reject) => {
                    p.then(value => {
                        try {
                            let ret = call(f, [value])
                            try_to_forward_promise(ret, resolve, reject)
                        } catch (error) {
                            reject(error)
                        }
                    }, error => {
                        if (is_fatal(error)) {
                            reject(error)
                        } else {
                            try {
                                let ret = call(f, [error])
                                try_to_forward_promise(ret, resolve, reject)
                            } catch (error) {
                                reject(error)
                            }
                        }
                    })
                })
            }
    )
})
