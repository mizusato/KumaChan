let for_loop_e = fun (
    'function for_loop_enumerate (e: Enumerable) -> Iterator',
        e => {
            let list = enum_(e)
            let keys = list.get('keys')
            let values = list.get('values')
            return (function* () {
                assert(keys.length == values.length)
                let L = keys.length
                for (let i = 0; i < L; i += 1) {
                    yield { key: keys[i], value: values[i] }
                }
            })()
        }
)


let for_loop_i = fun (
    'function for_loop_iterate (i: Iterable) -> Iterator',
        i => map(iter(i), (e, i) => ({ key: i, value: e }))
)


let for_loop_a = fun (
    'function for_loop_async (ai: AsyncIterable) -> AsyncIterator',
        ai => (async function* () {
            for await (let e of async_iter(ai)) {
                yield { key: Nil, value: e }
            }
        })()
)
