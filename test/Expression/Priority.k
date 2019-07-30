invoke {
    assert -abs(-5) + 10 == 5
    assert 2 + 5*3 == 17
    assert 1 - 6/2 == -2
    assert 2^10 / 2 == 512
    assert -2^3 + 8 == 0
    assert (5 % 3)^10 == 1024
    assert (1 + 1 < 3) == true
    assert (false && true == false) == false
    assert 1 + 1 == 2 && 2 + 2 == 4
    assert (false && true || !false) == true
    assert (abs << -5 + 3) == 2
    assert (ceil << 3 ^ 0.5) == 2
    assert (.{ .x } << 1 + 1 == 2 || false) == true
    assert (1 + 1 or 3) == 2
    assert (1 - 11*2 + 3*5 % 7 - 8/2 * 6^2) == -164
}
