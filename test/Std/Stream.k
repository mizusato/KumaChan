promise {
    import Stream, AsyncStream from Std
    let s = Stream<Int>([3,5,7])
    var p = 1
    for element in s {
        reset p *= element
    }
    assert p == 3*5*7
    let a = AsyncStream<Int>(async iterator {
        for i in [7,11,13] {
            await postpone(20)
            yield i
        }
    })
    reset p = 1
    await element in a {
        reset p *= element
    }
    assert p == 7*11*13
}
