const TRACE_DEPTH = 10

class RuntimeError extends Error {
    constructor (msg) {
        super(msg)
        this.name = "RuntimeError"
    }
}

class UserlandEnsureFailed extends Error {
    constructor (name) {
        super(`validation for '${name}' failed`)
        this.name = "EnsureFailed"
    }
}

class UserlandTryFailed extends Error {
    constructor (name) {
        super(`try of action '${name}' failed`)
        this.name = "TryFailed"
    }
}


let call_stack = []

function push_call (call_type, desc, file = null, row = -1, col = -1) {
    assert(exists([1,2,3], i => call_type === i))
    call_stack.push({ call_type, desc, file, row, col })
}

function pop_call () {
    call_stack.pop()
}

function collect_stack () {
    let info_list = list(map(call_stack, frame => {
        let point = 'from <unknown>'
        if (frame.call_type == 1) {
            // userland
            point = `from ${frame.file} (row ${frame.row}, column ${frame.col})`
        } else if (frame.call_type == 2) {
            // overload
            point = 'from <overload>'
        } else if (frame.call_type == 3) {
            // built-in
            point = 'from <built-in>'
        }
        return (frame.desc + LF + INDENT + point)
    }))
    call_stack = []
    return info_list
}

function produce_error (msg) {
    let trace = collect_stack()
    let err = new RuntimeError (
        msg + LF + LF + join(take(rev(trace), TRACE_DEPTH), LF) + LF
    )
    err.trace = trace
    throw err
}

function assert (value) {
    if(!value) { produce_error('Assertion Failed') }
    return value
}

function panic (msg) {
    produce_error(`panic: ${msg}`)
}

function get_msg (msg_type, args) {
    assert(typeof msg_type == 'string')
    assert(args instanceof Array)
    assert(forall(args, arg => typeof arg == 'string'))
    let msg = MSG[msg_type]
    if (typeof msg == 'string') {
        return msg
    } else {
        assert(typeof msg == 'function')
        return msg.apply(null, args)
    }
}

function ensure (bool, msg_type, ...args) {
    if (bool) { return }
    produce_error(get_msg(msg_type, args))
}
