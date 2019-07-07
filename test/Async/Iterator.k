promise {
    function delayed_sequence (n: Int) -> AsyncIterator {
        return async iterator {
            for i in seq(n) {
                await postpone(25)
                yield i
            }
        }
    }
    var sum = 0
    await x in delayed_sequence(5) {
        reset sum += x
    }
    assert sum == 10
}
