class ScanningError extends Error {}


const Repeat = Enum('once', 'more', 'maybe_once', 'maybe_more')


function Unit (char_set, repeat = '', next_char = Any) {
    let object = {
        char_set: char_set,
        next_char: next_char,
        repeat: ({
            '': 'once',
            '?': 'maybe_once',
            '+': 'more',
            '*': 'maybe_more'
        })[repeat]
    }
    assert(object.is(Unit))
    return object
}


SetEquivalent(Unit, Struct({
    char_set: Concept,
    next_char: Concept,
    repeat: Repeat
}))


function Pattern (category, name, units) {
    check(Pattern, arguments, {
        category: Str, name: Str, units: ArrayOf(Unit)
    })
    assert(units.length > 0)
    return {
        maker: Pattern,
        category: category,
        name: name,
        match: function (iterable) {
            assert(iterable.is(Iterable))
            let iterator = lookahead(iterable, '')
            let repeat = val => $(unit => unit.repeat == val)
            let links = map(units, unit => transform(unit, [
                { when_it_is: repeat('once'), use: unit => ([
                    { to: 'next', unit: unit }                        
                ])},
                { when_it_is: repeat('maybe_once'), use: unit => ([
                    { to: 'next', unit: unit },
                    { to: 'next', unit: null }
                ])},
                { when_it_is: repeat('more'), use: unit => ([
                    { to: 'self', unit: unit },
                    { to: 'next', unit: unit }
                ])},
                { when_it_is: repeat('maybe_more'), use: unit => ([
                    { to: 'self', unit: unit },
                    { to: 'next', unit: null }
                ])}
            ]))
            function check_ok (unit, I) {
                if (I.done) { return false }
                let char = I.value.current
                let next = I.value.next
                let char_ok = unit.char_set.contains(char)
                let next_ok = unit.next_char.contains(next)
                let ok = char_ok && next_ok
                return ok
            }
            let cache = []
            function get_at (n) {
                if (n < cache.length) {
                    return cache[n]
                } else {
                    for (let i=cache.length; i<=n; i++) {
                        cache[i] = iterator.next()
                    }
                    return cache[n]
                }
            }
            let FINAL = links.length
            function run_machine (state = 0, count = 0) {
                if (state == FINAL) {
                    return count
                }
                let I = get_at(count)
                let r = find(map_lazy(links[state], function (link) {
                    let target = ({
                        self: state,
                        next: (state + 1)
                    })[link.to]
                    if ( link.unit != null ) {
                        let ok = check_ok(link.unit, I)
                        return ok? run_machine(target, count+1): null
                    } else {
                        return run_machine(target, count)
                    }
                }), x => x != null)
                if ( r != NotFound ) {
                    return r
                } else {
                    return null
                }
            }
            let read_count = run_machine()
            if (read_count == null) {
                return null
            }
            let matched_string = cache.transform_by(chain(
                x => take_while(x, (_, index) => index < read_count),
                x => map(x, I => I.value.current),
                x => join(x, '')
            ))
            return matched_string
        }
    }
}


SetMakerConcept(Pattern)


const Token = name => Struct({
    matched: Struct({
        name: $1(name)
    })
})


pour(Token, {
    Null: $(x => x === Token.Null),
    Valid: Struct({
        position: Struct({ row: Num, col: Num }),
        matched: Struct({
            category: Str,
            name: Str,
            string: Str
        })
    })
})


SetEquivalent(Token, $u(Token.Null, Token.Valid))


function Matcher (patterns) {
    return {
        match: function (iterator) {
            let match = find(map_lazy(patterns,
                pattern => ({
                    category: pattern.category,
                    name: pattern.name,
                    string: pattern.match(iterator())                    
                })
            ), (x => x.string != null))
            return (match == NotFound)? null: match
        }
    }
}


function CodeScanner (string) {
    let chars = map(string, x => x)
    let info = fold(chars, [], function (char, info) {
        let prev = info[info.length-1] || { row: 1, col: 0 }
        info.push(transform(char, [
            { when_it_is: $1(LF),
              use: () => ({ row: (prev.row + 1), col: 0 }) },
            { when_it_is: Otherwise,
              use: () => ({ row: prev.row, col: (prev.col + 1) }) },
        ]))
        return info
    })
    let pos = 0
    let iterator = function* () {
        for (let i=pos; i<chars.length; i++) {
            yield chars[i]
        }
    }
    return {
        match: function* (matcher) {
            while (pos < chars.length) {
                let match = matcher.match(iterator)
                if (match != null) {
                    yield {
                        position: info[pos],
                        matched: match
                    }
                    pos += match.string.length
                } else {
                    let p = info[pos]
                    throw new ScanningError(
                        `unable to match at row ${p.row} column ${p.col}`
                    )
                }
            }
        }
    }
}
