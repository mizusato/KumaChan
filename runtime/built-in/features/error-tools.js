let wrapped_assert = fun (
    'function assert (p: Bool) -> Void',
        p => p? Void: panic('Assertion Failed')
)


let wrapped_panic = f (
    'panic',
    'function panic (e: Error) -> Never',
        e => {
            throw new RuntimeError('panic: ' + e.message)
        },
    'function panic (msg: String) -> Never',
        msg => {
            panic(msg)
        }
)


let wrapped_throw = fun (
    'function throw (e: Error) -> Never',
        e => {
            if (e instanceof CustomError) {
                e.trace = get_trace()
                throw e
            } else {
                throw e
            }
        }
)


function ensure_failed (e, name, args, file, row, col) {
    if (e) {
        e[DUMP_TYPE] = DUMP_ENSURE
        e[DUMP_NAME] = name
        e[DUMP_ARGS] = args
    }
    throw new EnsureFailed(name, file, row, col)
}


function try_failed (e, error, name) {
    if (e) {
        e[DUMP_TYPE] = DUMP_TRY
        e[DUMP_NAME] = name
    }
    throw error
}


function inject_ensure_args (scope, names, e) {
    assert(scope instanceof Scope)
    assert(is(names, TypedList.of(Types.String)))
    assert(is(e[DUMP_ARGS], Types.List))
    foreach(names, (name, i) => {
        if (i < e[DUMP_ARGS].length) {
            scope.declare(name, e[DUMP_ARGS][i])
        } else {
            scope.declare(name, Nil)
        }
    })
}
