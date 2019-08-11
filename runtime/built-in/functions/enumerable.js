pour(built_in_functions, {
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
    has: fun (
        'function has (h: Hash, k: String) -> Bool',
            (h, k) => has(k, h)
    ),
    delete: fun (
        'function delete (h: Hash, k: String) -> Void',
            (h, k) => {
                ensure(has(k, h), 'hash_invalid_delete')
                delete h[k]
                return Void
            }
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
    )
})
