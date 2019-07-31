invoke {
    let C = 700
    function get_value (need_error: Bool) -> Int {
        if need_error {
            throw custom_error('oops')
        } else {
            return C
        }
    }
    function test (value: Int, need_error: Bool) -> Int {
        ensure positive { value > 0 }
        ensure odd { value % 2 == 1 }
        try to get_v1 {
            let value1 = get_value(false)
        }
        try to get_v2 {
            let value2 = get_value(need_error)
        }
        return value + value1 + value2
        ...
        handle error {
            unless positive {
                return -1
            }
            unless odd {
                return -2
            }
            failed to get_v1 {
                return -3
            }
            failed to get_v2 {
                return -4
            }
        }
    }
    assert test(-1, false) == -1
    assert test(-1, true) == -1
    assert test(6, false) == -2
    assert test(6, true) == -2
    assert test(7, false) == 2*C + 7
    assert test(7, true) == -4
}
