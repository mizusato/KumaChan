promise {
    // Ordinary IIFE
    var x = 0
    let y = invoke {
        reset x = 1
        return 2^10
    }
    assert x == 1
    assert y == 1024
    // Iterator (Generator IIFE)
    let i = iterator {
        for n in [3,5,7] {
            yield n
        }
    }
    assert i is Iterator
    var sum = 0
    for n in i {
        reset sum += n
    }
    assert sum == 3+5+7
    // Promise (Async IIFE)
    var q = 'a'
    let p = promise {
        reset q = 'b'
        return 1
    }
    assert p is Promise
    await r = p
    assert q == 'b'
    assert r == 1
    // Async Iterator (Async Generator IIFE)
    let ai = async iterator {
        for n in [2,3,5] {
            await v = promise { return n }
            yield v
        }
    }
    assert ai is AsyncIterator
    var product = 1
    await element in ai {
        reset product *= element
    }
    assert product == 2*3*5
}
