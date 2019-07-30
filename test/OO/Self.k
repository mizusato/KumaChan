invoke {
    class A {
        init (x: Int) {
            do nothing
        }
        foo () -> Int {
            return x
        }
        bar () -> Int {
            assert self is Instance
            assert self is A
            return self->foo() + 1
        }
    }
    assert A(2)->foo() == 2
    assert A(2)->bar() == 3
}
