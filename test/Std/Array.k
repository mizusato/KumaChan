invoke {
    import Array from Std
    let a = Array<Int>([1,2,3,4])
    let b = a as Array<Number>
    set a[0] = -1
    // [-1, 2, 3, 4]
    assert a[0] == b[0]
    b -> append(5)
    assert len(b) == 5 && len(a) == 5 && a[4] == 5
    b -> pop()
    assert len(b) == 4 && len(a) == 4
    a -> prepend(-2)
    assert len(b) == 5 && len(a) == 5 && b[0] == -2
    a -> shift()
    assert len(b) == 4 && len(a) == 4
    a -> insert(1, 22)
    // [-1, 2, 22, 3, 4]
    b -> insert(4, 5)
    // [-1, 2, 22, 3, 4, 5]
    assert len(a) == 6 && len(b) == len(a)
    assert b[2] == 22 && a[5] == 5
    let check_sort = lambda (from, to) {
        let t = Array<Int>(iter(from))
        t -> sort -> .{ .x < .y }
        assert t == Array<Int>(iter(to))
    }
    check_sort([], [])
    check_sort([1], [1])
    check_sort([2,1], [1,2])
    check_sort (
        [4,1,5,72,42,61,66,88,999,123,44,55,6764,223,123],
        [1,4,5,42,44,55,61,66,72,88,123,123,223,999,6764]
    )
    assert Array<String>(['a','b']) != Array<String>(['a','b1'])
}
