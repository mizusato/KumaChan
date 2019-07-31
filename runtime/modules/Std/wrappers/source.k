class Source<T> {
    init (o: Observable) {
        let type_error = custom_error('TypeError', MSG.source_invalid_value())
        let wrapped = observer {
            let unsub = o -> subscribe(Subscriber {
                next: lambda (x) {
                    ensure valid { x is T }
                    push x
                    ...
                    handle error {
                        unless valid {
                            unsub()
                            panic type_error
                        }
                    }
                },
                error: .{ push(.e) },
                complete: .{ push(Complete) }
            })
            return unsub
        }
    }
    get_wrapped () -> Observer {
        return wrapped
    }
    operator obsv (s) {
        return s -> get_wrapped()
    }
}
