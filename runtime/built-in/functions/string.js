pour(built_in_functions, {
    utf8_size: fun (
        'function utf8_size (s: String) -> Size',
            s => fold(map(s, c => c.codePointAt(0)), 0, (code_point, sum) => {
                if ((code_point & 0xFFFFFF80) == 0) {
                    return sum + 1
                } else if ((code_point & 0xFFFFF800) == 0) {
                    return sum + 2
                } else if ((code_point & 0xFFFF0000) == 0) {
                    return sum + 3
                } else if ((code_point & 0xFFE00000) == 0) {
                    return sum + 4
                } else {
                    assert(false)
                }
            })
    ),
    ord: fun (
        'function ord (c: Char) -> Index',
            c => c.codePointAt(0)
    ),
    chr: fun (
        'function chr (i: Index) -> Char',
            i => {
                try {
                    return String.fromCodePoint(i)
                } catch (e) {
                    if (e instanceof RangeError) {
                        ensure(false, 'invalid_code_point', i.toString(16))
                    } else {
                        throw e
                    }
                }
            }
    ),
    to_lower_case: fun (
        'function to_lower_case (s: String) -> String',
            s => s.toLowerCase()
    ),
    to_upper_case: fun (
        'function to_upper_case (s: String) -> String',
            s => s.toUpperCase()
    ),
    match: fun (
        'function match (s: String, regexp: String) -> Maybe<List>',
            (s, regexp) => {
                let r = null
                try {
                    r = new RegExp(regexp, 'su')
                } catch (e) {
                    ensure(!(e instanceof SyntaxError), 'regexp_invalid')
                    throw e
                }
                let m = s.match(r)
                return (m != null)? list(m): Nil
            }
    ),
    match_all: fun (
        'function match_all (s: String, regexp: String) -> Iterator',
        (s, regexp) => (function* () {
            let r = null
            try {
                r = new RegExp(regexp, 'sgu')
            } catch (e) {
                ensure(!(e instanceof SyntaxError), 'regexp_invalid')
                throw e
            }
            let m = null
            while ((m = r.exec(s)) !== null) {
                yield list(m)
            }
        })()
    ),
    replace: fun (
        'function replace (s: String, regex: String, f: Function) -> String',
            (s, regexp, f) => {
                let r = null
                try {
                    r = new RegExp(regexp, 'sgu')
                } catch (e) {
                    ensure(!(e instanceof SyntaxError), 'regexp_invalid')
                    throw e
                }
                return s.replace(r, (...args) => {
                    let n = args.length - 3
                    let t = call(f, args.slice(0, args.length-2))
                    ensure(is(t, Types.String), 'replace_not_string')
                    return t
                })
            }
    )
})
