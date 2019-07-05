invoke {
    // Type 1
    struct Vector {
        x: Number,
        y: Number
    }
    config {
        operator as (v, T) {
            assert T ~~ Pair
            return Pair(v.x, v.y)
        }
    }
    // Type2
    class Pair {
        init (a: Any, b: Any) {
            do nothing
        }
        get_a () -> Object {
            return a
        }
        get_b () -> Object {
            return b
        }
        operator as (p, T) {
            assert T ~~ Vector
            assert p->get_a() is Number
            assert p->get_b() is Number
            return Vector { x: p->get_a(), y: p->get_b() }
        }
    }
    // Cast between Type1 and Type2
    let v = Vector { x: 1, y: 3 }
    let vp = v as Pair
    assert vp -> get_a == 1
    assert vp -> get_b == 3
    let pv = vp as Vector
    assert pv.x == 1
    assert pv.y == 3
}
