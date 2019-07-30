invoke {
    // inline lambda
    let f = .{ .x*.x }
    assert f is Function
    assert f is Arity<1>
    assert f(5) == 25
    let f1 = .{ .x*.y }
    assert f1 is Function
    assert f1 is Arity<2>
    assert f1(7,7) == 49
    let f2 = .{ x,y --> x^y }
    assert f2 is Arity<2>
    assert f2(2,10) == 1024
    let f3 = .{ _,y --> y }
    assert f3 is Arity<2>
    assert f3(7,11) == 11
    let f4 = .{ 59 }
    assert f4 is Arity<0>
    assert f4() == 59
    let f5 = .{ _ --> 59 }
    assert f5({}) == 59
    let f6 = .{ .{ .x + 1 } }
    assert f6()(4) == 5
    let f7 = .{ (.{ .y*5 })(.x) }
    assert f7(9) == 45
    let f8 = .{ x --> .{ x + .y } }
    assert f8('3')('5') == '35'
    assert f8('5')('3') == '53'
    let f9 = .{ .x + .y }
    assert f9('3', '5') == '35'
    assert f9('5', '3') == '53'
    // block lambda
    let g = lambda { return 59 }
    assert g is Function
    assert g is Arity<0>
    assert g() == 59
    let g1 = lambda -> Int { return 59 }
    assert g1() == 59
    let g2 = lambda() -> Int { return 59 }
    assert g2() == 59
    let g3 = lambda() { return 59 }
    assert g3() == 59
    let g4 = lambda -> { return 59 }
    assert g4() == 59
    let h = lambda(x: Int, y: Int) { return x + y }
    assert h is Function
    assert h is Arity<2>
    assert h(4,5) == 9
    let h1 = lambda (x: Int, y: Int) -> Int { return x * y }
    assert h is Arity<2>
    assert h1(4,5) == 20
    let h2 = lambda(x: String) -> { return x + '1' }
    assert h2 is Arity<1>
    assert h2('3') == '31'
}
