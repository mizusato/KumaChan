invoke {
    let i = .[ x, for x in [1,2] ]
    assert i is Iterator
    let l = collect(i)
    assert len(l) == 2 && l[0] == 1 && l[1] == 2
    let l1 = [ [x, y], for x in [2, 3], y in [5, 7, 9] ]
    assert len(l1) == 2
    assert l1[0][0] == 2 && l1[0][1] == 5
    assert l1[1][0] == 3 && l1[1][1] == 7
    let l2 = [ x, for x in range(0,10), where x % 7 == 0 ]
    assert len(l2) == 2
    assert l2[0] == 0
    assert l2[1] == 7
}
