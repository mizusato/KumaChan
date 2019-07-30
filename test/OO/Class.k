invoke {
    class A {
        init (x: Int) {
            do nothing
        }
        unwrap () -> Int {
            return x
        }
    }
    assert A is Class
    assert A(0) is Instance
    assert A(0) is A
    assert A(0)->unwrap() == 0
}
