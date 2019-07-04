invoke {
    type Maybe1<T> = T | Nil
    assert Maybe1<Int> <=> Nil | Int
    assert Maybe1<Int> ~~ Maybe1<Int>
    assert 1 is Maybe1<Int>
    assert Nil is Maybe1<Int>
    assert 1.5 is not Maybe1<Int>
    assert '1' is not Maybe1<Int>
    assert Maybe1<(Int | String)> ~~ Maybe1<(String | Int)>
    assert Maybe1<(Int | String)> <=> Maybe1<(String | Int)>
}
