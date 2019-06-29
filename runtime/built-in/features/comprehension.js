let iterator_comprehension = fun (
    'function comprehension (v: Function, l: List, f: Function) -> Iterator',
        (v, l, f) => {
            foreach(l, (element, index) => {
                ensure(is(element, Types.Iterable), 'not_iterable', index+1)
            })
            l = l.map(element => iter(element))
            return (function* () {
                for (let values of zip(l, x => x)) {
                    let ok = call(f, values)
                    assert(is(ok, Types.Bool))
                    if (ok) {
                        yield call(v, values)
                    }
                }
            })()
        }
)


let list_comprehension = fun (
    'function list_comprehension (v: Function, l: List, f: Function) -> List',
        (v, l, f, scope) => list(iterator_comprehension[WrapperInfo].raw(scope))
)
