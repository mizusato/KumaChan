invoke {
    assert 5 + 3 == 8
    assert 5 * 3 == 15
    assert 5 / 2 == 2.5
    assert 5 % 2 == 1
    assert 5 ^ 3 == 125
    assert -5 ^ 3 == -125
    assert 'a' + 'b' == 'ab'
    assert (str 5) == '5'
    assert (true || true) == true
    assert (true || false) == true
    assert (false || true) == true
    assert (false || false) == false
    assert (true && true) == true
    assert (true && false) == false
    assert (false && true) == false
    assert (false && false) == false
    assert !false == true
    assert !true == false
    assert not false == true
    assert not true == false
    assert ((1+1 == 2) => 1) == 1
    assert ((1+1 == 3) => 1) is Nil
    assert (1 or 2) == 1
    assert (Nil or 2) == 2
    assert (abs << -5) == 5 
    let h1 = { a: 1 }
    let h2 = { b: 2 }
    let h3 = {} << h1 << h2
    assert h3.a == 1 && h3.b == 2
    assert h1 ~~ h1
    assert h1 !~ { a: 1 }
    let h4 = (copy h1)
    set h1.a = 4
    assert h4.a == 1
    assert len([]) == 0
    assert len([1,2,3]) == 3
}
