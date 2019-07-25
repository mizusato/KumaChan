let FinalPattern = format({
    is_final: one_of(true),
    extract: Types.Any,
    target: Types.String,
    allow_nil: Types.Bool
})

let Pattern = Uni(FinalPattern, format({
    is_final: one_of(false),
    extract: Types.Any,
    allow_nil: Types.Bool,
    items: TypedList.of($(x => is(x, Pattern)))
}))

function match_pattern (scope, is_fixed, pattern, value) {
    assert(scope instanceof Scope)
    assert(is(is_fixed, Types.Bool))
    assert(is(pattern, Pattern))
    ensure(is(value, Uni(Nil, Types.GeneralGetter)), 'match_non_getter')
    let all_nil = false
    if (is(value, Types.Nil)) {
        if (pattern.allow_nil) {
            all_nil = true
        } else {
            ensure(false, 'match_nil')
        }
    }
    for (let item of pattern.items) {
        let v = all_nil? Nil: (
            call(get_data, [value, item.extract, item.allow_nil])
        )
        if (item.is_final) {
            scope.declare(item.target, v, is_fixed)
        } else {
            match_pattern(scope, is_fixed, item, v)
        }
    }
    return Void
}
