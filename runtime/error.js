const TRACE_DEPTH = 10

class RuntimeError extends Error {
    constructor (msg) {
        super(msg)
        this.name = "RuntimeError"
    }
}

class EnsureFailed extends Error {
    constructor (name, file, row, col) {
        super (
            `validation for '${name}' failed`
            + ` at ${file} (row ${row}, column ${col})`
        )
        this.name = "EnsureFailed"
    }
}

class CustomError extends Error {
    constructor (message, name, data) {
        let trace = get_trace()
        super(message + LF + LF + stringify_trace(trace) + LF)
        this.name = name
        this.data = data
        this.trace = trace
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

function get_top_frame () {
    if (call_stack.length > 0) {
        return call_stack[call_stack.length-1]
    } else {
        return null
    }
}

function get_trace () {
    let info_list = list(map(call_stack, frame => {
        let point = 'from <unknown>'
        if (frame.call_type == 1) {
            // userland
            let pos = ''
            if (frame.row != -1) {
                pos = `(row ${frame.row}, column ${frame.col})`
            }
            point = `from ${frame.file} ${pos}`
        } else if (frame.call_type == 2) {
            // overload
            point = 'from <overload>'
        } else if (frame.call_type == 3) {
            // built-in
            point = 'from <built-in>'
        }
        return (frame.desc + LF + INDENT + point)
    }))
    return info_list
}

function clear_call_stack () {
    call_stack = []
}

function stringify_trace (trace) {
    return join(take(rev(trace), TRACE_DEPTH), LF)
}

function get_call_stack_pointer () {
    return call_stack.length - 1
}

function restore_call_stack (pointer) {
    call_stack = call_stack.slice(0, pointer+1)
}

function produce_error (msg) {
    let trace = get_trace()
    clear_call_stack()
    let err = new RuntimeError (
        msg + LF + LF + stringify_trace(trace) + LF
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

function create_error (msg, name = 'CustomError', data = {}) {
    return new CustomError(msg, name, data)
}

function get_msg (msg_type, args) {
    assert(typeof msg_type == 'string')
    assert(args instanceof Array)
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
