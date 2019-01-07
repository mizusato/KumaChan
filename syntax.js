const Char = {
    Digit: Regex(/^[0-9]$/),
    NonZero: Regex(/^[1-9]$/),
    NotDigit: Regex(/^[^0-9]$/),
    Alphabet: Regex(/^[A-Za-z]$/),
    DoubleQuote: $1('"'),
    SingleQuote: $1("'"),
    Dot: $1('.'),
    Space: one_of('　', ' '),
    NotSpace: $_(one_of('　', ' ')),
    Any: $u(Str, $(s => s.length >= 1)),
    CompactOperator: one_of.apply({}, map('()[]{}:,.*^', x=>x))
}


Pattern.PrefixOperator = function (name, operator) {
    check(Pattern.PrefixOperator, arguments, {
        name: Str, operator: Str
    })
    return Pattern('Operator', name, map(operator, (char, index) => Unit(
        $1(char), '', (index < operator.length-1)? Any: Char.NotSpace
    )))
}


Pattern.InfixOperator = function (name, operator) {
    check(Pattern.InfixOperator, arguments, {
        name: Str, operator: Str
    })
    return Pattern('Operator', name, map(operator, (char, index) => Unit(
        $1(char), '', (index < operator.length-1)? Any: Char.Space
    )))
}


Pattern.CompactOperator = function (name, operator) {
    check(Pattern.InfixOperator, arguments, {
        name: Str, operator: Str
    })
    return Pattern('Operator', name, map(operator, (char, index) => Unit(
        $1(char)
    )))
}


const Tokens = [
    Pattern('String', 'RawString', [
        Unit(Char.SingleQuote),
        Unit($_(Char.SingleQuote), '*'),
        Unit(Char.SingleQuote)
    ]),
    Pattern('String', 'FormatString', [
        Unit(Char.DoubleQuote),
        Unit($_(Char.DoubleQuote), '*'),
        Unit(Char.DoubleQuote)
    ]),
    Pattern('Comment', 'Comment', [
        Unit($1('/')),
        Unit($1('*')),
        Unit($_($1('/')), '*'),
        Unit($1('*')),
        Unit($1('/'))
    ]),
    Pattern('Blank', 'Space', [
        Unit($u(Char.Space, one_of(CR, TAB)), '+')
    ]),
    Pattern('Blank', 'Linefeed', [
        Unit($1(LF), '+')
    ]),
    /**
     * chars used by compact operators must be
     * registered at Char.CompactOperator
     */
    Pattern.CompactOperator('(', '('),
    Pattern.CompactOperator(')', ')'),
    Pattern.CompactOperator('[', '['),
    Pattern.CompactOperator(']', ']'),
    Pattern.CompactOperator('{', '{'),
    Pattern.CompactOperator('}', '}'),
    Pattern.CompactOperator(',', ','),
    Pattern.CompactOperator(':', ':'),
    Pattern.CompactOperator('.', '.'),
    Pattern.PrefixOperator('Not', '!'),
    Pattern.InfixOperator('Or', '||'),
    Pattern.InfixOperator('And', '&&'),
    Pattern.PrefixOperator('Complement', '~'),
    Pattern.InfixOperator('Union', '|'),
    Pattern.InfixOperator('Intersect', '&'),
    Pattern.PrefixOperator('Negative', '-'),
    Pattern.InfixOperator('Minus', '-'),
    Pattern.PrefixOperator('Positive', '+'),
    Pattern.InfixOperator('Plus', '+'),
    Pattern.CompactOperator('Times', '*'),
    Pattern.InfixOperator('Over', '/'),
    //Pattern.PrefixOperator('Parameter', '%'),  // ugly, use dot
    Pattern.InfixOperator('Modulo', '%'),
    Pattern.CompactOperator('Power', '^'),
    Pattern.InfixOperator('Assign', '='),
    Pattern.InfixOperator('Equal', '=='),
    Pattern.InfixOperator('NotEqual', '!='),
    Pattern.InfixOperator('LessThan', '<'),
    Pattern.InfixOperator('GreaterThan', '>'),
    Pattern.InfixOperator('LessThanOrEqual', '<='),
    Pattern.InfixOperator('GreaterThanOrEqual', '>='),
    Pattern.InfixOperator('PushLeft', '<<'),
    Pattern.InfixOperator('PushRight', '>>'),
    Pattern('Number', 'Exponent', [
        Unit(Char.Digit, '+'),
        Unit(Char.Dot),
        Unit(Char.Digit, '+'),
        Unit(one_of('E', 'e')),
        Unit(one_of('+', '-'), '?'),
        Unit(Char.Digit, '+')
    ]),
    Pattern('Number', 'Float', [
        Unit(Char.Digit, '+'),
        Unit(Char.Dot),
        Unit(Char.Digit, '+')        
    ]),
    Pattern('Number', 'Integer', [
        Unit(Char.Digit, '+')
    ]),
    Pattern('Name', 'Name', [
        Unit(Char.NotDigit),
        Unit($n(Char.NotSpace, $_(Char.CompactOperator)), '*')
    ])
]


const SimpleOperandInfo = {
    Exponent: { type: 'float' },
    Float: { type: 'float' },
    Integer: { type: 'integer' },
    Identifier: { type: 'identifier' },
    Member: { type: 'member' },
    RawString: { type: 'string-raw' },
    FormatString: { type: 'string-format' },
    // argument list (...) is also an oprand
    argument: { type: 'argument' },
    // key of get operation [...] is also an oprand
    key: { type: 'key' }
}


const SimpleOperatorInfo = {
    '(': { type: 'parentheses', which: 'left' },
    ')': { type: 'parentheses', which: 'right' },
    '[': { type: 'parentheses', which: 'left' },
    ']': { type: 'parentheses', which: 'right' },
    Parameter: { type: 'prefix' },
    Access: { type: 'infix', priority: 99, assoc: 'left' },
    Call: { type: 'infix', priority: 95, assoc: 'left' },
    Get: { type: 'infix', priority: 95, assoc: 'left' },
    Plus: { type: 'infix', priority: 50, assoc: 'left' },
    Negative: { type: 'prefix' },
    Minus: { type: 'infix', priority: 50, assoc: 'left' },
    Times: { type: 'infix', priority: 80, assoc: 'left' },
    Over: { type: 'infix', priority: 70, assoc: 'left' },
    Modulo: { type: 'infix', priority: 75, assoc: 'left' },
    Power: { type: 'infix', priority: 85, assoc: 'right' },
    Equal: { type: 'infix', priority: 30, assoc: 'left' },
    NotEqual: { type: 'infix', priority: 20, assoc: 'left' },
    GreaterThan: { type: 'infix', priority: 20, assoc: 'left' },
    GreaterThanOrEqual: { type: 'infix', priority: 20, assoc: 'left' },
    LessThan: { type: 'infix', priority: 20, assoc: 'left' },
    LessThanOrEqual: { type: 'infix', priority: 20, assoc: 'left' },
    Not: { type: 'prefix' },
    And: { type: 'infix', priority: 40, assoc: 'left' },
    Or: { type: 'infix', priority: 30, assoc: 'left' },
    Complement: { type: 'prefix' },
    Intersect: { type: 'infix', priority: 40, assoc: 'left' },
    Union: { type: 'infix', priority: 30, assoc: 'left' }
}


const Syntax = {
    Program: {
        derivations: [ [ 'Simple' ] ]
    },
    Simple: {
        derivations: [ [ 'Identifier', 'SimpleNext' ] ]
    },
    SimpleNext: {
        derivations: [
            [ 'Plus', 'Identifier', 'SimpleNext' ],
            [ 'Times', 'Identifier', 'SimpleNext' ],
            []
        ]
    },
    ArgumentList: {
        reducers: [
            tokens => {}
        ]
    },
    /*
    {
        name: 'Simple',
        reducers: [
            {
                condition: iterator => true,
                operation: iterator => tokens
            }
        ]
    },
    */
    /*
    {
        name: 'Argument',
        derviations: [
            [ '(', 'Simple', 'NextArg', ')' ]
        ]
    },
    {
        name: 'NextArg',
        derivations: [
            [ ',', 'Simple', 'NextArg' ],
            []
        ]
    }
    */
}
