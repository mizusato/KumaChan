class Stream<T> {
    init (i: Iterable) {
        let type_error = custom_error('TypeError', MSG.stream_invalid_value())
        let wrapped = iterator {
            for e in iter(i) {
                if e is not T {
                    panic type_error
                }
                yield e
            }
        }
    }
    get_wrapped () -> Iterator {
        return wrapped
    }
    operator iter (s) {
        return s -> get_wrapped()
    }
}


class AsyncStream<T> {
    init (ai: AsyncIterable) {
        let type_error = custom_error('TypeError', MSG.stream_invalid_value())
        let wrapped = async iterator {
            await e in iter(ai) {
                if e is not T {
                    panic type_error
                }
                yield e
            }
        }
    }
    get_wrapped () -> AsyncIterator {
        return wrapped
    }
    operator async_iter (s) {
        return s -> get_wrapped()
    }
}
