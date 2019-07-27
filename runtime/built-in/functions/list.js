pour(built_in_functions, {
    first: fun (
        'function first (l: List) -> Object',
            l => (ensure(l.length > 0, 'empty_list'), l[0])
    ),
    last: fun (
        'function last (l: List) -> Object',
            l => (ensure(l.length > 0, 'empty_list'), l[l.length-1])
    ),
    prepend: fun (
        'function prepend (l: List, item: Any) -> Void',
            (l, item) => {
                l.unshift(item)
                return Void
            }
    ),
    append: fun (
        'function append (l: List, item: Any) -> Void',
            (l, item) => {
                l.push(item)
                return Void
            }
    ),
    push: fun (
        'function push (l: List, item: Any) -> Void',
            (l, item) => {
                l.push(item)
                return Void
            }
    ),
    shift: fun (
        'function shift (l: List) -> Void',
            l => {
                ensure(l.length > 0, 'empty_list')
                l.shift()
                return Void
            }
    ),
    pop: fun (
        'function pop (l: List) -> Void',
            l => {
                ensure(l.length > 0, 'empty_list')
                l.pop()
                return Void
            }
    ),
    splice: fun (
        'function splice (l: List, i: Index, amount: Size) -> Void',
            (l, i, amount) => {
                ensure(i < l.length, 'index_error', i)
                ensure(i+amount <= l.length, 'invalid_splice', amount)
                l.splice(i, amount)
                return Void
            }
    ),
    insert: fun (
        'function insert (l: List, i: Index, item: Any) -> Void',
            (l, i, item) => {
                ensure(i < l.length, 'index_error', i)
                if (i == l.length-1) {
                    l.push(item)
                    return Void
                }
                let target = i+1
                l.push(l[l.length-1])
                for (let j=l.length-1; j>target; j--) {
                    l[j] = l[j-1]
                }
                l[target] = item
                return Void
            }
    ),
    index_of: fun (
        'function index_of (l: List, f: Arity<1>) -> Maybe<Index>',
            (l, f) => {
                for (let i = 0; i < l.length; i += 1) {
                    let c = call(f, [l[i]])
                    ensure(is(c, Types.Bool), 'cond_not_bool')
                    if (c) {
                        return i
                    }
                }
                return Nil
            }
    )
})
