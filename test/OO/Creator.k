invoke {
    class A {
        init (x: Int) {
            do nothing
        }
        create (x: Int, y: Int) {
            return A(x * y)
        }
        create (x: Int, y: Int, z: Int) {
            return A(x * y * z)
        }
        create (x: Int) {
            // won't be called
            return A(-x)
        }
        unwrap () -> Int {
            return x
        }
    }
    assert A(2,3)->unwrap == 6
    assert A(2,3,5)->unwrap == 30
    assert A(9)->unwrap == 9
}
