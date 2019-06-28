invoke {
    assert 1 + 2 == 3
    assert 2 * 3 == 6
    assert 16 / 2 == 8
    assert 16 % 2 == 0
    assert 17 % 2 == 1
    assert 2^10 == 1024
    assert 0 is Number
    assert 0 is Int
    assert 2e5 is Number
    assert 2e5 is Int
    assert 0xff is Number
    assert 0xff is Int
    assert 0xA0 is Number
    assert 0xA0 is Int
    assert 1.5 is Number
    assert 1.5 is not Int
    assert 2e-5 is Number
    assert 2e-5 is not Int
    assert 0/0 is not Number
    assert 0/0 is NaN
    assert 1/0 is not Number
    assert 1/0 is Infinite
}
