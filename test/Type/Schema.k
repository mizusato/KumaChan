invoke {
    struct P2 {
        x: Int,
        y: Int
    }
    struct P3 {
        fields of P2,
        z: Int
    }
    let p3 = P3 { x: 3, y: 0, z: -2 }
    let p2 = P2(p3)
    assert p3 is P3
    assert p2 is P2
    assert p3 is not P2
    assert p2 is not P3
    set p3.z = 4
    // p3: (3, 0, 4)
    assert p3.z == 4
    set p2.y = -1
    // p3: (3, -1, 4)
    assert p2.y == -1
    assert p3.y == -1
    set p3.x = 1
    // p3: (1, -1, 4)
    assert p3.x == 1
    assert p2.x == 1
    var guard_called = false
    struct P {
        x: Number,
        y: Number
    }
    config {
        guard {
            assert fields is Hash
            assert fields.x is Number
            assert fields.y is Number
            reset guard_called = true
        }
        operator + (p1, p2) {
            return P { x: p1.x+p2.x, y: p1.y+p2.y }
        }
        operator str (p) {
            return '(${x}, ${y})' << p
        }
    }
    // p2: (1, -1)
    let p = P(p2)
    assert guard_called
    reset guard_called = false
    assert p is P
    assert p is not P2
    assert p2 is not P
    set p.y = 9
    // p, p2: (1, 9)
    assert guard_called
    reset guard_called = false
    assert p2.y == 9
    let a = p + P{ x: -1, y: 1 }
    assert guard_called
    reset guard_called = false
    assert a.x == 0
    assert a.y == 10
    assert str(a) == '(0, 10)'
    struct StringBox {
        value: String = '123'
    }
    let s = StringBox {}
    assert s.value == '123'
    set s.value = '456'
    assert s.value == '456'
    let s1 = StringBox { value: '789' }
    assert s1.value == '789'
    struct Box<T> {
        value: T
    }
    assert Box<Int> ~~ Box<Int>
    assert Box<Int> !~ Box<String>
    let b = Box<String>(s1)
    assert b.value == '789'
    assert b is not StringBox
    assert s1 is not Box<String>
    assert b is Box<String>
    assert b is Box
    let b1 = copy(b)
    assert b1 is Box<String>
    set b1.value = '000'
    assert b1.value == '000'
    assert b.value == '789'
}
