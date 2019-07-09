invoke {
    type A = singleton
    type B = singleton
    assert A is Type
    assert B is Type
    assert A is A
    assert B is B
    assert A is not B
    assert B is not A
    assert A <!=> B
    assert A !~ B
    assert Nil is Nil
    assert Void is Void
    assert NotFound is NotFound
    let h = {}
    assert h.key? ~~ Nil
    assert (invoke { do nothing }) ~~ Void
    assert ([1,2,3,4] -> find -> .{ .x == 5 }) ~~ NotFound
}
