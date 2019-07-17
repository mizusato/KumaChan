invoke {
    import Set from Std
    let s = Set<Int>()
    assert s -> size() == 0
    s -> add(5)
    s -> add(7)
    s -> add(11)
    s -> add(13)
    assert s -> size() == 4
    s -> remove(13)
    assert s -> size() == 3
    assert !(s -> has(13))
    assert s -> has(11)
    var p = 1
    for element in s {
        reset p *= element
    }
    assert p == 5*7*11
    s -> clear()
    assert s -> size() == 0
    assert !(s -> has(5))
    let s1 = copy(s)
    assert s1 is Set<Int>
    s1 -> add(9)
    assert s1 -> has(9)
    assert not (s -> has(9))
    assert s -> size() == 0
}
