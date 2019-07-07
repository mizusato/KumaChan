class Set<T> {
    init () {
        static {
            let CreateSet = ES.new(__Set)
        }
        let set = CreateSet()
    }
    has (element: T) -> Bool {
        return set -> has(element)
    }
    add (element: T) -> Void {
        set -> add(element)
    }
    remove (element: T) -> Void {
        ensure element_exists { set -> has(element) }
        set -> delete(element)
        @handle error {
            unless element_exists {
                panic MSG.set_invalid_remove()
            }
        }
    }
    clear () -> Void {
        set -> clear()
    }
    size () -> Size {
        return set.size
    }
    get_element_list () -> List {
        return [ element, for element in set ]
    }
    operator enum (s) {
        let elements = s -> get_element_list()
        return EntryList {
            keys: elements,
            values: elements
        }
    }
}
