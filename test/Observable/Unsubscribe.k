promise {
    let clock = observer {
        var enabled = true
        promise {
            while enabled {
                await postpone(50)
                push Nil
            }
        }
        return lambda {
            reset enabled = false
        }
    }
    var n = 0
    let unsub = clock -> subscribe -> lambda {
        reset n += 1
    }
    await postpone(180)
    unsub()
    assert n == 3
    await postpone(120)
    assert n == 3
}
