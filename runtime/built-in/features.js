function require_bool (value) {
    // if (require_bool(may_not_boolean_value)) { ... }
    ensure(is(value, Types.Bool), 'not_bool')
    return value
}


function require_promise (object) {
    // await should_be_promise_or_future
    // TODO: Types.Future (Typed Promise)
    ensure(is(object, Types.Promise), 'not_promise')
    return object
}


let get_data = f (
    'get_data',
    'function get_data (C: ES_Class, k: ES_Key, nf: Bool) -> Object',
        (C, k, nf) => (k in C)? C[k]: (ensure(nf, 'key_error', k), Nil),
    'function get_data (o: ES_Object, k: ES_Key, nf: Bool) -> Object',
        (o, k, nf) => (k in o)? o[k]: (ensure(nf, 'key_error', k), Nil),
    'function get_data (nil: Nil, k: Any, nf: Bool) -> Object',
        () => Nil,
    'function get_data (e: Error, k: String, nf: Bool) -> Object',
        (e, k, nf) => {
            if (is(e.data, Types.Hash) && has(k, e.data)) {
                return e.data[k]
            } else {
                ensure(nf, 'key_error', k)
                return Nil
            }
        },
    'function get_data (s: Structure, k: String, nf: Bool) -> Object',
        s => (ensure(!nf, 'struct_nil_flag'), s.get(k)),
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
    'function set_data (nil: Nil, k: Any, v: Any) -> Void',
        () => Void,
    'function set_data (e: Error, k: String, v: Any) -> Void',
        (e, k, v) => {
            if (!is(e.data, Types.Hash)) {
                e.data = {}
            }
            e.data[k] = v
            return Void
        },
    'function set_data (s: Structure, k: String, v: Any) -> Void',
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


let for_loop = f (
    'for_loop',
    'function for_loop (h: Hash) -> Iterable',
        h => mapkv(h, (k, v) => ({ key: k, value: v })),
    'function for_loop (i: Iterable) -> Iterable',
        i => map(i, (e, i) => ({ key: i, value: e }))
)


let iterator_comprehension = fun (
    'function comprehension (v: Function, l: List, f: Function) -> Iterator',
        (v, l, f) => {
            foreach(l, (element, index) => {
                ensure(is(element, Types.Iterable), 'not_iterable', index+1)
            })
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
    'function string_format (s: String, v: Any) -> String',
        (s, v) => {
            return s.replace('${}', call(operators['str'], [v]))
        },
    'function string_format (s: String, h: Hash) -> String',
        (s, h) => {
            return s.replace(/\$\{([^}]+)\}/g, (match, p1) => {
                let key = p1
                let ok = has(key, h)
                ensure(ok, 'format_invalid_key', key)
                return call(operators['str'], [h[key]])
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
                return call(operators['str'], [l[index]])
            })
            let ok = (used === l.length)
            ensure(ok, 'format_not_all_converted')
            return result
        }
)
