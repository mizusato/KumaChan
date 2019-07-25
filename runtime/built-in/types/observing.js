class Observer {
    constructor (init) {
        assert(is(init, Types.Arity.inflate(0)))
        this.init = init
        Object.frezee(this)
    }
    subscribe (sub) {
        assert(is(sub, Types.Subscriber))
        let closed = false
        let old_context = this.init[WrapperInfo].context
        let new_context = new Scope(old_context)
        let complete = sub.get('complete')
        let error = sub.get('error')
        let next = sub.get('next')
        let push = object => {
            ensure(!closed, 'observer_closed')
            if (is(object, Types.Complete)) {
                closed = true
                if (complete !== Nil) {
                    call(complete, [])
                }
            } else if (is(object, Types.Error)) {
                closed = true
                ensure(error !== Nil, 'no_error_handler')
                call(error, [object])
            } else {
                call(next, [object])
            }
            return object
        }
        new_context.define_push(inject_desc(push, 'operator_push'))
        let unsub = call(embrace_in_context(this.init, new_context), [])
        return fun (
            'function cancel_subscription () -> Void',
                () => {
                    if (is(unsub, Types.Arity.inflate(0))) {
                        call(unsub, [])
                    }
                    closed = true
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
