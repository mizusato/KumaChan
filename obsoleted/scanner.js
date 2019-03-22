'use strict';


class ScanningError extends Error {}


const Repeat = Enum('once', 'more', 'maybe_once', 'maybe_more')
const RepeatFlagValue = {
    '': 'once',
    '?': 'maybe_once',
    '+': 'more',
    '*': 'maybe_more'
}


function Unit (char_set, repeat = '', next_char = Any) {
    let object = {
        char_set: char_set,
        next_char: next_char,
        repeat: RepeatFlagValue[repeat]
    }
    assert(is(object, Unit))
    return object
}


function CustomUnit (checker, repeat) {
    let object = {
        checker: checker,
        repeat: RepeatFlagValue[repeat]
    }
    assert(is(object, Unit))
    return object    
}


SetEquivalent(Unit, $u(
    Struct({
        char_set: Concept,
        next_char: Concept,
        repeat: Repeat
    }),
    Struct({
        checker: Fun,
        repeat: Repeat
    })
))


function Pattern (category, name, units) {
    check(Pattern, arguments, {
        category: Str, name: Str, units: ListOf(Unit)
    })
    assert(units.length > 0)
    return {
        maker: Pattern,
        category: category,
        name: name,
        match: function (iterable) {
            assert(is(iterable, Iterable))
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
                let third = I.value.third
                if (has(unit, 'checker')) {
                    return unit.checker(char, next, third)
                } else {
                    let char_ok = unit.char_set.contains(char)
                    let next_ok = unit.next_char.contains(next)
                    let ok = char_ok && next_ok
                    return ok
                }
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
            let matched_string = apply_on(cache, chain(
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
        position: Struct({
            row: Int,
            col: Int
        }),
        matched: Struct({
            category: Str,
            name: Str,
            string: Str
        })
    }),
    Operator: $n(
        $(x => is(x, Token.Valid)),
        $(t => t.matched.category == 'Operator')
    ),
    create_from: function (token, category, name, string) {
        check(Token.create_from, arguments, {
            token: Token.Valid,
            category: Str,
            name: Str,
            string: Optional(Str)
        })
        return {
            position: token.position,
            matched: {
                category: category,
                name: name,
                string: string || token.matched.string
            }
        }
    },
    pos_compare: function (x, y) {
        let key_order = ['row', 'col']
        return (function compare (i) {
            let L = x[key_order[i]]
            let R = y[key_order[i]]
            if (L != R) {
                return (L - R)
            } else {
                if (i+1 < key_order.length) {
                    return compare(i+1)
                } else {
                    return 0
                }
            }
        })(0)
    }
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