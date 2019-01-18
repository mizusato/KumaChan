'use strict';


const Char = {
    Digit: Regex(/^[0-9]$/),
    NonZero: Regex(/^[1-9]$/),
    NotDigit: Regex(/^[^0-9]$/),
    Alphabet: Regex(/^[A-Za-z]$/),
    DoubleQuote: $1('"'),
    SingleQuote: $1("'"),
    Dot: $1('.'),
    Semicolon: $1(';'),
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
        name: Str, operator: Optional(Str)
    })
    operator = operator || name
    return Pattern('Operator', name, map(operator, (char, index) => Unit(
        $1(char), '', (index < operator.length-1)? Any: Char.NotSpace
    )))
}


Pattern.TextOperator = function (name, operator) {
    check(Pattern.PrefixOperator, arguments, {
        name: Str, operator: Optional(Str)
    })
    operator = operator || name
    return Pattern('Operator', name, map(operator, (char, index) => Unit(
        $1(char), '', (index < operator.length-1)? Any: Char.Space
    )))
}


Pattern.Operator = function (name, operator) {
    check(Pattern.Operator, arguments, {
        name: Str, operator: Optional(Str)
    })
    operator = operator || name
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
    Pattern('Blank', 'Linefeed', [
        Unit(Char.Semicolon, '+')
    ]),
    Pattern.Operator('('),
    Pattern.Operator(')'),
    Pattern.Operator('['),
    Pattern.Operator(']'),
    Pattern.Operator('{'),
    Pattern.Operator('}'),
    Pattern.Operator(','),
    Pattern.Operator(':'),
    Pattern.Operator('.{'),
    Pattern.Operator('..{'),
    Pattern.Operator('...{'),
    // Pattern.Operator('.['),
    Pattern.Operator('..['),
    Pattern.Operator('..'),  // ..Identifier = Inline Comment
    Pattern.Operator('.'),
    // name of simple operator starts with a capital alphabet
    Pattern.Operator('!='),
    Pattern.PrefixOperator('!'),
    Pattern.Operator('||'),
    Pattern.Operator('&&'),
    Pattern.PrefixOperator('~'),
    Pattern.Operator('|'),
    Pattern.Operator('&'),
    Pattern.Operator('->'),
    Pattern.Operator('<-'),
    Pattern.PrefixOperator('Negate', '-'),
    Pattern.Operator('-'),
    //Pattern.PrefixOperator('Positive', '+'),
    Pattern.Operator('+'),
    Pattern.Operator('*'),
    Pattern.Operator('/'),
    //Pattern.PrefixOperator('Parameter', '%'),  // ugly, use dot
    Pattern.Operator('%'),
    Pattern.Operator('^'),
    Pattern.Operator('=='),
    Pattern.Operator('='),
    Pattern.Operator('<<'),
    Pattern.Operator('>>'),
    Pattern.Operator('<'),
    Pattern.Operator('>'),
    Pattern.Operator('<='),
    Pattern.Operator('>='),
    Pattern.TextOperator('Is', 'is'),
    Pattern.TextOperator('~', 'not'),
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
        Unit($n(Char.NotDigit, Char.NotSpace, $_(Char.ForbiddenInId))),
        Unit($n(Char.NotSpace, $_(Char.ForbiddenInId)), '*')
    ])
]


const SimpleOperand = $n(
    Token.Valid,
    $(token => is(token.matched.name, one_of(
        'Exponent', 'Float', 'Integer',
        'RawString', 'FormatString',
        'Identifier', 'Member'
    ))
))


const SimpleOperator = {
    Sentinel:  { type: 'N/A',    priority: -1,  assoc: 'left'  },
    Parameter: { type: 'prefix', priority: 100, assoc: 'right' },
    Access:    { type: 'infix',  priority: 99,  assoc: 'left'  },
    Call:      { type: 'infix',  priority: 98,  assoc: 'left'  },
    Get:       { type: 'infix',  priority: 98,  assoc: 'left'  },
    Is:        { type: 'infix',  priority: 10,  assoc: 'left'  },
    Negate:    { type: 'prefix', priority: 85,  assoc: 'right' },
    '+':       { type: 'infix',  priority: 50,  assoc: 'left'  },
    '-':       { type: 'infix',  priority: 50,  assoc: 'left'  },
    '*':       { type: 'infix',  priority: 80,  assoc: 'left'  },
    '/':       { type: 'infix',  priority: 70,  assoc: 'left'  },
    '%':       { type: 'infix',  priority: 75,  assoc: 'left'  },
    '^':       { type: 'infix',  priority: 90,  assoc: 'right' },
    '==':      { type: 'infix',  priority: 30,  assoc: 'left'  },
    '!=':      { type: 'infix',  priority: 20,  assoc: 'left'  },
    '>':       { type: 'infix',  priority: 20,  assoc: 'left'  },
    '>=':      { type: 'infix',  priority: 20,  assoc: 'left'  },
    '<':       { type: 'infix',  priority: 20,  assoc: 'left'  },
    '<=':      { type: 'infix',  priority: 20,  assoc: 'left'  },
    '!':       { type: 'prefix', priority: 85,  assoc: 'right' },
    '&&':      { type: 'infix',  priority: 40,  assoc: 'left'  },
    '||':      { type: 'infix',  priority: 30,  assoc: 'left'  },
    '~':       { type: 'prefix', priority: 85,  assoc: 'right' },
    '&':       { type: 'infix',  priority: 40,  assoc: 'left'  },
    '|':       { type: 'infix',  priority: 30,  assoc: 'left'  }
}


SetEquivalent(SimpleOperator, $(
    token => has(SimpleOperator, token.matched.name)
))


const Syntax = mapval({
    
    Id: ['Identifier', 'RawString'],
    Constraint: 'Identifier',
    
    Module: 'Program',
    Program: 'Command NextCommand',
    Command: [
        'RawCode',
        'FunDef',
        'Let',
        'Return',
        'Assign',
        'Outer',
        'MapExpr'
    ],
    NextCommand: ['Command NextCommand', ''],

    Let: '~let Id = Expr',
    Assign: 'LeftVal = Expr',
    Outer: '~outer Id = Expr',
    LeftVal: 'Id MemberNext KeyNext',
    MemberNext: ['Access Member MemberNext', ''],
    KeyNext: ['Get Key KeyNext', ''],
    
    Return: '~return Expr',
    
    Expr: [
        'RawCode',
        'MapExpr',
    ],
    
    Concept: '{ Id That FilterList }',
    That: ['|', '~that'],
    FilterList: 'Filter NextFilter',
    NextFilter: [', Filter NextFilter', ''],
    Filter: 'Simple',
    
    Struct: '~struct Hash',
    
    MapExpr: 'MapOperand MapNext',
    MapNext: [
        'MapOperator MapOperand MapNext',
        ''
    ],
    MapOperator: [
        '->', '<-',
        '>>', '<<',
        '~to', '~by'
    ],
    MapOperand: [
        'Hash', 'HashLambda',
        'List', 'ListLambda',
        'SimpleLambda',
        'BodyLambda',
        'FunExpr',
        'Concept',
        'Struct',
        'Simple'
    ],
    
    ItemList: 'Item NextItem',
    Item: 'Expr',
    NextItem: [', Item NextItem', ''],
    
    PairList: 'Pair NextPair',
    Pair: 'Id : Expr',
    NextPair: [', Pair NextPair', ''],
    
    Hash: ['{ }', '{ PairList }'],
    List: [
        '[ ]', 'Get [ ]',  // Get operator may be inserted by get_tokens()
        '[ ItemList ]', 'Get [ ItemList ]'
    ],
    
    HashLambda: ['..{ }', '..{ PairList }'],
    ListLambda: ['..[ ]', '..[ ItemList ]'],
    
    SimpleLambda: [
        '.{ LambdaParaList SimpleLambda }',
        '.{ LambdaParaList Simple }'
    ],
    
    BodyLambda: '...{ Program }',
    
    LambdaParaList: [
        '( LambdaPara NextLambdaPara ) ->',
        'LambdaPara ->',
        ''
    ],
    LambdaPara: ['Identifier'],
    NextLambdaPara: [', LambdaPara NextLambdaPara', ''],
    
    ParaList: ['( )', '( Para NextPara )'],
    Para: 'Constraint PassFlag Id',
    NextPara: [', Para NextPara', ''],
    PassFlag: ['&', ''],
    Target: ['-> Constraint', '->', ''],
    Body: '{ Program }',
    
    FunFlag: ['~g :', '~h :', '~f :', '~global :', '~upper :', '~local :'],
    NoFlag: ['Call', ''],
    FunExpr: [
        'FunFlag ParaList Target Body',
        'NoFlag ParaList Target Body'
    ],
    
    Effect: ['~global', '~upper', '~local'],
    FunDef: 'Effect Id Call ParaList Target Body',
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
        { when_it_is: List, use: r => derivations(map(r, d => split(d))) },
        { when_it_is: Otherwise, use: r => r }
    ])
})
