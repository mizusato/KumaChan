invoke {
    type PositiveInteger = { x | x is Int, x > 0 }
    assert 1 is PositiveInteger
    assert -1 is not PositiveInteger
    assert 1.5 is not PositiveInteger
    assert PositiveInteger <=> PositiveInteger
}
