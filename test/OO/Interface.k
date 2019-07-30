invoke {
    interface I {}
    class A is I {
        init () {
            do nothing
        }
        unwrap() -> Int {
            return 0
        }
    }
    assert A() is I
    assert A is Impl<I>
    interface IntBox {
        unwrap () -> Int
        square_unwrap () -> Int {
            return unwrap() ^ 2
        }
    }
    assert A is not IntBox
    class B is I, IntBox {
        init (x: Int) {
            do nothing
        }
        unwrap () -> Int {
            return x
        }
    }
    assert B is Impl<I>
    assert B is Impl<IntBox>
    let b = B(3)
    assert b is I
    assert b is IntBox
    assert b->unwrap() == 3
    assert b->square_unwrap() == 9
    class C is B {
        init (x: Int) {
            mount B(x)
        }
    }
    assert C is Impl<I>
    assert C is Impl<IntBox>
    let c = C(3)
    assert c is I
    assert c is IntBox
    assert c->unwrap() == 3
    assert c->square_unwrap() == 9
}
