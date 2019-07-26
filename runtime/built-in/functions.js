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


let built_in_functions = {
    // Date & Time
    get_local_unix_time: fun (
        'function get_local_unix_time () -> Int',
            () => Date.now()
    ),
    get_local_timezone_offset: fun (
        'function get_local_timezone_offset () -> Int',
            () => -((new Date()).getTimezoneOffset()) / 60
    ),
    // Number Operations
    abs: fun (
        'function abs (x: Number) -> Number',
            x => Math.abs(x)
    ),
    rand: fun (
        'function rand () -> Number',
            () => Math.random()
    ),
    floor: fun (
        'function floor (x: Number) -> Number',
            x => Math.floor(x)
    ),
    ceil: fun (
        'function ceil (x: Number) -> Number',
            x => Math.ceil(x)
    ),
    round: fun (
        'function round (x: Number) -> Number',
            x => Math.round(x)
    ),
    // String Operations
    utf8_size: fun (
        'function utf8_size (s: String) -> Size',
            s => fold(map(s, c => c.codePointAt(0)), 0, (code_point, sum) => {
                if ((code_point & 0xFFFFFF80) == 0) {
                    return sum + 1
                } else if ((code_point & 0xFFFFF800) == 0) {
                    return sum + 2
                } else if ((code_point & 0xFFFF0000) == 0) {
                    return sum + 3
                } else if ((code_point & 0xFFE00000) == 0) {
                    return sum + 4
                } else {
                    assert(false)
                }
            })
    ),
    ord: fun (
        'function ord (c: Char) -> Index',
            c => c.codePointAt(0)
    ),
    chr: fun (
        'function chr (i: Index) -> Char',
            i => {
                try {
                    return String.fromCodePoint(i)
                } catch (e) {
                    if (e instanceof RangeError) {
                        ensure(false, 'invalid_code_point', i.toString(16))
                    } else {
                        throw e
                    }
                }
            }
    ),
    to_lower_case: fun (
        'function to_lower_case (s: String) -> String',
            s => s.toLowerCase()
    ),
    to_upper_case: fun (
        'function to_upper_case (s: String) -> String',
            s => s.toUpperCase()
    ),
    match: fun (
        'function match (s: String, regexp: String) -> Maybe<List>',
            (s, regexp) => {
                let r = null
                try {
                    r = new RegExp(regexp, 'su')
                } catch (e) {
                    ensure(!(e instanceof SyntaxError), 'regexp_invalid')
                    throw e
                }
                let m = s.match(r)
                return (m != null)? list(m): Nil
            }
    ),
    match_all: fun (
        'function match_all (s: String, regexp: String) -> Iterator',
        (s, regexp) => (function* () {
            let r = null
            try {
                r = new RegExp(regexp, 'sgu')
            } catch (e) {
                ensure(!(e instanceof SyntaxError), 'regexp_invalid')
                throw e
            }
            let m = null
            while ((m = r.exec(s)) !== null) {
                yield list(m)
            }
        })()
    ),
    replace: fun (
        'function replace (s: String, regex: String, f: Function) -> String',
            (s, regexp, f) => {
                let r = null
                try {
                    r = new RegExp(regexp, 'sgu')
                } catch (e) {
                    ensure(!(e instanceof SyntaxError), 'regexp_invalid')
                    throw e
                }
                return s.replace(r, (...args) => {
                    let n = args.length - 3
                    let t = call(f, args.slice(0, args.length-2))
                    ensure(is(t, Types.String), 'replace_not_string')
                    return t
                })
            }
    ),
    // Operations on Iterables and Observables
    subscribe: f (
        'subscribe',
        'function subscribe (o: Observable, s: Subscriber) -> Arity<0>',
            (o, s) => o.subscribe(s),
        'function subscribe (o: Observable, f: Arity<1>) -> Arity<0>',
            (o, f) => o.subscribe(new_struct(Types.Subscriber, {
                next: f
            })),
        'function subscribe (o: Observable, f: Arity<0>) -> Arity<0>',
            (o, f) => o.subscribe(new_struct(Types.Subscriber, {
                next: fun (
                    'function callback (_: Any) -> Object',
                        _ => call(f, [])
                )
            }))
    ),
    seq: f (
        'seq',
        'function seq (n: Size) -> Iterator',
            n => count(n),
        'function seq (start: Index, amount: Size) -> Iterator',
            (start, amount) => map(count(amount), i => start + i)
    ),
    repeat: fun (
        'function repeat (object: Any, n: Size) -> Iterator',
            (object, n) => (function* () {
                for (let i = 0; i < n; i++) {
                    yield object
                }
            })()
    ),
    concat: f (
        'concat',
        'function concat (o1: Observable, o2: Observable) -> Observer',
            (o1, o2) => observer(push => {
                let unsub = obsv(o1).subscribe(subs({
                    next: x => push(x),
                    error: e => push(e),
                    complete: () => {
                        unsub = obsv(o2).subscribe(subs({
                            next: x => push(x),
                            error: e => push(e),
                            complete: () => push(Types.Complete)
                        }))
                    }
                }))
                return () => unsub()
            }),
        'function concat (i1: Iterable, i2: Iterable) -> Iterator',
            (i1, i2) => (function* () {
                for (let e of iter(i1)) {
                    yield e
                }
                for (let e of iter(i2)) {
                    yield e
                }
            })()
    ),
    range: f (
        'range',
        'function range (begin: Index, end: Index) -> Iterator',
            (begin, end) => {
                ensure(begin <= end, 'invalid_range', begin, end)
                return (function* () {
                    for (let i = begin; i < end; i += 1) {
                        yield i
                    }
                })()
            },
        'function range (begin: Index, end: Index, step: Size) -> Iterator',
            (begin, end, step) => {
                ensure(begin <= end, 'invalid_range', begin, end)
                return (function* () {
                    for (let i = begin; i < end; i += step) {
                        yield i
                    }
                })()
            }
    ),
    map: f (
        'map',
        'function map (o: Observable, f: Arity<2>) -> Observer',
            (o, f) => observer(push => {
                let i = 0
                return obsv(o).subscribe(subs({
                    next: x => {
                        push(call(f, [x, i]))
                        i += 1
                    },
                    error: e => push(e),
                    complete: () => push(Types.Complete)
                }))
            }),
        'function map (o: Observable, f: Arity<1>) -> Observer',
            (o, f) => observer(push => obsv(o).subscribe(subs({
                next: x => push(call(f, [x])),
                error: e => push(e),
                complete: () => push(Types.Complete)
            }))),
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
    count: fun (
        'function count (i: Iterable) -> Size',
            i => fold(iter(i), 0, (_, v) => v+1)
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
            })(),
        'function reversed (l: List) -> Iterator',
            l => rev(l)
    ),
    flat: fun (
        'function flat (i: Iterable) -> Iterator',
            i => (function* () {
                for (let e of iter(i)) {
                    ensure(is(e, Types.Iterable), 'element_not_iterable')
                    for (let ee of iter(e)) {
                        yield ee
                    }
                }
            })()
    ),
    collect: fun (
        'function collect (i: Iterable) -> List',
            i => list(iter(i))
    ),
    // Enumerable Object Operations
    get_keys: fun (
        'function get_keys (e: Enumerable) -> List',
            e => copy(enum_(e).get('keys'))
    ),
    get_values: fun (
        'function get_values (e: Enumerable) -> List',
            e => copy(enum_(e).get('values'))
    ),
    get_entries: fun (
        'function get_entries (e: Enumerable) -> List',
            e => list((function* () {
                let entry_list = enum_(e)
                let keys = entry_list.get('keys')
                let values = entry_list.get('values')
                assert(keys.length == values.length)
                let L = keys.length
                for (let i = 0; i < L; i += 1) {
                    yield { key: keys[i], value: values[i] }
                }
            })())
    ),
    map_entry: fun (
        'function map_entry (e: Enumerable, f: Arity<2>) -> Iterator',
            (e, f) => (function* () {
                let entry_list = enum_(e)
                let keys = entry_list.get('keys')
                let values = entry_list.get('values')
                assert(keys.length == values.length)
                let L = keys.length
                for (let i = 0; i < L; i += 1) {
                    yield call(f, [keys[i], values[i]])
                }
            })()
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
                l.push(l[l.length-1])
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
    // OO
    get_class: fun (
        'function get_class (i: Instance) -> Class',
            i => i.class_
    ),
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
            msg => new CustomError(msg),
        'function custom_error (name: String, msg: String) -> Error',
            (name, msg) => new CustomError(msg, name),
        'function custom_error (name: String, msg: String, data: Hash) -> Error'
            ,(name, msg, data) => new CustomError(msg, name, data)
    ),
    // Async
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
                    push(Types.Complete)
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
}
