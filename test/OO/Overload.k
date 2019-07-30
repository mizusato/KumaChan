invoke {
    interface I {
        foo (x: Int) -> Int
        foo (x: Int, y: Int) -> Int
        bar () -> Int
        baz (x: Int) -> Int {
            return foo(x) + bar()
        }
        baz () -> Int {
            return bar()
        }
    }
    class A is I {
        init () {
            do nothing
        }
        foo (x: Int) -> Int {
            return x + 1
        }
        foo (x: Int, y: Int) -> Int {
            return x * y
        }
        bar () -> Int {
            return 999
        }
        bar (x: Int) -> String {
            return str(x)
        }
    }
    let a = A()
    assert a->foo(5) == 6
    assert a->foo(5,7) == 35
    assert a->bar() == 999
    assert a->bar(9) == '9'
    assert a->baz() == 999
    assert a->baz(1) == 1001
}
