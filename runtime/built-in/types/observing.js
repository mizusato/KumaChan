class Observer {
    constructor (init) {
        assert(is(init, Types.Arity.inflate(0)))
        this.init = init
        Object.freeze(this)
    }
    subscribe (sub) {
        assert(is(sub, Types.Subscriber))
        let old_context = this.init[WrapperInfo].context
        let new_context = new Scope(old_context)
        new_context.define_push(inject_desc(x => push(x), 'operator_push'))
        let init = embrace_in_context(this.init, new_context)
        let complete = sub.get('complete')
        let error = sub.get('error')
        let next = sub.get('next')
        let closed = false
        let close = () => {
            closed = true
            if (complete !== Nil) {
                call(complete, [])
            }
        }
        let push = object => {
            // ensure(!closed, 'push_observer_closed')
            if (closed) { return object }
            if (is(object, Complete)) {
                close()
            } else if (is(object, Types.Error)) {
                closed = true
                // ensure(error !== Nil, 'push_no_error_handler')
                if (error == Nil) {
                    let { name, message } = object
                    crash(`UnhandledObserverError: ${name}: ${message}`)
                }
                call(error, [object])
            } else {
                call(next, [object])
            }
            return object
        }
        let unsub = call(init, [])
        return fun (
            'function cancel_subscription () -> Void',
                () => {
                    if (!is(unsub, Types.Arity.inflate(0))) {
                        return Void
                    }
                    // ensure(!closed, 'redundant_unsub')
                    if (closed) { return Void }
                    call(unsub, [])
                    close()
                    return Void
                }
        )
    }
}

let create_observer = f => new Observer(f)

Types.Observer = $(x => x instanceof Observer)
Types.Observable = Uni(Types.Observer, Types.Operand.inflate('obsv'))

let Complete = create_value('Complete')
Types.Complete = Complete
Types.Subscriber = create_schema('Subscriber', {
    next: Types.Arity.inflate(1),
    error: Types.Maybe.inflate(Types.Arity.inflate(1)),
    complete: Types.Maybe.inflate(Types.Arity.inflate(0))
}, {
    error: Nil,
    complete: Nil
}, [], {})


let observer = f => create_observer (
    fun (
        'function observer () -> Object',
            scope => {
                let p = scope.push.bind(scope)
                let ret = f(object => call(call(p, []), [object]))
                if (is(ret, ES.Function)) {
                    return fun (
                        'function unsubscribe () -> Void',
                            _ => {
                                ret()
                                return Void
                            }
                    )
                } else {
                    return Void
                }
            }
    )
)


let subs = hooks => {
    let { next, error, complete } = hooks
    return new_struct(Types.Subscriber, {
        next: fun (
            'function next (x: Any) -> Void',
                x => (next(x), Void)
        ),
        error: fun (
            'function error (e: Error) -> Void',
                e => (error(e), Void)
        ),
        complete: fun (
            'function complete () -> Void',
                () => (complete(), Void)
        )
    })
}
