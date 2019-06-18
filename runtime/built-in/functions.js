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
        'function custom_error (name: String, msg: String, data: Hash) -> Error'
            ,(name, msg, data) => create_error(msg, name, data)
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
            (i, f) => map(iter(i), (e, n) => call(f, [e, n])),
        'function map (i: Iterable, f: Arity<1>) -> Iterator',
            (i, f) => map(iter(i), e => call(f, [e]))
    ),
    filter: f (
        'filter',
        'function filter (i: Iterable, T: Type) -> Iterator',
            (i, T) => filter(iter(i), e => call(operator_is, [e, T])),
        'function filter (i: Iterable, f: Arity<2>) -> Iterator',
            (i, f) => filter(iter(i), (e, n) => {
                let ok = call(f, [e, n])
                ensure(is(ok, Types.Bool), 'filter_not_bool')
                return ok
            }),
        'function filter (i: Iterable, f: Arity<1>) -> Iterator',
            (i, f) => filter(iter(i), e => {
                let ok = call(f, [e])
                ensure(is(ok, Types.Bool), 'filter_not_bool')
                return ok
            })
    ),
    find: f (
        'find',
        'function find (i: Iterable, T: Type) -> Object',
            (i, T) => {
                let r = find(iter(i), e => call(operator_is, [e, T]))
                return (r !== NotFound)? r: Types.NotFound
            },
        'function find (i: Iterable, f: Arity<2>) -> Object',
            (i, f) => {
                let r = find(iter(i), (e, n) => {
                    let c = call(f, [e, n])
                    ensure(is(c, Types.Bool), 'cond_not_bool')
                    return c
                })
                return (r !== NotFound)? r: Types.NotFound
            },
        'function find (i: Iterable, f: Arity<1>) -> Object',
            (i, f) => {
                let r = find(iter(i), e => {
                    let c = call(f, [e])
                    ensure(is(c, Types.Bool), 'cond_not_bool')
                    return c
                })
                return (r !== NotFound)? r: Types.NotFound
            }
    ),
    fold: f (
        'fold',
        'function fold (i: Iterable, initial: Any, f: Arity<3>) -> Object',
            (i, initial, f) => fold(iter(i), initial, (e, v, n) => {
                return call(f, [e, v, n])
            }),
        'function fold (i: Iterable, initial: Any, f: Arity<2>) -> Object',
            (i, initial, f) => fold(iter(i), initial, (e,v) => {
                return call(f, [e, v])
            })
    ),
    every: f (
        'every',
        'function every (i: Iterable, f: Arity<2>) -> Bool',
            (i, f) => forall(iter(i), (e, i) => {
                let v = call(f, [e, i])
                ensure(is(v, Types.Bool), 'cond_not_bool')
                return v
            }),
        'function every (i: Iterable, f: Arity<1>) -> Bool',
            (i, f) => forall(iter(i), e => {
                let v = call(f, [e])
                ensure(is(v, Types.Bool), 'cond_not_bool')
                return v
            })
    ),
    some: f (
        'some',
        'function some (i: Iterable, f: Arity<2>) -> Bool',
            (i, f) => exists(iter(i), (e, i) => {
                let v = call(f, [e, i])
                ensure(is(v, Types.Bool), 'cond_not_bool')
                return v
            }),
        'function some (i: Iterable, f: Arity<1>) -> Bool',
            (i, f) => exists(iter(i), e => {
                let v = call(f, [e])
                ensure(is(v, Types.Bool), 'cond_not_bool')
                return v
            })
    ),
    join: fun (
        'function join (i: Iterable, sep: String) -> String',
            (i, sep) => {
                let string = ''
                let first = true
                for (let e of i) {
                    ensure(is(e, Types.String), 'element_not_string')
                    if (first) {
                        first = false
                    } else {
                        string += sep
                    }
                    string += e
                }
                return string
            }
    ),
    reversed: fun (
        'function reversed (i: Iterable) -> Iterator',
            i => (function* () {
                let buf = []
                for (let e of iter(i)) {
                    buf.push(e)
                }
                for (let e of rev(buf)) {
                    yield e
                }
            })()
    ),
    collect: fun (
        'function collect (i: Iterable) -> List',
            i => list(iter(i))
    ),
    // List Operations
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
    push: fun (
        'function push (l: List, item: Any) -> Void',
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
    splice: fun (
        'function splice (l: List, i: Index, amount: Size) -> Void',
            (l, i, amount) => {
                ensure(i < l.length, 'index_error', i)
                ensure(i+amount <= l.length, 'invalid_splice', amount)
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
    index_of: fun (
        'function index_of (l: List, f: Arity<1>) -> Maybe<Index>',
            (l, f) => {
                for (let i = 0; i < l.length; i += 1) {
                    let c = call(f, [l[i]])
                    ensure(is(c, Types.Bool), 'cond_not_bool')
                    if (c) {
                        return i
                    }
                }
                return Nil
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
    ),
    // Copy
    copy: f (
        'copy',
        'function copy (l: List) -> List',
            l => copy(l),
        'function copy (h: Hash) -> Hash',
            h => copy(h)
    )
}