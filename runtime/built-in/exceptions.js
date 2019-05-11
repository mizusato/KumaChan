let wrapped_assert = fun (
    'function assert (p: Bool) -> Void',
        p => (assert(p), Void)
)

let wrapped_panic = fun (
    'function panic (msg: String) -> Void',
        msg => panic(msg)
)

let wrapped_throw = fun (
    'function throw (e: Error) -> Void',
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
        e.type = 1
        pour(e, { name, args })
    }
    throw new EnsureFailed(name, file, row, col)
}

function try_failed (e, error, name) {
    if (e) {
        e.type = 2
        e.name = name
    }
    throw error
}

function inject_ensure_args (scope, names, e) {
    assert(scope instanceof Scope)
    assert(is(names, TypedList.of(Types.String)))
    assert(is(e.args, Types.List))
    foreach(names, (name, i) => {
        if (i < e.args.length) {
            scope.declare(name, e.args[i])
        } else {
            scope.declare(name, Nil)
        }
    })
}
