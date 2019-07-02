invoke {
    assert !true == false
    assert !false == true
    assert !true == not true
    assert !false == not false
    assert (true && false) == false
    assert (false && true) == false
    assert (true && true) == true
    assert (false && false) == false
    assert (true || false) == true
    assert (false || true) == true
    assert (true || true) == true
    assert (false || false) == false
}
