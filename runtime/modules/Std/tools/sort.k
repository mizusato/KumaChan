function swap (a: Array, i: Index, j: Index) -> Void {
    if i == j { return }
    let t = a[i]
    set a[i] = a[j]
    set a[j] = t
}


function i_sort (a: Array, compare: Arity<2>) -> Void {
    let L = len(a)
    if L == 0 { return }
    if L == 1 { return }
    for i in range(1, L) {
        let e = a[i]
        for j in range(0, i) {
            let e_small = compare(e, a[j])
            ensure compare_valid { e_small is Bool }
            if e_small {
                var k = i
                while k > j {
                    set a[k] = a[k-1]
                    reset k -= 1
                }
                set a[j] = e
                break
            }
        }
    }
    @handle error {
        unless compare_valid {
            panic MSG.sort_compare_invalid()
        }
    }
}


function sort (a: Array, compare: Arity<2>) -> Void {
    let L = len(a)
    if L <= 5 {
        i_sort(a, compare)
        return
    }
    let p = __Math.floor(__Math.random() * L)
    let pivot = a[p]
    swap(a, p, 0)
    var m = 1
    for i in range(1, L) {
        let big = compare(pivot, a[i])
        ensure compare_valid { big is Bool }
        if not big {
            swap(a, i, m)
            reset m += 1
        }
    }
    swap(a, 0, m-1)
    let lo = a[:m-1]
    let hi = a[m:]
    sort(lo, compare)
    sort(hi, compare)
    @handle error {
        unless compare_valid {
            panic MSG.sort_compare_invalid()
        }
    }
}
