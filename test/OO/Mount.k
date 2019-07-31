invoke {
    class Delegate {
        init (f: Callable) {
            do nothing
        }
        work () -> Int {
            return f() + 1
        }
    }
    class A is Delegate {
        init (x: Int) {
            let d = mount Delegate(f)
        }
        private f () -> Int {
            return x
        }
        double_work () -> Int {
            return (d->work())^2
        }
    }
    let a = A(9)
    assert a is A
    assert a is Delegate
    assert a -> work() == 10
    assert a -> double_work() == 100
    class B is A {
        init (x: Int) {
            mount A(x)
        }
    }
    let b = B(9)
    assert b is B
    assert b is A
    assert b is Delegate
    assert b -> work() == 10
    assert b -> double_work() == 100
    class C { init() {} }
    class D {
        init () {
            do nothing
        }
        hello () -> String {
            return 'Hello'
        }
    }
    class E is A, C, D {
        init (x: Int) {
            mount C()
            mount A(x)
            mount D()
        }
    }
    let e = E(1)
    assert e is A
    assert e is C
    assert e is D
    assert e -> work() == 2
    assert e -> hello() == 'Hello'
}
