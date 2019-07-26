promise {
    import Source from Std
    let source = Source<Nil>(set_timeout(50))
    await value = create_promise << lambda(resolve, reject) {
        source -> subscribe -> .{ resolve(.x) }
    }
    assert value is Nil
    let source_sync = Source<Int>(observer {
        for i in [3,5,7] {
            push i
        }
    })
    var p = 1
    source_sync -> subscribe -> lambda(x) {
        reset p *= x
    }
    assert p == 3*5*7
}
