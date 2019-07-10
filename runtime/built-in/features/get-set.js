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
        (_, __, nf) => (ensure(nf, 'get_from_nil'), Nil),
    'function get_data (M: Module, k: String, nf: Bool) -> Object',
        (M, k, nf) => M.has(k)? M.get(k): (ensure(nf, 'key_error', k), Nil),
    'function get_data (C: Class, k: String, nf: Bool) -> Object',
        (C, k, nf) => C.has(k)? C.get(k): (ensure(nf, 'key_error', k), Nil),
    'function get_data (e: Enum, k: String, nf: Bool) -> Object',
        (e, k, nf) => e.has(k)? e.get(k): (ensure(nf, 'key_error', k), Nil),
    'function get_data (g: Getter, k: Any, nf: Bool) -> Object',
        (g, k, nf) => call_method(null, g, 'get', [k, nf]),
    'function get_data (s: Struct, k: String, nf: Bool) -> Object',
        (s, k, nf) => (ensure(!nf, 'struct_nil_flag'), s.get(k)),
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
