/**
 *  Error Handling Mechanics
 */


/* Depth Limit of Stack Backtrace */
const TRACE_DEPTH = 16

/* Unrecoverable Fatal Error */
class RuntimeError extends Error {
    constructor (msg) {
        super(msg)
        this.name = 'RuntimeError'
    }
}

/* Recoverable Error Produced by `ensure` Command */
class EnsureFailed extends Error {
    constructor (name, file, row, col) {
        super (
            `validation for '${name}' failed`
            + ` at ${file} (row ${row}, column ${col})`
        )
        this.name = "EnsureFailed"
    }
}

/* Recoverable Error Produced by `throw` Commmand */
class CustomError extends Error {
    constructor (message, name, data) {
        let trace = get_trace()
        super(message + LF + LF + stringify_trace(trace) + LF)
        this.name = name
        this.data = data
        this.trace = trace
    }
}

/* Shorthand for `new CustomError(...)` */
function create_error (msg, name = 'CustomError', data = {}) {
    // this function is used by `built-in/exception.js`
    return new CustomError(msg, name, data)
}


/**
 *  Call Stack (only used to store debug info)
 *
 *  Fields of frame:
 *     call_type: Integer,  (1: userland, 2: overload, 3: built-in)
 *     desc: String,  (description of function)
 *     file: String,  (position of call expression in source code)
 *     row: Integer,  (position of call expression in source code)
 *     col: Integer   (position of call expression in source code)
 */
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
            // called at script file
            let pos = ''
            if (frame.row != -1) {
                pos = `(row ${frame.row}, column ${frame.col})`
            }
            point = `from ${frame.file} ${pos}`
        } else if (frame.call_type == 2) {
            // specific function called by overloaded function
            point = 'from <overload>'
        } else if (frame.call_type == 3) {
            // called by built-in function
            point = 'from <built-in>'
        }
        return (frame.desc + LF + INDENT + point)
    }))
    return info_list
}

function clear_call_stack () {
    // Note: before throwing a RuntimeError this function should be called
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


/**
 *  Produces a RuntimeError with Stack Backtrace
 *
 *  This function is called by `ensure()` and `panic()`.
 *     `ensure()` is called by built-in functions to produce fatal error.
 *     `panic()` is used by `assert` command and `panic` command.
 */
function produce_error (msg) {
    // produce a RuntimeError (fatal error)
    let trace = get_trace()
    clear_call_stack()
    let err = new RuntimeError (
        msg + LF + LF + stringify_trace(trace) + LF
    )
    err.trace = trace
    throw err
}

/* produces a built-in panic if given bool condition not satisfied */
function ensure (bool, msg_type, ...args) {
    if (bool) { return }
    produce_error(get_msg(msg_type, args))
}

/* produces a userland panic */
function panic (msg) {
    // this function is used by `built-in/exception.js`
    produce_error(`panic: ${msg}`)
}

/* get error message from MSG defined in `msg.js` */
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


/**
 *  Convert from Unhandled Recoverable Error to Fatal Error
 *
 *  This function is called at the end of handle hook.
 *  As in this language we use function-level error handling,
 *    if the end of handle hook is reached, it means that
 *    the error caught by the handle hook is not handled correctly,
 *    therefore it is necessary to covert the error to a fatal error.
 *  Note that RuntimeError is thrown at the start of handle hook,
 *    so the argument `error` can't be a RuntimeError.
 */
function convert_to_fatal (error) {
    assert(error instanceof Error)
    assert(!(error instanceof RuntimeError))
    if (error instanceof EnsureFailed || error instanceof CustomError) {
        /**
         *  This branch may be entered becase of:
         *    1. missing corresponding `unless` command for `ensure` command.
         *    2. missing corresponding `failed` command for `try` command.
         *    3. both `return` command and `throw` command not called in
         *         the handle hook.
         */
        clear_call_stack()
        return new RuntimeError(`Unhandled ${error.name}: ${error.message}`)
    } else {
        /**
         *  This branch may be enter because of:
         *    1. `assert()` (in `assertion.js`) raised an AssertionError
         *    2. other errors such as 'cannot read *** of undefined' happended
         */
        let msg = ''
        if (error instanceof AssertionFailed) {
            msg = 'Internal Assertion Failed'
        } else {
            msg = `Internal ${error.name}: ${error.message}`
        }
        clear_call_stack()
        let e = new RuntimeError(msg)
        // inform the REPL to print the internal JS call stack backtrace
        e.is_internal = true
        return e
    }
}


/**
 *  Prevents RuntimeError from being caught by Promise.prototype.catch(),
 *    used to wrap async functions.
 */
function async_e_wrap (async_raw_function) {
    assert(typeof async_raw_function == 'function')
    return function (scope) {
        return new Promise((resolve, reject) => {
            let p = async_raw_function(scope)
            assert(p instanceof Promise)
            p.then(value => {
                resolve(value)
            }).catch(error => {
                if (error instanceof RuntimeError) {
                    // just throw it, preserve the pending state
                    throw error
                } else {
                    reject(error)
                }
            })
        })
    }
}
