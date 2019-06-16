class Array<T> is Getter<Index, T>, Setter<Index, T>, SliceGetter {
    init (source: Iterable) {
        let list = []
        for element[i] in source {
            ensure valid(i) { element is T }
            list -> append(element)
        }
        @handle error {
            unless valid(i) {
                panic MSG.array_invalid_init(i)
            }
        }
    }
    get (i: Index, nf: Bool) -> Maybe<T> {
        if not (i < len(list)) {
            ensure nil_allowed { nf }
            return Nil
        }
        let element = list[i]
        ensure consistent { element is T }
        return element
        @handle error {
            unless nil_allowed {
                panic MSG.array_out_of_range(i, len(list))
            }
            unless consistent {
                panic MSG.array_inconsistent(i)
            }
        }
    }
    set (i: Index, value: T) -> Void {
        ensure index_valid { i < len(list) }
        set list[i] = value
        @handle error {
            unless index_valid {
                panic MSG.array_out_of_range(i, len(list))
            }
        }
    }
    slice (lo: SliceIndex, hi: SliceIndex) -> Object {
        return Array<T>(list[lo:hi])
    }
    length () -> Size {
        return len(list)
    }
    append (value: T) -> Void {
        list -> append(value)
    }
    prepend (value: T) -> Void {
        list -> prepend(value)
    }
    operator len (array) {
        return array -> length()
    }
    operator iter (array) {
        return iterator {
            for i in count(len(array)) {
                yield array[i]
            }
        }
    }
    operator + (a1, a2) {
        return Array<T>(iter(a1) + iter(a2))
    }
}
