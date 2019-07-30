invoke {
    interface Box<T> {
        unwrap () -> T
    }
    class SimpleBox<T> is Box<T> {
        init (x: T) {
            do nothing
        }
        unwrap () -> T {
            return x
        }
    }
    assert SimpleBox<Int> ~~ SimpleBox<Int>
    assert SimpleBox<Int> is TypeType<SimpleBox>
    assert SimpleBox<Int> is Impl<Box<Int>>
    let b = SimpleBox<String>('b')
    assert b is SimpleBox
    assert b is SimpleBox<String>
    assert b is not SimpleBox<Int>
    assert b is Box
    assert b is Box<String>
    assert b is not Box<Int>
    assert b->unwrap() == 'b'
}
