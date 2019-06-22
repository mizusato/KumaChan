/**
 *  Internal Assertion
 *
 *  Mechanics to prevent unexpected value from being spread
 *    by the weak type system of JavaScript.
 *  When an unexcepted value is detected, the execution of script
 *    terminates immediately due to an AssertionFailed error.
 */


class AssertionFailed extends Error {
    constructor () {
        super('Internal Assertion Failed')
        this.name = 'RuntimeError'
    }
}


function assert (value) {
    if(!value) {
        throw new AssertionFailed()
    }
    // this `return` is intentional, don't remove it
    return true
}
