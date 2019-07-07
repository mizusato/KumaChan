promise {
    await x = promise {
        throw custom_error('intentional error 1')
    } -> catch -> lambda(err) {
        return 5
    }
    await y = promise {
        await postpone(50)
        return 1
    } -> then -> lambda(val) {
        throw custom_error('intentional error 2')
    } -> catch -> lambda(err) {
        return 7
    }
    await z = promise {
        await postpone(50)
        throw custom_error('intentional error 3')
    } -> catch -> lambda(err) {
        throw custom_error('intentional error 4')
    } -> catch -> lambda(err) {
        return 11
    }
    await w = promise {
        throw custom_error('intentional error 5')
    } -> finally -> lambda(err) {
        throw custom_error('intentional error 6')
    } -> catch -> lambda(err) {
        return 13
    }
    assert x*y*z*w == 5*7*11*13
}
