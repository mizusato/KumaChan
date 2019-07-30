invoke {
    let ob = observer {
        push 5
        push 7
        push Complete
    }
    assert ob is Observer
    assert ob is Observable
    var product = 1
    ob -> subscribe -> lambda(x) {
        reset product *= x
    }
    assert product == 35
}
