function require_bool (value) {
    // if (require_bool(may_not_boolean_value)) { ... }
    ensure(is(value, Types.Bool), 'not_bool')
    return value
}


function require_promise (object) {
    // await should_be_promise_or_awaitable
    if (is(object, Types.Promise)) {
        return object
    } else {
        ensure(is(object, Types.Promiser), 'not_awaitable')
        return call_method(null, object, 'promise', [])
    }
}


function when_expr_failed () {
    ensure(false, 'when_expr_failed')
}


let get_data = f (
    'get_data',
    'function get_data (C: ES_Class, k: ES_Key, nf: Bool) -> Object',
        (C, k, nf) => (k in C)? C[k]: (ensure(nf, 'key_error', k), Nil),
    'function get_data (o: ES_Object, k: ES_Key, nf: Bool) -> Object',
        (o, k, nf) => (k in o)? o[k]: (ensure(nf, 'key_error', k), Nil),
    'function get_data (e: Error, k: String, nf: Bool) -> Object',
        (e, k, nf) => {
            if (is(e.data, Types.Hash) && has(k, e.data)) {
                return e.data[k]
            } else {
                ensure(nf, 'key_error', k)
                return Nil
            }
        },
    'function get_data (nil: Nil, k: Any, nf: Bool) -> Object',
        () => Nil,
    'function get_data (M: Module, k: String, nf: Bool) -> Object',
        (M, k, nf) => M.has(k)? M.get(k): (ensure(nf, 'key_error', k), Nil),
    'function get_data (I: Interface, k: String, nf: Bool) -> Object',
        (I, k, nf) => (k == 'Impl')? I.Impl: (ensure(nf, 'key_error', k), Nil),
    'function get_data (C: Class, k: String, nf: Bool) -> Object',
        (C, k, nf) => C.has(k)? C.get(k): (ensure(nf, 'key_error', k), Nil),
    'function get_data (e: Enum, k: String, nf: Bool) -> Object',
        (e, k, nf) => e.has(k)? e.get(k): (ensure(nf, 'key_error', k), Nil),
    'function get_data (g: Getter, k: Any, nf: Bool) -> Object',
        (g, k, nf) => call_method(null, g, 'get', [k, nf]),
    'function get_data (s: Struct, k: String, nf: Bool) -> Object',
        (s, k, nf) => s.has(k)? s.get(k): (ensure(nf, 'key_error', k), Nil),
    'function get_data (l: List, i: Index, nf: Bool) -> Object',
        (l, i, nf) => (i < l.length)? l[i]: (ensure(nf, 'index_error', i), Nil),
    'function get_data (h: Hash, k: String, nf: Bool) -> Object',
        (h, k, nf) => has(k, h)? h[k]: (ensure(nf, 'key_error', k), Nil)
)

let set_data = f (
    'set_data',
    'function set_data (o: ES_Object, k: ES_Key, v: Any) -> Void',
        (o, k, v) => {
            o[k] = v
            return Void
        },
    'function set_data (e: Error, k: String, v: Any) -> Void',
        (e, k, v) => {
            if (!is(e.data, Types.Hash)) {
                e.data = {}
            }
            e.data[k] = v
            return Void
        },
    'function set_data (nil: Nil, k: Any, v: Any) -> Void',
        () => Void,
    'function set_data (s: Setter, k: Any, v: Any) -> Void',
        (s, k, v) => call_method(null, s, 'set', [k, v]),
    'function set_data (s: Struct, k: String, v: Any) -> Void',
        (s, k, v) => {
            s.set(k, v)
            return Void
        },
    'function set_data (l: List, i: Index, v: Any) -> Void',
        (l, i, v) => {
            ensure(i < l.length, 'index_error', i)
            l[i] = v
            return Void
        },
    'function set_data (h: Hash, k: String, v: Any) -> Void',
        (h, k, v) => {
            h[k] = v
            return Void
        }
)

let get_slice = f (
    'get_slice',
    'function get_slice (o: SliceGetter, lo: SliceIndex, hi: SliceIndex) -> Any'
        ,(o, lo, hi) => call_method(null, o, 'slice', [lo, hi]),
    'function get_slice (l: List, lo: SliceIndex, hi: SliceIndex) -> List',
        (l, lo, hi) => {
            lo = (lo === Types.SliceIndexDefault)? 0: lo
            hi = (hi === Types.SliceIndexDefault)? l.length: hi
            ensure(lo <= hi, 'invalid_slice')
            ensure(hi <= l.length, 'slice_index_error', hi)
            return l.slice(lo, hi)
        }
)


let for_loop = f (
    'for_loop',
    'function for_loop (e: Enumerable) -> Iterable',
        e => map(enum_(e), k => {
            if (is(e, Types.Hash)) {
                return { key: k, value: e[k] }
            } else if (is(e, Types.Enum) || is(e, Types.Struct)) {
                return { key: k, value: e.get(k) }
            } else if (is(e, Types.Getter)) {
                let v = call_method(null, e, 'get', [k, false])
                return { key: k, value: v }
            } else {
                return { key: k, value: Nil }
            }
        }),
    'function for_loop (i: Iterable) -> Iterable',
        i => map(iter(i), (e, i) => ({ key: i, value: e }))
)


let iterator_comprehension = fun (
    'function comprehension (v: Function, l: List, f: Function) -> Iterator',
        (v, l, f) => {
            foreach(l, (element, index) => {
                ensure(is(element, Types.Iterable), 'not_iterable', index+1)
            })
            l = l.map(element => iter(element))
            return (function* () {
                for (let values of zip(l, x => x)) {
                    let ok = call(f, values)
                    assert(is(ok, Types.Bool))
                    if (ok) {
                        yield call(v, values)
                    }
                }
            })()
        }
)


let list_comprehension = fun (
    'function list_comprehension (v: Function, l: List, f: Function) -> List',
        (v, l, f, scope) => list(iterator_comprehension[WrapperInfo].raw(scope))
)


let string_format = f (
    'string_format',
    'function string_format (s: String, v: Representable) -> String',
        (s, v) => {
            return s.replace('${}', str(v))
        },
    'function string_format (s: String, t: Struct) -> String',
        (s, t) => {
            return s.replace(/\$\{([^}]+)\}/g, (match, p1) => {
                let key = p1
                ensure(t.schema.has_key(key), 'format_invalid_key', key)
                let value = t.get(key)
                ensure(is(value, Types.Representable), 'not_repr', p1)
                return str(value)
            })
        },
    'function string_format (s: String, h: Hash) -> String',
        (s, h) => {
            return s.replace(/\$\{([^}]+)\}/g, (match, p1) => {
                let key = p1
                let ok = has(key, h)
                ensure(ok, 'format_invalid_key', key)
                let value = h[key]
                ensure(is(value, Types.Representable), 'not_repr', p1)
                return str(value)
            })
        },
    'function string_format (s: String, l: List) -> String',
        (s, l) => {
            let used = 0
            let result = s.replace(/\$\{(\d+)\}/g, (match, p1) => {
                let index = parseInt(p1) - 1
                let ok = (0 <= index && index < l.length)
                ensure(ok, 'format_invalid_index', index)
                used += 1
                let value = l[index]
                ensure(is(value, Types.Representable), 'not_repr', p1)
                return str(value)
            })
            let ok = (used === l.length)
            ensure(ok, 'format_not_all_converted')
            return result
        }
)