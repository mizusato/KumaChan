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
    return true
}
