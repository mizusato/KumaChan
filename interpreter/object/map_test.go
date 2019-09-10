package object

import "testing"
import "strconv"

func Check(t *testing.T, condition bool) {
    if !condition {
        t.Fail()
    }
}

func TestStrMapGetSet(t *testing.T) {
    var d = NewStrMap()
    d.Set("a", NewInt(1))
    d.Set("b", NewInt(2))
    d.Set("c", NewInt(3))
    Check(t, UnwrapInt(d.Get("a")) == 1)
    Check(t, UnwrapInt(d.Get("b")) == 2)
    Check(t, UnwrapInt(d.Get("c")) == 3)
}

func TestStrMapMiscOperations(t *testing.T) {
    var d = NewStrMap()
    Check(t, d.Size() == 0)
    d.Set("a", NewInt(1))
    Check(t, d.Size() == 1)
    d.Set("b", NewInt(2))
    d.Set("c", NewInt(3))
    Check(t, d.Has("c"))
    d.Delete("c")
    Check(t, !d.Has("c"))
    Check(t, d.Size() == 2)
    d.Set("d", NewInt(4))
    d.Set("d", NewInt(5))
    Check(t, UnwrapInt(d.Get("d")) == 5)
    Check(t, d.Size() == 3)
}

func TestStrMapShrink (t *testing.T) {
    var d = NewStrMap()
    var N = 10000
    for i := 0; i < N; i++ {
        d.Set(strconv.Itoa(i), NewInt(-i))
    }
    for i := 0; i < N-1; i++ {
        d.Delete(strconv.Itoa(i))
    }
    Check(t, d.Size() == 1)
    Check(t, UnwrapInt(d.Get(strconv.Itoa(N-1))) == -(N-1))
}

func TestIntMapGetSet(t *testing.T) {
    var m = NewIntMap()
    m.Set(1, NewString("a"))
    m.Set(2, NewString("b"))
    m.Set(3, NewString("c"))
    Check(t, UnwrapString(m.Get(1)) == "a")
    Check(t, UnwrapString(m.Get(2)) == "b")
    Check(t, UnwrapString(m.Get(3)) == "c")
}

func TestIntMapMiscOperations(t *testing.T) {
    var m = NewIntMap()
    Check(t, m.Size() == 0)
    m.Set(1, NewString("a"))
    Check(t, m.Size() == 1)
    m.Set(2, NewString("b"))
    m.Set(3, NewString("c"))
    Check(t, m.Has(3))
    m.Delete(3)
    Check(t, !m.Has(3))
    Check(t, m.Size() == 2)
    m.Set(4, NewString("d"))
    m.Set(4, NewString("e"))
    Check(t, UnwrapString(m.Get(4)) == "e")
    Check(t, m.Size() == 3)
}

func TestIntMapShrink (t *testing.T) {
    var m = NewIntMap()
    var N = 10000
    for i := 0; i < N; i++ {
        m.Set(i, NewString(strconv.Itoa(-i)))
    }
    for i := 0; i < N-1; i++ {
        m.Delete(i)
    }
    Check(t, m.Size() == 1)
    Check(t, UnwrapString(m.Get(N-1)) == strconv.Itoa(-(N-1)))
}
