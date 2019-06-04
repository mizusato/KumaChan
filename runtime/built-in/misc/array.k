class Array<T> {
    init (source: Iterable) {
        let list = collect(source)
    }
    get (i: Index) -> T {
        ensure index_valid { i < length(list) }
        let element = list[i]
        ensure consistent { element is T }
        return element
        @handle error {
            unless index_valid {
                throw custom_error('invalid array index ${}' << i)
            }
            unless consistent {
                throw custom_error ('element violated type constraint')
            }
        }
    }
    set (i: Index, value: Any) -> Void {
        ensure index_valid { i < length(list) }
        ensure value_valid { value is T }
        set list[i] = T
        @handle error {
            unless index_valid {
                throw custom_error('invalid array index ${}' << i)
            }
            unless value_valid {
                throw custom_error('invalid element value')
            }
        }
    }
    append (value: T) -> Void {
        list -> append(value)
    }
}
