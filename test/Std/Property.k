invoke {
    import Property, PropertyCollection from Std
    class A is PropertyCollection {
        init (ro: Int, rw: Int) {
            mount PropertyCollection([
                Property { name: 'ro', get: .{ ro } },
                Property { name: 'rw', get: .{ rw }, set: lambda(v) -> Void {
                    reset rw = v
                } }
            ])
        }
    }
    let a = A(1, 2)
    assert a.ro == 1
    assert a.rw == 2
    set a.rw = 9
    assert a.rw == 9
    assert a.covfefe? is Nil
}
