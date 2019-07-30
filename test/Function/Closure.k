invoke {
    function adder (n: Int) -> Function {
        return lambda (x: Int) -> Int {
            return n + x
        }
    }
    assert adder(5) is Function
    assert adder(7)(9) == 16
    function counter () -> Function {
        var n = 0
        return lambda () -> Int {
            reset n += 1
            return n
        }
    }
    let c = counter()
    assert c() == 1
    assert c() == 2
    assert c() == 3
}
