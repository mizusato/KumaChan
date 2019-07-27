pour(built_in_functions, {
    subscribe: f (
        'subscribe',
        'function subscribe (o: Observable, s: Subscriber) -> Arity<0>',
            (o, s) => obsv(o).subscribe(s),
        'function subscribe (o: Observable, f: Arity<1>) -> Arity<0>',
            (o, f) => obsv(o).subscribe(new_struct(Types.Subscriber, {
                next: f
            })),
        'function subscribe (o: Observable, f: Arity<0>) -> Arity<0>',
            (o, f) => obsv(o).subscribe(new_struct(Types.Subscriber, {
                next: fun (
                    'function callback (_: Any) -> Object',
                        _ => call(f, [])
                )
            }))
    ),
    iter2obsv: fun (
        'function iter2obsv (i: Iterable) -> Observer',
            i => observer(push => {
                for (let e of iter(i)) {
                    push(e)
                }
            }),
    ),
    seq: f (
        'seq',
        'function seq (n: Size) -> Iterator',
            n => count(n),
        'function seq (start: Index, amount: Size) -> Iterator',
            (start, amount) => map(count(amount), i => start + i)
    ),
    repeat: fun (
        'function repeat (object: Any, n: Size) -> Iterator',
            (object, n) => (function* () {
                for (let i = 0; i < n; i++) {
                    yield object
                }
            })()
    ),
    concat: f (
        'concat',
        'function concat (o1: Observable, o2: Observable) -> Observer',
            (o1, o2) => observer(push => {
                let unsub = obsv(o1).subscribe(subs({
                    next: x => push(x),
                    error: e => push(e),
                    complete: () => {
                        unsub = obsv(o2).subscribe(subs({
                            next: x => push(x),
                            error: e => push(e),
                            complete: () => push(Complete)
                        }))
                    }
                }))
                return () => unsub()
            }),
        'function concat (i1: Iterable, i2: Iterable) -> Iterator',
            (i1, i2) => (function* () {
                for (let e of iter(i1)) {
                    yield e
                }
                for (let e of iter(i2)) {
                    yield e
                }
            })()
    ),
    merge: fun (
        'function merge (o1: Observable, o2: Observable) -> Observer',
            (o1, o2) => observer(push => {
                let f1 = false
                let f2 = false
                let unsub1 = obsv(o1).subscribe(subs({
                    next: x => push(x),
                    error: e => {
                        unsub2()
                        push(e)
                    },
                    complete: () => {
                        f1 = true
                        if (f2) {
                            push(Complete)
                        }
                    }
                }))
                let unsub2 = obsv(o2).subscribe(subs({
                    next: x => push(x),
                    error: e => {
                        unsub1()
                        push(e)
                    },
                    complete: () => {
                        f2 = true
                        if (f1) {
                            push(Complete)
                        }
                    }
                }))
                return () => { unsub1(); unsub2() }
            })
    ),
    range: f (
        'range',
        'function range (begin: Index, end: Index) -> Iterator',
            (begin, end) => {
                ensure(begin <= end, 'invalid_range', begin, end)
                return (function* () {
                    for (let i = begin; i < end; i += 1) {
                        yield i
                    }
                })()
            },
        'function range (begin: Index, end: Index, step: Size) -> Iterator',
            (begin, end, step) => {
                ensure(begin <= end, 'invalid_range', begin, end)
                return (function* () {
                    for (let i = begin; i < end; i += step) {
                        yield i
                    }
                })()
            }
    ),
    map: f (
        'map',
        'function map (o: Observable, f: Arity<2>) -> Observer',
            (o, f) => observer(push => {
                let i = 0
                return obsv(o).subscribe(subs({
                    next: x => {
                        push(call(f, [x, i]))
                        i += 1
                    },
                    error: e => push(e),
                    complete: () => push(Complete)
                }))
            }),
        'function map (o: Observable, f: Arity<1>) -> Observer',
            (o, f) => observer(push => obsv(o).subscribe(subs({
                next: x => push(call(f, [x])),
                error: e => push(e),
                complete: () => push(Complete)
            }))),
        'function map (i: Iterable, f: Arity<2>) -> Iterator',
            (i, f) => map(iter(i), (e, n) => call(f, [e, n])),
        'function map (i: Iterable, f: Arity<1>) -> Iterator',
            (i, f) => map(iter(i), e => call(f, [e]))
    ),
    filter: f (
        'filter',
        'function filter (o: Observable, f: Arity<2>) -> Observer',
            (o, f) => observer(push => {
                let i = 0
                return obsv(o).subscribe(subs({
                    next: x => {
                        let ok = call(f, [x, i])
                        ensure(is(ok, Types.Bool), 'filter_not_bool')
                        if (ok) {
                            push(x)
                        }
                        i += 1
                    },
                    error: e => push(e),
                    complete: () => push(Complete)
                }))
            }),
        'function filter (o: Observable, f: Arity<1>) -> Observer',
            (o, f) => observer(push => obsv(o).subscribe(subs({
                next: x => {
                    let ok = call(f, [x])
                    ensure(is(ok, Types.Bool), 'filter_not_bool')
                    if (ok) {
                        push(x)
                    }
                },
                error: e => push(e),
                complete: () => push(Complete)
            }))),
        'function filter (i: Iterable, f: Arity<2>) -> Iterator',
            (i, f) => filter(iter(i), (e, n) => {
                let ok = call(f, [e, n])
                ensure(is(ok, Types.Bool), 'filter_not_bool')
                return ok
            }),
        'function filter (i: Iterable, f: Arity<1>) -> Iterator',
            (i, f) => filter(iter(i), e => {
                let ok = call(f, [e])
                ensure(is(ok, Types.Bool), 'filter_not_bool')
                return ok
            })
    ),
    take: f (
        'take',
        'function take (o: Observable, f: Arity<2>) -> Observer',
            (o, f) => observer(push => {
                let i = 0
                let unsub = obsv(o).subscribe(subs({
                    next: x => {
                        let ok = call(f, [x, i])
                        ensure(is(ok, Types.Bool), 'filter_not_bool')
                        if (ok) {
                            push(x)
                        } else {
                            unsub()
                            push(Complete)
                        }
                        i += 1
                    },
                    error: e => push(e),
                    complete: () => push(Complete)
                }))
                return unsub
            }),
        'function take (o: Observable, f: Arity<1>) -> Observer',
            (o, f) => observer(push => {
                let unsub = obsv(o).subscribe(subs({
                    next: x => {
                        let ok = call(f, [x])
                        ensure(is(ok, Types.Bool), 'filter_not_bool')
                        if (ok) {
                            push(x)
                        } else {
                            unsub()
                            push(Complete)
                        }
                    },
                    error: e => push(e),
                    complete: () => push(Complete)
                }))
                return unsub
            }),
        'function take (i: Iterable, f: Arity<2>) -> Iterator',
            (i, f) => (function* () {
                let n = 0
                for (let e of iter(i)) {
                    let ok = call(f, [e, n])
                    ensure(is(ok, Types.Bool), 'filter_not_bool')
                    if (ok) {
                        yield e
                        n += 1
                    } else {
                        break
                    }
                }
            })(),
        'function take (i: Iterable, f: Arity<1>) -> Iterator',
            (i, f) => (function* () {
                for (let e of iter(i)) {
                    let ok = call(f, [e])
                    ensure(is(ok, Types.Bool), 'filter_not_bool')
                    if (ok) {
                        yield e
                    } else {
                        break
                    }
                }
            })()
    ),
    find: f (
        'find',
        'function find (i: Iterable, T: Type) -> Object',
            (i, T) => {
                let r = find(iter(i), e => call(operator_is, [e, T]))
                return (r !== NotFound)? r: Types.NotFound
            },
        'function find (i: Iterable, f: Arity<2>) -> Object',
            (i, f) => {
                let r = find(iter(i), (e, n) => {
                    let c = call(f, [e, n])
                    ensure(is(c, Types.Bool), 'cond_not_bool')
                    return c
                })
                return (r !== NotFound)? r: Types.NotFound
            },
        'function find (i: Iterable, f: Arity<1>) -> Object',
            (i, f) => {
                let r = find(iter(i), e => {
                    let c = call(f, [e])
                    ensure(is(c, Types.Bool), 'cond_not_bool')
                    return c
                })
                return (r !== NotFound)? r: Types.NotFound
            }
    ),
    count: fun (
        'function count (i: Iterable) -> Size',
            i => fold(iter(i), 0, (_, v) => v+1)
    ),
    scan: f (
        'scan',
        'function scan (o: Observable, base: Any, f: Arity<3>) -> Observer',
            (o, base, f) => observer(push => {
                let v = base
                let n = 0
                return obsv(o).subscribe(subs({
                    next: x => {
                        v = push(call(f, [x, v, n]))
                        n += 1
                    },
                    error: e => push(e),
                    complete: () => push(Complete)
                }))
            }),
        'function scan (o: Observable, base: Any, f: Arity<2>) -> Observer',
            (o, base, f) => observer(push => {
                let v = base
                return obsv(o).subscribe(subs({
                    next: x => {
                        v = push(call(f, [x, v]))
                    },
                    error: e => push(e),
                    complete: () => push(Complete)
                }))
            }),
        'function scan (i: Iterable, base: Any, f: Arity<3>) -> Iterator',
                (i, base, f) => (function* () {
                    let v = base
                    let n = 0
                    for (let e of iter(i)) {
                        v = call(f, [e, v, n])
                        yield v
                        n += 1
                    }
                })(),
        'function scan (i: Iterable, base: Any, f: Arity<2>) -> Iterator',
                (i, base, f) => (function* () {
                    let v = base
                    for (let e of iter(i)) {
                        v = call(f, [e, v])
                        yield v
                    }
                })()
    ),
    fold: f (
        'fold',
        'function fold (o: Observable, base: Any, f: Arity<3>) -> Observer',
            (o, base, f) => observer(push => {
                let v = base
                let n = 0
                return obsv(o).subscribe(subs({
                    next: x => {
                        v = call(f, [x, v, n])
                        n += 1
                    },
                    error: e => push(e),
                    complete: () => {
                        push(v)
                        push(Complete)
                    }
                }))
            }),
        'function fold (o: Observable, base: Any, f: Arity<2>) -> Observer',
            (o, base, f) => observer(push => {
                let v = base
                return obsv(o).subscribe(subs({
                    next: x => {
                        v = call(f, [x, v])
                    },
                    error: e => push(e),
                    complete: () => {
                        push(v)
                        push(Complete)
                    }
                }))
            }),
        'function fold (i: Iterable, base: Any, f: Arity<3>) -> Object',
            (i, base, f) => fold(iter(i), base, (e, v, n) => {
                return call(f, [e, v, n])
            }),
        'function fold (i: Iterable, base: Any, f: Arity<2>) -> Object',
            (i, base, f) => fold(iter(i), base, (e,v) => {
                return call(f, [e, v])
            })
    ),
    every: f (
        'every',
        'function every (i: Iterable, f: Arity<2>) -> Bool',
            (i, f) => forall(iter(i), (e, i) => {
                let v = call(f, [e, i])
                ensure(is(v, Types.Bool), 'cond_not_bool')
                return v
            }),
        'function every (i: Iterable, f: Arity<1>) -> Bool',
            (i, f) => forall(iter(i), e => {
                let v = call(f, [e])
                ensure(is(v, Types.Bool), 'cond_not_bool')
                return v
            })
    ),
    some: f (
        'some',
        'function some (i: Iterable, f: Arity<2>) -> Bool',
            (i, f) => exists(iter(i), (e, i) => {
                let v = call(f, [e, i])
                ensure(is(v, Types.Bool), 'cond_not_bool')
                return v
            }),
        'function some (i: Iterable, f: Arity<1>) -> Bool',
            (i, f) => exists(iter(i), e => {
                let v = call(f, [e])
                ensure(is(v, Types.Bool), 'cond_not_bool')
                return v
            })
    ),
    join: fun (
        'function join (i: Iterable, sep: String) -> String',
            (i, sep) => {
                let string = ''
                let first = true
                for (let e of i) {
                    ensure(is(e, Types.String), 'element_not_string')
                    if (first) {
                        first = false
                    } else {
                        string += sep
                    }
                    string += e
                }
                return string
            }
    ),
    reversed: fun (
        'function reversed (i: Iterable) -> Iterator',
            i => (function* () {
                let buf = []
                for (let e of iter(i)) {
                    buf.push(e)
                }
                for (let e of rev(buf)) {
                    yield e
                }
            })(),
        'function reversed (l: List) -> Iterator',
            l => rev(l)
    ),
    flat: f (
        'flat',
        'function flat (o: Observable, limit: PosInt) -> Observer',
            (o, limit) => observer(push => {
                let source_complete = false
                let stopped = false
                let waiting = []
                let unsub = new Set()
                let unsub_source = obsv(o).subscribe(subs({
                    next: function next_callback (x) {
                        if (stopped) { return }
                        let ok = is(x, Types.Observable)
                        ensure(ok, 'value_not_observable')
                        if (unsub.size < limit) {
                            let unsub_i = obsv(x).subscribe(subs({
                                next: y => push(y),
                                error: e => {
                                    stop()
                                    push(e)
                                },
                                complete: () => {
                                    unsub.delete(unsub_i)
                                    if (waiting.length > 0) {
                                        next_callback(waiting.shift())
                                    } else if (source_complete) {
                                        if (unsub.size == 0) {
                                            push(Complete)
                                        }
                                    }
                                }
                            }))
                            unsub.add(unsub_i)
                        } else {
                            waiting.push(x)
                        }
                    },
                    error: e => {
                        stop()
                        push(e)
                    },
                    complete: () => {
                        source_complete = true
                    }
                }))
                let stop = () => {
                    unsub_source()
                    foreach(unsub, u => u())
                    unsub.clear()
                    waiting = []
                    stopped = true
                }
                return stop
            }),
        'function flat (o: Observable) -> Observer',
            o => observer(push => {
                let source_complete = false
                let stopped = false
                let unsub = new Set()
                let unsub_source = obsv(o).subscribe(subs({
                    next: x => {
                        if (stopped) { return }
                        let ok = is(x, Types.Observable)
                        ensure(ok, 'value_not_observable')
                        let unsub_i = obsv(x).subscribe(subs({
                            next: y => push(y),
                            error: e => {
                                stop()
                                push(e)
                            },
                            complete: () => {
                                unsub.delete(unsub_i)
                                if (source_complete && unsub.size == 0) {
                                    push(Complete)
                                }
                            }
                        }))
                        unsub.add(unsub_i)
                    },
                    error: e => {
                        stop()
                        push(e)
                    },
                    complete: () => {
                        source_complete = true
                    }
                }))
                let stop = () => {
                    unsub_source()
                    foreach(unsub, u => u())
                    unsub.clear()
                    stopped = true
                }
                return stop
            }),
        'function flat (i: Iterable) -> Iterator',
            i => (function* () {
                for (let e of iter(i)) {
                    ensure(is(e, Types.Iterable), 'element_not_iterable')
                    for (let ee of iter(e)) {
                        yield ee
                    }
                }
            })()
    ),
    tap: f (
        'tap',
        'function tap (o: Observable, f: Arity<2>) -> Observer',
            (o, f) => observer(push => {
                let i = 0
                return obsv(o).subscribe(subs({
                    next: x => {
                        call(f, [x, i])
                        push(x)
                        i += 1
                    },
                    error: e => push(e),
                    complete: () => push(Complete)
                }))
            }),
        'function tap (o: Observable, f: Arity<1>) -> Observer',
            (o, f) => observer(push => obsv(o).subscribe(subs({
                next: x => {
                    call(f, [x])
                    push(x)
                },
                error: e => push(e),
                complete: () => push(Complete)
            }))),
        'function tap (i: Iterable, f: Arity<2>) -> Iterator',
            (i, f) => (function* () {
                let n = 0
                for (let e of iter(i)) {
                    call(f, [e, n])
                    yield e
                    n += 1
                }
            })(),
        'function tap (i: Iterable, f: Arity<1>) -> Iterator',
            (i, f) => (function* () {
                for (let e of iter(i)) {
                    call(f, [e])
                    yield e
                }
            })()
    ),
    zip: f (
        'zip',
        'function zip (o1: Observable, o2: Observable) -> Observer',
            (o1, o2) => observer(push => {
                let q1 = []
                let q2 = []
                let stopped = false
                let unsub1 = obsv(o1).subscribe(subs({
                    next: x => {
                        if (stopped) { return }
                        if (q2.length > 0) {
                            push([x, q2.shift()])
                        } else {
                            q1.push(x)
                        }
                    },
                    error: e => { stop(); push(e) },
                    complete: () => { unsub2(); push(Complete) }
                }))
                let unsub2 = obsv(o2).subscribe(subs({
                    next: x => {
                        if (stopped) { return }
                        if (q1.length > 0) {
                            push([q1.shift(), x])
                        } else {
                            q2.push(x)
                        }
                    },
                    error: e => { stop(); push(e) },
                    complete: () => { unsub1(); push(Complete) }
                }))
                let stop = () => {
                    unsub1()
                    unsub2()
                    q1 = []
                    q2 = []
                    stopped = true
                }
                return stop
            }),
        'function zip (i1: Iterable, i2: Iterable) -> Iterator',
            (i1, i2) => zip([i1, i2], x => x)
    ),
    collect: fun (
        'function collect (i: Iterable) -> List',
            i => list(iter(i))
    )
})
