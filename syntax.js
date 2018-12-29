const Strings = {
    Blank: one_of('　', ' ', TAB, CR, LF)
}


const Repeat = Enum('once', 'more', 'maybe_once', 'maybe_more')


function Unit (char_set, repeat = '', next_char = Any, process = (x => x)) {
    let object = {
        char_set: char_set,
        next_char: next_char,
        process: process,
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
    repeat: Repeat,
    process: Function
}))


function Pattern (category, name, units) {
    check(Pattern, arguments, {
        category: Str, name: Str, units: ArrayOf(Unit)
    })
    assert(units.length > 0)
    return {
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
                    return 0
                }
            }
            let read_count = run_machine()
            let matched_string = cache.transform_by(chain(
                x => take_while(x, (_, index) => index < read_count),
                x => map(x, I => I.value.current),
                x => join(x, '')
            ))
            return matched_string
        }
    }
}


const num = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9']
const non_zero = ['1', '2', '3', '4', '5', '6', '7', '8', '9']
const int_pattern = Pattern('Number', 'Int', [
    Unit(one_of.apply(null, non_zero)),
    Unit(one_of.apply(null, num), '*')
])


function Matcher () {
    return {
        try_to_match: function (iterator) {
        }
    }
}


function StringProcessor (string) {
    let chars = map(string, x => x)
    let pos = 0
    let iterator = function* () {
        for (let i=pos; i<chars.length; i++) {
            yield chars[i]
        }
    }
    return {
        terminated: () => (pos == chars.length),
        match: function (matcher) {
            let match = matcher.try_to_match(iterator)
            pos += match.length
            return match
        }
    }
}


const TOKEN_ORDER = [
    'string',
    'extend_string',
    'comment',
    'space',
    'line_feed',
    'symbol',
    'identifier'
]


const TOKEN = {
    string: {
        pattern: /^'([^']*)'/,
        extract: 1
    },
    extend_string: {
        pattern: /^"[^"]*"/,
        extract: 1
    },
    comment: {
        pattern: /^\/\*(.*)\*\//,
        extract: 1
    },
    space: {
        pattern: /^[ \t\r　]+/,
        extract: 0
    },
    line_feed: {
        pattern: /^\n+/,
        extract: 0
    },
    symbol: {
        pattern: (
            /^(<<|>>|<=|>=|\&\&|\|\||[\+\-\*\/%^!~\&\|><=\{\}\[\]\(\)\.\,])/
        ),
        extract: 0
    },
    identifier: {
        pattern: (
/^[^0-9\~\&\- \t\r\n　\*\.\,\{\[\('"\)\]\}\/][^ \t\r\n　\*\.\,\{\[\('"\)\]\}\/]*/
        ),
        extract: 0
    }
}
