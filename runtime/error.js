/**
 *  Error Handling Mechanics
 */
const TRACE_DEPTH = 16
const CALL_FROM_SCRIPT = 1
const CALL_FROM_OVERLOAD = 2
const CALL_FROM_BUILT_IN = 3


/**
 *  Unrecoverable Fatal Error
 */
class RuntimeError extends Error {
    constructor (msg) {
        super(msg)
        this.name = 'RuntimeError'
    }
}

/**
 *  Recoverable Error Produced by `ensure` Command
 */
class EnsureFailed extends Error {
    constructor (name, file, row, col) {
        super (
            `validation for '${name}' failed`
            + ` at ${file} (row ${row}, column ${col})`
        )
        this.name = "EnsureFailed"
    }
}

/**
 *  Recoverable Error Produced by `throw` Commmand
 */
class CustomError extends Error {
    constructor (message, name = 'CustomError', data = {}) {
        let trace = get_trace()
        super(message + LF + LF + stringify_trace(trace) + LF)
        this.name = name
        this.data = data
        this.trace = trace
    }
}


/**
 *  Fatal errors are defined here.
 *  They won't be caught by handle hooks.
 */
let FatalErrorClasses = [
    RuntimeError, AssertionFailed,
    RangeError, ReferenceError, SyntaxError, TypeError
]
function is_fatal (error) {
    if (!(error instanceof Error)) {
        return true
    }
    return exists(FatalErrorClasses, E => error instanceof E)
}


/**
 *  Call Stack (only used to store debug info)
 *
 *  Fields of frame:
 *     call_type: Integer,  (1: script, 2: overload, 3: built-in)
 *     desc: String,  (description of function)
 *     file: String,  (position of call expression in source code)
 *     row: Integer,  (position of call expression in source code)
 *     col: Integer   (position of call expression in source code)
 */
let call_stack = []

function push_call (call_type, desc, file = null, row = -1, col = -1) {
    assert(Number.isInteger(call_type))
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
        if (frame.call_type == CALL_FROM_SCRIPT) {
            // called at script file
            let pos = ''
            if (frame.row != -1) {
                pos = `(row ${frame.row}, column ${frame.col})`
            }
            point = `from ${frame.file} ${pos}`
        } else if (frame.call_type == CALL_FROM_OVERLOAD) {
            // specific function called by overloaded function
            point = 'from <overload>'
        } else if (frame.call_type == CALL_FROM_BUILT_IN) {
            // called by built-in function
            point = 'from <built-in>'
        }
        return (frame.desc + LF + INDENT + point)
    }))
    return info_list
}

function stringify_trace (trace) {
    return join(take(rev(trace), TRACE_DEPTH), LF)
}


/**
 *  Produces a RuntimeError with Stack Backtrace
 *
 *  This function is called by `ensure()` and `panic()`.
 *     `ensure()` is called by built-in functions to produce fatal error.
 *     `panic()` is used by `assert` command and `panic` command.
 */
function crash (msg) {
    let trace = get_trace()
    let err = new RuntimeError (
        msg + LF + LF + stringify_trace(trace) + LF
    )
    err.trace = trace
    throw err
}

/**
 *  Crashes if given bool condition is not satisfied
 */
function ensure (bool, msg_type, ...args) {
    if (bool) { return }
    crash(get_msg(msg_type, args))
}

/**
 *  Crashes with a message
 */
function panic (msg) {
    // this function is used by `built-in/exception.js`
    crash(`panic: ${msg}`)
}

/**
 *  Gets error message from MSG defined in `msg.js`
 */
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
 *  This function is called at the beginning of handle hooks.
 *
 *  Since fatal errors are considered unrecoverable,
 *    they should not be caught by handle hooks.
 *  At the beginning of handle hooks, we simply re-throw
 *    any kind of fatal errors.
 */
function enter_handle_hook (error) {
    if (is_fatal(error)) {
        throw error
    }
}


/**
 *  This function is called at the end of handle hooks.
 *
 *  Because in this language we use function-level error handling,
 *    if the end of handle hook is reached, it means that
 *    the error caught by the handle hook is not handled correctly,
 *    therefore it is necessary to covert the error to a fatal error.
 *  Note that fatal errors have been thrown at the start of handle hook,
 *    so the argument `error` can't be a fatal error.
 */
function exit_handle_hook (error) {
    assert(!is_fatal(error))
    throw new RuntimeError(`Unhandled ${error.name}: ${error.message}`)
}


/**
 *  Wrappers to prevent fatal errors from being caught by promise.catch(...)
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
                if (is_fatal(error)) {
                    throw error
                } else {
                    reject(error)
                }
            })
        })
    }
}

function async_gen_e_wrap (raw_ag) {
    assert(typeof raw_ag == 'function')
    return async function* (scope) {
        try {
            for await (let element of raw_ag(scope)) {
                yield element
            }
        } catch (error) {
            if (is_fatal(error)) {
                await new Promise(() => {
                    throw error
                })
            } else {
                throw error
            }
        }
    }
}
