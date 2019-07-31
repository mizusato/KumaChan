invoke {
    function test () -> String {
        ensure ok { 1+1 == 3 }
        return 'impossible'
        ...
        handle error {
            unless ok {
                return 'caught'
            }
        }
    }
    assert test() == 'caught'
}
