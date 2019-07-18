invoke {
    import * from Math
    function float_equal (a: Number, b: Number) -> Bool {
        return abs(a - b) < 1e-6
    }
    assert float_equal(E, exp(1))
    assert float_equal(PI, atan2(1,1)*4)
    assert float_equal(sqrt(2)^2, 2)
    assert float_equal(cbrt(-8), -2)
    assert float_equal(exp(10), E^10)
    assert float_equal(log(exp(10)), 10)
    assert float_equal(sin(PI/3), sqrt(3)/2)
    assert float_equal(cos(PI/3), 1/2)
    assert float_equal(tan(PI/3), sqrt(3))
    assert float_equal(PI/3, asin(sqrt(3)/2))
    assert float_equal(PI/3, acos(1/2))
    assert float_equal(PI/3, atan(sqrt(3)))
    assert float_equal(atan2(sqrt(3), 1), atan(sqrt(3)))
}
