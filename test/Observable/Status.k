promise {
    enum Status { Running, Error, Complete }
    let make_status = lambda (s: Status) {
        return observer {
            switch {
            case s is Status.Running {
                push(1)
            }
            case s is Status.Error {
                push(custom_error('Oops!'))
            }
            case s is Status.Complete {
                push(Complete)
            }
            }
        }
    }
    let sub = lambda (s: Status) {
        return create_promise << lambda (resolve, _) {
            make_status(s) -> subscribe -> Subscriber {
                next: .{ _ --> resolve(Status.Running) },
                error: .{ _ --> resolve(Status.Error) },
                complete: .{ resolve(Status.Complete) }
            }
        }
    }
    for S in Status {
        await t = sub(S)
        assert t is S
    }
}
