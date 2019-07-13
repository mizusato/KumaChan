struct ArrayData {
    get: $<Index><Object>,
    set: $<Index, Object><Void>,
    length: $<><Size>,
    prepend: Maybe<$<Object><Void>> = Nil,
    append: Maybe<$<Object><Void>> = Nil,
    shift: Maybe<$<><Void>> = Nil,
    pop: Maybe<$<><Void>> = Nil
}


class Array<T> is Getter<Index, T>, Setter<Index, T>, SliceGetter {
    init (data: ArrayData) {
        reset data = copy(data)
    }
    create (i: Iterable) {
        let list = [ x, for x in i ]
        return Array<T>(list)
    }
    create (l: List) {
        for v[i] in l {
            ensure type_correct(i) { v is T }
        }
        return Array<T>(ArrayData {
            get: lambda (i: Index) -> Object {
                return l[i]
            },
            set: lambda (i: Index, v: Object) -> Void {
                set l[i] = v
            },
            length: lambda -> Size {
                return len(l)
            },
            prepend: lambda (v: Object) -> Void {
                l -> prepend(v)
            },
            append: lambda (v: Object) -> Void {
                l -> append(v)
            },
            shift: lambda -> Void {
                l -> shift()
            },
            pop: lambda -> Void {
                l -> pop()
            }
        })
        @handle error {
            unless type_correct (i) {
                panic MSG.array_invalid_init(i)
            }
        }
    }
    unwrap () -> ArrayData {
        return copy(data)
    }
    get (i: Index, nf: Bool) -> Maybe<T> {
        let L = data.length()
        if not (i < L) {
            ensure index_valid { nf }
            return Nil
        }
        let element = data.get(i)
        ensure consistent { element is T }
        return element
        @handle error {
            unless index_valid {
                panic MSG.array_invalid_index(i, L)
            }
            unless consistent {
                panic MSG.array_inconsistent(i)
            }
        }
    }
    set (i: Index, value: T) -> Void {
        let L = data.length()
        ensure index_valid { i < L }
        data.set(i, value)
        @handle error {
            unless index_valid {
                panic MSG.array_invalid_index(i, L)
            }
        }
    }
    slice (lo: SliceIndex, hi: SliceIndex) -> Object {
        let L = data.length()
        let Default = SliceIndexDefault
        if lo is Default { reset lo = 0 }
        if hi is Default { reset hi = L }
        ensure slice_valid { lo <= hi }
        ensure index_valid { hi <= L }
        let check_consistent = lambda {
            if data.length() != L {
                panic MSG.array_slice_inconsistent()
            }
        }
        return Array<T>(ArrayData {
            get: lambda (i: Index) -> Object {
                check_consistent()
                return data.get(lo + i)
            },
            set: lambda (i: Index, v: Object) -> Void {
                check_consistent()
                data.set((lo + i), v)
            },
            length: lambda -> Size {
                check_consistent()
                return (hi - lo)
            }
        })
        @handle error {
            unless slice_valid {
                panic MSG.array_invalid_slice()
            }
            unless index_valid {
                panic MSG.array_invalid_slice_index(hi, L)
            }
        }
    }
    length () -> Size {
        return data.length()
    }
    append (value: T) -> Void {
        ensure supported { data.append is not Nil }
        data.append(value)
        @handle error {
            unless supported {
                panic MSG.array_unsupported('append()')
            }
        }
    }
    prepend (value: T) -> Void {
        ensure supported { data.prepend is not Nil }
        data.prepend(value)
        @handle error {
            unless supported {
                panic MSG.array_unsupported('prepend()')
            }
        }
    }
    shift () -> Void {
        ensure supported { data.shift is not Nil }
        data.shift()
        @handle error {
            unless supported {
                panic MSG.array_unsupported('shift()')
            }
        }
    }
    pop () -> Void {
        ensure supported { data.pop is not Nil }
        data.pop()
        @handle error {
            unless supported {
                panic MSG.array_unsupported('pop()')
            }
        }
    }
    sort (compare: Arity<2>) -> Void {
        i_sort(self, compare)
    }
    operator len (array) {
        return array -> length()
    }
    operator iter (array) {
        return iterator {
            for i in seq(len(array)) {
                yield array[i]
            }
        }
    }
    operator + (a1, a2) {
        return Array<T>((iter a1) + (iter a2))
    }
    operator == (a1, a2) {
        let T = get_class(a1).ElementType
        assert T <=> get_class(a2).ElementType
        ensure type_valid { T is EqualityDefined }
        let L = len(a1)
        if L != len(a2) {
            return false
        }
        for i in seq(L) {
            if a1[i] != a2[i] {
                return false
            }
        }
        return true
        @handle error {
            unless type_valid {
                panic MSG.array_equal_type_invalid()
            }
        }
    }
    operator as (self, T) {
        ensure cast_valid { T is TypeType<Array> }
        let E = T.ElementType
        for value[i] in self {
            ensure type_correct(i) { value is E }
        }
        return Array<E>(self -> unwrap())
        @handle error {
            unless cast_valid {
                panic MSG.array_cast_invalid()
            }
            unless type_correct (i) {
                panic MSG.array_cast_type_incorrect(i)
            }
        }
    }
    operator copy (array) {
        let E = get_class(array).ElementType
        return Array<E>(iter(array))
    }
    data {
        ElementType: T
    }
}


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

function qi_sort (a: Array, compare: Arity<2>) -> Void {
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
    qi_sort(lo, compare)
    qi_sort(hi, compare)
    @handle error {
        unless compare_valid {
            panic MSG.sort_compare_invalid()
        }
    }
}
