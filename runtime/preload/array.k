class Array<T> is Getter<Index, T>, Setter<Index, T> {
    init (source: Iterable) {
        let list = collect(source)
    }
    get (i: Index, nf: Bool) -> T {
        if not (i < length(list)) {
            ensure nil_allowed { nf }
            return Nil
        }
        let element = list[i]
        ensure consistent { element is T }
        return element
        @handle error {
            unless nil_allowed {
                throw custom_error('index ${} out of range' << i)
            }
            unless consistent {
                throw custom_error ('element violated its type constraint')
            }
        }
    }
    set (i: Index, value: T) -> Void {
        ensure index_valid { i < length(list) }
        set list[i] = value
        @handle error {
            unless index_valid {
                throw custom_error('index ${} out of range' << i)
            }
        }
    }
    append (value: T) -> Void {
        list -> append(value)
    }
    prepend (value: T) -> Void {
        list -> prepend(value)
    }
}
