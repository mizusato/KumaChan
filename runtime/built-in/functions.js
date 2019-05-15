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
    postpone: fun (
        'function postpone (time: Size, callback: Arity<0>) -> Void',
            (time, callback) => {
                let frame = get_top_frame()
                let pos = ''
                if (frame != null) {
                    let { file, row, col } = frame
                    pos = `at ${file} (row ${row}, column ${col})`
                }
                setTimeout (
                    () => {
                        call(callback, [], `postpone(${time}) ${pos}`)
                    }, time
                )
                return Void
            }
    ),
    timeout: fun (
        'function timeout (time: Size) -> Promise',
            time => new Promise(resolve => {
                setTimeout(() => resolve(Nil), time)
            })
    ),
    count: f (
        'count',
        'function count (n: Size) -> Iterator',
            n => count(n),
        'function count (start: Index, amount: Size) -> Iterator',
            (start, amount) => map(count(amount), i => start + i)
    ),
    range: fun (
        'function range (begin: Index, end: Index) -> Iterator',
            (begin, end) => {
                ensure(begin <= end, 'invalid_range', begin, end)
                return (function* () {
                    for (let i = begin; i < end; i++) {
                        yield i
                    }
                })()
            }
    ),
    map: f (
        'map',
        'function map (i: Iterable, f: Arity<1>) -> Iterator',
            (i, f) => map(i, e => call(f, [e])),
        'function map (i: Iterable, f: Arity<2>) -> Iterator',
            (i, f) => map(i, (e, n) => call(f, [e, n]))
    ),
    filter: f (
        'filter',
        'function filter (i: Iterable, T: Type) -> Iterator',
            (i, T) => filter(i, e => is(e, T)),
        'function filter (i: Iterable, f: Arity<1>) -> Iterator',
            (i, f) => filter(i, e => {
                let ok = call(f, [e])
                ensure(is(ok, Types.Bool), 'filter_not_bool')
                return ok
            }),
        'function filter (i: Iterable, f: Arity<2>) -> Iterator',
            (i, f) => filter(i, (e, n) => {
                let ok = call(f, [e, n])
                ensure(is(ok, Types.Bool), 'filter_not_bool')
                return ok
            })
    ),
    reversed: fun (
        'function reversed (i: Iterable) -> Iterator',
            i => (function* () {
                let cache = []
                for (let e of i) {
                    cache.push(e)
                }
                for (let e of rev(cache)) {
                    yield e
                }
            })()
    ),
    collect: fun (
        'function collect (i: Iterable) -> List',
            i => list(i)
    )
}
