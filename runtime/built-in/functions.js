let built_in_functions = {
    print: f (
        'print',
        'function print (p: Bool) -> Void',
            x => (console.log(x.toString()), Void),
        'function print (x: Number) -> Void',
            x => (console.log(x.toString()), Void),
        'function print (s: String) -> Void',
            s => (console.log(s), Void)
    ),
    custom_error: f (
        'custom_error',
        'function custom_error (msg: String) -> Error',
            msg => create_error(msg),
        'function custom_error (name: String, msg: String) -> Error',
            (name, msg) => create_error(msg, name),
        'function custom_error (name: String, msg: String, data: Hash) -> Error',
            (name, msg, data) => create_error(msg, name, data)
    ),
    map: f (
        'map',
        'function map (i: Iterable, f: Arity<1>) -> Iterable',
            (i, f) => map(i, e => call(f, [e])),
        'function map (i: Iterable, f: Arity<2>) -> Iterable',
            (i, f) => map(i, (e, n) => call(f, [e, n]))
    ),
    filter: f (
        'filter',
        'function filter (i: Iterable, f: Arity<1>) -> Iterable',
            (i, f) => filter(i, e => {
                let ok = call(f, [e])
                ensure(is(ok, Types.Bool), 'filter_not_bool')
                return ok
            }),
        'function filter (i: Iterable, f: Arity<2>) -> Iterable',
            (i, f) => filter(i, (e, n) => {
                let ok = call(f, [e, n])
                ensure(is(ok, Types.Bool), 'filter_not_bool')
                return ok
            })
    ),
    collect: fun (
        'function collect (i: Iterable) -> List',
            i => list(i)
    )
}
