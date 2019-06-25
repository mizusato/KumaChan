class Map<K, V> is Getter<K, V>, Setter<K, V> {
    init () {
        static {
            let CreateMap = ES.new(__Map)
        }
        let map = CreateMap()
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
    for_each_entry (f: Arity<2>) -> Void {
        map -> forEach -> .{ (v, k, _) --> f(k, v) }
    }
    operator enum (instance) {
        let keys = []
        let values = []
        instance -> for_each_entry -> lambda(key, value) {
            keys -> append(key)
            values -> append(value)
        }
        return EntryList { ::keys, ::values }
    }
}
