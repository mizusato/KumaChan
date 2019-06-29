invoke {
    function raise_error () -> Void {
        throw custom_error('oops')
    }
    function test () -> String {
        try to do_sth {
            raise_error()
        }
        return 'impossible'
        @handle error {
            failed to do_sth {
                return 'caught'
            }
        }
    }
    assert test() == 'caught'
}
