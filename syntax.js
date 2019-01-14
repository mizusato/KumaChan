'use strict';


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
    ForbiddenInId: one_of.apply({}, map(
        '`' + `\\'"()[]{}:,.!~&|+-*%^<>=`,
        // divide "/" is available 
    x => x))
}


Pattern.PrefixOperator = function (name, operator) {
    check(Pattern.PrefixOperator, arguments, {
        name: Str, operator: Str
    })
    return Pattern('Operator', name, map(operator, (char, index) => Unit(
        $1(char), '', (index < operator.length-1)? Any: Char.NotSpace
    )))
}


Pattern.Operator = function (name, operator) {
    check(Pattern.Operator, arguments, {
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
    Pattern('RawCode', 'RawCode', [
        Unit($1('/')),
        Unit($1('~')),
        CustomUnit((char, next) => `${char}${next}` != '~/', '*'),
        Unit($1('~')),
        Unit($1('/'))
    ]),
    Pattern('Comment', 'Comment', [
        Unit($1('/')),
        Unit($1('*')),
        CustomUnit((char, next) => `${char}${next}` != '*/', '*'),
        Unit($1('*')),
        Unit($1('/'))
    ]),
    Pattern('Blank', 'Space', [
        Unit($u(Char.Space, one_of(CR, TAB)), '+')
    ]),
    Pattern('Blank', 'Linefeed', [
        Unit($1(LF), '+')
    ]),
    Pattern.Operator('(', '('),
    Pattern.Operator(')', ')'),
    Pattern.Operator('[', '['),
    Pattern.Operator(']', ']'),
    Pattern.Operator('{', '{'),
    Pattern.Operator('}', '}'),
    Pattern.Operator(',', ','),
    Pattern.Operator(':', ':'),
    Pattern.Operator('.{', '.{'),
    Pattern.Operator('.[', '.['),
    Pattern.Operator('..', '..'),
    Pattern.Operator('.', '.'),
    Pattern.PrefixOperator('Not', '!'),
    Pattern.Operator('Or', '||'),
    Pattern.Operator('And', '&&'),
    Pattern.PrefixOperator('Complement', '~'),
    Pattern.Operator('Union', '|'),
    Pattern.Operator('Intersect', '&'),
    Pattern.Operator('->', '->'),
    Pattern.Operator('<-', '<-'),
    Pattern.PrefixOperator('Negative', '-'),
    Pattern.Operator('Minus', '-'),
    //Pattern.PrefixOperator('Positive', '+'),
    Pattern.Operator('Plus', '+'),
    Pattern.Operator('Times', '*'),
    Pattern.Operator('Over', '/'),
    //Pattern.PrefixOperator('Parameter', '%'),  // ugly, use dot
    Pattern.Operator('Modulo', '%'),
    Pattern.Operator('Power', '^'),
    Pattern.Operator('=', '='),
    Pattern.Operator('Equal', '=='),
    Pattern.Operator('NotEqual', '!='),
    Pattern.Operator('<<', '<<'),
    Pattern.Operator('>>', '>>'),
    Pattern.Operator('Less', '<'),
    Pattern.Operator('Greater', '>'),
    Pattern.Operator('LessEqual', '<='),
    Pattern.Operator('GreaterEqual', '>='),
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
        Unit($n(Char.NotDigit, Chat.NotSpace, $_(Char.ForbiddenInId)),
        Unit($n(Char.NotSpace, $_(Char.ForbiddenInId)), '*')
    ])
]


const SimpleOperand = $n(
    Token.Valid,
    $(token => token.matched.name.is(one_of(
        'Exponent', 'Float', 'Integer',
        'RawString', 'FormatString',
        'Identifier', 'Member'
    ))
))


const SimpleOperator = {
    Sentinel:     { type: 'N/A',    priority: -1,  assoc: 'left'  },
    Parameter:    { type: 'prefix', priority: 100, assoc: 'right' },
    Access:       { type: 'infix',  priority: 99,  assoc: 'left'  },
    Call:         { type: 'infix',  priority: 98,  assoc: 'left'  },
    Get:          { type: 'infix',  priority: 98,  assoc: 'left'  },
    Plus:         { type: 'infix',  priority: 50,  assoc: 'left'  },
    Negative:     { type: 'prefix', priority: 85,  assoc: 'right' },
    Minus:        { type: 'infix',  priority: 50,  assoc: 'left'  },
    Times:        { type: 'infix',  priority: 80,  assoc: 'left'  },
    Over:         { type: 'infix',  priority: 70,  assoc: 'left'  },
    Modulo:       { type: 'infix',  priority: 75,  assoc: 'left'  },
    Power:        { type: 'infix',  priority: 90,  assoc: 'right' },
    Equal:        { type: 'infix',  priority: 30,  assoc: 'left'  },
    NotEqual:     { type: 'infix',  priority: 20,  assoc: 'left'  },
    Greater:      { type: 'infix',  priority: 20,  assoc: 'left'  },
    GreaterEqual: { type: 'infix',  priority: 20,  assoc: 'left'  },
    Less:         { type: 'infix',  priority: 20,  assoc: 'left'  },
    LessEqual:    { type: 'infix',  priority: 20,  assoc: 'left'  },
    Not:          { type: 'prefix', priority: 85,  assoc: 'right' },
    And:          { type: 'infix',  priority: 40,  assoc: 'left'  },
    Or:           { type: 'infix',  priority: 30,  assoc: 'left'  },
    Complement:   { type: 'prefix', priority: 85,  assoc: 'right' },
    Intersect:    { type: 'infix',  priority: 40,  assoc: 'left'  },
    Union:        { type: 'infix',  priority: 30,  assoc: 'left'  }
}


SetEquivalent(SimpleOperator, $(
    token => SimpleOperator.has(token.matched.name)
))


const Syntax = mapval({
    
    Id: ['Identifier', 'RawString'],
    Concept: 'Identifier',
    
    Module: 'Program',
    Program: 'Command NextCommand',
    Command: [
        'RawCode',
        'FuncDef',
        'Let',
        'Return',
        'Assign',
        'MapExpr'
    ],
    NextCommand: ['Command NextCommand', ''],

    Let: '~let Id = Expr',
    Assign: 'LeftVal = Expr',
    LeftVal: 'Id MemberNext KeyNext',
    MemberNext: ['Access Member MemberNext', ''],
    KeyNext: ['Get Key KeyNext', ''],
    
    Return: '~return Expr',
    
    Expr: [
        'RawCode',
        'FuncExpr',
        'MapExpr'
    ],
    
    MapExpr: 'MapOperand MapNext',
    MapNext: [
        'MapOperator FuncExpr',
        'MapOperator MapOperand MapNext',
        ''
    ],
    MapOperator: [
        '->', '<-',
        '>>', '<<'
    ],
    MapOperand: [
        'Hash', 'HashLambda',
        'List', 'ListLambda',
        'SimpleLambda', 'Simple'
    ],
    
    ItemList: 'Item NextItem',
    Item: 'Expr',
    NextItem: [', Item NextItem', ''],
    
    PairList: 'Pair NextPair',
    Pair: 'Id : Expr',
    NextPair: [', Pair NextPair', ''],
    
    Hash: ['{ }', '{ PairList }'],
    List: ['[ ]', '[ ItemList ]'],
    
    HashLambda: '.{ PairList }',
    ListLambda: '.[ ItemList ]',
    
    SimpleLambda: [
        '.{ SimpleLambda }',
        '.{ Simple }'
    ],
    
    ParaList: ['( )', '( Para NextPara )'],
    Para: 'Concept PassFlag Id',
    PassFlag: ['Intersect', ''],  // Intersect: &
    NextPara: [', Para NextPara', ''],
    Target: ['-> Concept', ''],
    Body: '{ Program }',
    
    FuncFlag: ['~g :', '~f :', ''],
    FuncExpr: 'FuncFlag ParaList Target Body',
    
    Effect: ['~global', '~local'],
    FuncDef: 'Effect Id Call ParaList Target Body',
    //                   ↑  call operator will be inserted automatically
    
    Simple: { reducers: [ () => parse_simple ] },
    Wrapped: '( Simple )',
    Key: '[ Simple ]',
    ArgList: [
        '( )',
        '( KeyArg NextKeyArg )',
        '( Arg NextArg )'
    ],
    Arg: 'Simple',
    NextArg: [', Arg NextArg', ''],
    KeyArg: 'Id : Simple',
    NextKeyArg: [', KeyArg NextKeyArg', '']
    
}, function (rule) {
    let split = (string => string? string.split(' '): [])
    let derivations = (array => ({ derivations: array }))
    return transform(rule, [
        { when_it_is: Str, use: d => derivations([split(d)]) },
        { when_it_is: Array, use: r => derivations(map(r, d => split(d))) },
        { when_it_is: Otherwise, use: r => r }
    ])
})
