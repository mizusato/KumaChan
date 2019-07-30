invoke {
    function f (x: Int) -> Int {
        return x + 1
    }
    function f (x: String) -> String {
        return x + '1'
    }
    assert f is not Function
    assert f is Overload
    assert f(3) == 4
    assert f('3') == '31'
    function g (x: Int, y: Int) -> Int {
        return x + y
    }
    function g (x: Int, y: Int) -> Int {
        return x * y
    }
    assert g(5,3) == 15
    invoke {
        function g (x: Int) -> Int {
            return x*x
        }
        assert g(7) == 49
        assert g(7,9) == 63
        invoke {
            let g = Nil
            assert g is Nil
        }
    }
}
