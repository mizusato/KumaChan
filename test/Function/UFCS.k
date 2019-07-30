invoke {
    let pair = .{ [.x,.y] }
    let p = 1 -> pair(2)
    assert p is List
    assert p[0] == 1 && p[1] == 2
    let double = .{ .x*2 }
    assert 3 -> double == 6
    invoke {
        let double = Number
        assert 5 -> double == 10
    }
}
