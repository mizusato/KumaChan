let built_in_functions = {
    // Ouput
    print: f (
        'print',
        'function print (p: Bool) -> Void',
            x => (console.log(x.toString()), Void),
        'function print (x: Number) -> Void',
            x => (console.log(x.toString()), Void),
        'function print (s: String) -> Void',
            s => (console.log(s), Void)
    ),
    // Error Handling
    custom_error: f (
        'custom_error',
        'function custom_error (msg: String) -> Error',
            msg => create_error(msg),
        'function custom_error (name: String, msg: String) -> Error',
            (name, msg) => create_error(msg, name),
        'function custom_error (name: String, msg: String, data: Hash) -> Error',
            (name, msg, data) => create_error(msg, name, data)
    ),
    // Singleton Value Creator
    custom_value: fun (
        'function custom_value (name: String) -> Singleton',
            name => create_value(name)
    ),
    // Async
    postpone: f (
        'postpone',
        'function postpone (time: Size) -> Promise',
            time => new Promise(resolve => {
                setTimeout(() => resolve(Nil), time)
            }),
        'function postpone (time: Size, callback: Arity<0>) -> Void',
            (time, callback) => {
                let frame = get_top_frame()
                let pos = ''
                if (frame !== null) {
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
    // Iterator Operations
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
        'function map (i: Iterable, f: Arity<2>) -> Iterator',
            (i, f) => map(i, (e, n) => call(f, [e, n])),
        'function map (i: Iterable, f: Arity<1>) -> Iterator',
            (i, f) => map(i, e => call(f, [e]))
    ),
    filter: f (
        'filter',
        'function filter (i: Iterable, T: Type) -> Iterator',
            (i, T) => filter(i, e => call(operator_is, [e, T])),
        'function filter (i: Iterable, f: Arity<2>) -> Iterator',
            (i, f) => filter(i, (e, n) => {
                let ok = call(f, [e, n])
                ensure(is(ok, Types.Bool), 'filter_not_bool')
                return ok
            }),
        'function filter (i: Iterable, f: Arity<1>) -> Iterator',
            (i, f) => filter(i, e => {
                let ok = call(f, [e])
                ensure(is(ok, Types.Bool), 'filter_not_bool')
                return ok
            })
    ),
    find: f (
        'find',
        'function find (i: Iterable, T: Type) -> Object',
            (i, T) => {
                let r = find(i, e => call(operator_is, [e, T]))
                return (r !== NotFound)? r: Types.NotFound
            },
        'function find (i: Iterable, f: Arity<2>) -> Object',
            (i, f) => {
                let r = find(i, (e, n) => call(f, [e, n]))
                return (r !== NotFound)? r: Types.NotFound
            },
        'function find (i: Iterable, f: Arity<1>) -> Object',
            (i, f) => {
                let r = find(i, e => call(f, [e]))
                return (r !== NotFound)? r: Types.NotFound
            }
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
    ),
    // List Operations
    length: f (
        'length',
        'function length (l: List) -> Size',
            l => l.length,
        'function length (s: String) -> Size',
            s => s.length
    ),
    first: fun (
        'function first (l: List) -> Object',
            l => (ensure(l.length > 0, 'empty_list'), l[0])
    ),
    last: fun (
        'function last (l: List) -> Object',
            l => (ensure(l.length > 0, 'empty_list'), l[l.length-1])
    ),
    prepend: fun (
        'function prepend (l: List, item: Any) -> Void',
            (l, item) => {
                l.unshift(item)
                return Void
            }
    ),
    append: fun (
        'function append (l: List, item: Any) -> Void',
            (l, item) => {
                l.push(item)
                return Void
            }
    ),
    shift: fun (
        'function shift (l: List) -> Void',
            l => {
                ensure(l.length > 0, 'empty_list')
                l.shift()
                return Void
            }
    ),
    pop: fun (
        'function pop (l: List) -> Void',
            l => {
                ensure(l.length > 0, 'empty_list')
                l.pop()
                return Void
            }
    ),
    slice: fun (
        'function slice (l: List, low: Index, high: Index) -> List',
            (l, low, high) => {
                ensure(low <= high, 'invalid_slice', low, high)
                ensure(high < l.length, 'index_error', high)
                return l.slice(low, high)
            }
    ),
    splice: fun (
        'function splice (l: List, i: Index, amount: Size) -> Void',
            (l, i, amount) => {
                ensure(i < l.length, 'index_error', i)
                ensure(i+amount < l.length, 'invalid_splice', amount)
                l.splice(i, amount)
                return Void
            }
    ),
    insert: fun (
        'function insert (l: List, i: Index, item: Any) -> Void',
            (l, i, item) => {
                ensure(i < l.length, 'index_error', i)
                if (i == l.length-1) {
                    l.push(item)
                    return Void
                }
                let target = i+1
                l.push(Nil)
                for (let j=l.length-1; j>target; j--) {
                    l[j] = l[j-1]
                }
                l[target] = item
                return Void
            }
    ),
    // Hash Operations
    get_keys: fun (
        'function get_keys (h: Hash) -> List',
            h => Object.keys(h)
    ),
    get_values: fun (
        'function get_values (h: Hash) -> List',
            h => list(map(Object.keys(h), k => h[k]))
    ),
    get_entries: fun (
        'function get_entries (h: Hash) -> List',
            h => list(map(Object.keys(h), k => ({ key: k, value: h[k] })))
    ),
    map_key: f (
        'map_key',
        'function map_key (h: Hash, f: Arity<1>) -> Hash',
            (h, f) => mapkey(h, k => call(f, [k])),
        'function map_key (h: Hash, f: Arity<2>) -> Hash',
            (h, f) => mapkey(h, (k, v) => call(f, [k, v]))
    ),
    map_value: f (
        'map_value',
        'function map_value (h: Hash, f: Arity<1>) -> Hash',
            (h, f) => mapval(h, v => call(f, [v])),
        'function map_value (h: Hash, f: Arity<2>) -> Hash',
            (h, f) => mapval(h, (v, k) => call(f, [v, k]))
    ),
    map_entry: fun (
        'function map_entry (h: Hash, f: Arity<2>) -> Iterator',
            (h, f) => mapkv(h, (k, v) => call(f, [k, v]))
    )
}
