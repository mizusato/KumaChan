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
            ensure(!closed, 'push_observer_closed')
            if (is(object, Types.Complete)) {
                close()
            } else if (is(object, Types.Error)) {
                closed = true
                ensure(error !== Nil, 'push_no_error_handler')
                call(error, [object])
            } else {
                call(next, [object])
            }
            return object
        }
        let unsub = call(init, [])
        return fun (
            'function cancel_subscription () -> Bool',
                () => {
                    if (!is(unsub, Types.Arity.inflate(0))) {
                        return false
                    }
                    ensure(!closed, 'redundant_unsub')
                    call(unsub, [])
                    close()
                    return true
                }
        )
    }
}

Types.Observer = $(x => x instanceof Observer)
Types.Observable = Uni(Types.Observer, Types.Operand.inflate('obsv'))

let create_observer = f => new Observer(f)


Types.Complete = create_value('Complete')

Types.Subscriber = create_schema('Subscriber', {
    next: Types.Arity.inflate(1),
    error: Types.Maybe.inflate(Types.Arity.inflate(1)),
    complete: Types.Maybe.inflate(Types.Arity.inflate(0))
}, {
    error: Nil,
    complete: Nil
}, [], {})
