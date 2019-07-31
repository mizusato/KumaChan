invoke {
    function test (need_error: Bool) -> Int {
        ensure ok { !need_error }
        return 1
        ...
        handle error {
            unless ok {
                return 2
            }
        }
        finally {
            return 3
        }
    }
    assert test(true) == 3
    assert test(false) == 3
}
