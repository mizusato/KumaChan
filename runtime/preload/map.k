class Map<K, V> is Getter<K, V>, Setter<K, V> {
    init () {
        static {
            let ES_Map = es.new(__Map)
        }
        let map = ES_Map()
    }
    has (key: K) -> Bool {
        return map->has(key)
    }
    get (key: K, nf: Bool) -> Maybe<V> {
        if not map->has(key) {
            ensure nil_allowed { nf }
            return Nil
        }
        let value = map->get(key)
        ensure consistent { value is V }
        return value
        @handle error {
            unless nil_allowed {
                panic MSG.map_value_not_found()
            }
            unless consistent {
                panic MSG.map_inconsistent()
            }
        }
    }
    set (key: K, value: V) -> Void {
        map->set(key, value)
    }
    remove (key: K) -> Void {
        ensure key_exists { map->has(key) }
        map->delete(key)
        @handle error {
            unless key_exists {
                panic MSG.map_invalid_remove()
            }
        }
    }
    clear () -> Void {
        map->clear()
    }
    size () -> Size {
        return map.size
    }
    entries () -> List {
        return [ { key: e[0], value: e[1] }, for e in map->entries() ]
    }
    map_entry (f: Arity<2>) -> Iterator {
        return .[ f(e[0], e[1]), for e in map->entries() ]
    }
    operator enum (instance) {
        return instance -> map_entry -> .{ (k, v) --> k } -> collect
    }
}
