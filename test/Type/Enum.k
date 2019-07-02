invoke {
    enum E { A, B, C }
    enum E1 { A, B, C }
    assert E.A is E
    assert E.B is E
    assert E.C is E
    assert E1.A is not E
    assert Nil is not E
    assert E1.A !~ E.A
    assert E1.A <!=> E.A
    assert E is Enumerable
    assert E is Iterable
    for Val in E {
        assert Val is E
    }
    for { name: Val } in E {
        assert Val is E
        assert E[name] ~~ Val
        assert E[name] <=> Val
    }
}
