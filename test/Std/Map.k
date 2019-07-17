invoke {
    import Map from Std
    let m = Map<Int,String>()
    assert m -> size() == 0
    set m[1] = 'one'
    set m[2] = 'two'
    assert m -> size() == 2
    assert m[1] == 'one'
    assert m[2] == 'two'
    for { i: v } in m {
        assert when {
            i == 1: v == 'one',
            i == 2: v == 'two'
        }
    }
    m -> remove(1)
    assert m -> size() == 1
    assert m[1]? is Nil
    assert not (m -> has(1))
    assert m -> has(2)
    m -> clear()
    assert m -> size() == 0
    assert m[2]? is Nil
    let m1 = copy(m)
    assert m1 is Map<Int,String>
    set m1[9] = '9'
    assert m1[9] == '9'
    assert m[9]? is Nil
    assert m -> size() == 0
}
