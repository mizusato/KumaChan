invoke {
    class A {
        init (x: Int) {
            do nothing
        }
        private f (y: Int) -> Int {
            return x * y
        }
        calc (y: Int) -> Int {
            assert f is Overload
            return f(y)
        }
    }
    assert A(4) -> calc(5) == 20
}
