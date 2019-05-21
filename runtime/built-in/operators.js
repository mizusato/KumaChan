function lazy_bool (arg, desc, name) {
    ensure(is(arg, Types.Bool), 'arg_require_bool', name)
    return arg
}

let operators = {
    'is': f (
        'operator.is',
        'function operator.is (x: Any, T: Type) -> Bool',
            (x, A) => {
                // if T is a custom type, its checker must be f: [Any] --> Bool
                // this constraint should be enforced by custom type creator
                return call(A[Checker], [x])
            }
    ),
    'str': f (
        'operator.str',
        `function operator.str (s: StructOperand<'str'>) -> String`,
            s => apply_unary('str', s),
        'function operator.str (p: Bool) -> String',
            p => p? 'true': 'false',
        'function operator.str (x: Number) -> String',
            x => Number.prototype.toString.call(x),
        'function operator.str (s: String) -> String',
            s => s
    ),
    /* Pull, Push, Derive, Otherwise */
    '<<': f (
        'operator.pull',
        'function operator.pull (f: Callable, x: Any) -> Any',
            (f, x) => f(x),
        'function operator.pull (l: Hash, r: Hash) -> Hash',
            (l, r) => Object.assign(l, r),
        'function operator.pull (s: String, x: Any) -> String',
            (s, x) => call(string_format, [s, x])
    ),
    '>>': f (
        'operator.push',
        'function operator.push (l: Any, r: Any) -> Any',
            (l, r) => operators['<<'](r, l)
    ),
    '=>': f (
        'operator.derive',
        'function operator.derive (p: Bool, ok: Callable) -> Any',
            (p, ok) => p? ok(): Nil
    ),
    'or': f (
        'operator.otherwise',
        'function operator.otherwise (x: Any, fallback: Callable) -> Any',
            (x, fallback) => (x !== Nil)? x: fallback()
    ),
    /* Comparsion */
    '<': f (
        'operator.less_than',
        `function operator.less_than (s1: StructOperand<'<'>, s2: StructOperand<'<'>) -> Bool`,
            (s1, s2) => apply_operator('<', s1, s2),
        'function operator.less_than (a: String, b: String) -> Bool',
            (a, b) => a < b,
        'function operator.less_than (x: Number, y: Number) -> Bool',
            (x, y) => x < y
    ),
    '>': f (
        'operator.greater_than',
        'function operator.greater_than (l: Any, r: Any) -> Bool',
            (l, r) => operators['<'](r, l)
    ),
    '<=': f (
        'operator.less_than_or_equal',
        'function operator.less_than_or_equal (l: Any, r: Any) -> Bool',
            (l, r) => !operators['<'](r, l)
    ),
    '>=': f (
        'operator.greater_than_or_equal',
        'function operator.greater_than_or_equal (l: Any, r: Any) -> Bool',
            (l, r) => !operators['<'](l, r)
    ),
    '==': f (
        'operator.equal',
        `function operator.equal (s1: StructOperand<'=='>, s2: StructOperand<'=='>) -> Bool`,
            (s1, s2) => apply_operator('==', s1, s2),
        'function operator.equal (l: Any, r: Any) -> Bool',
            (l, r) => (l === r)
    ),
    '!=': f (
        'operator.not_equal',
        'function operator.not_equal (l: Any, r: Any) -> Bool',
            (l, r) => !operators['=='](l, r)
    ),
    // TODO: == can be overloaded by EquailityRedefined interface
    '===': f (
        'operator.original_equal',
        'function operator.original_equal (l: Any, r: Any) -> Bool',
            (l, r) => (l === r)
    ),
    '!==': f (
        'operator.original_not_equal',
        'function operator.original_not_equal (l: Any, r: Any) -> Bool',
            (l, r) => !operators['==='](l, r)
    ),
    /* Logic */
    '&&': f (
        'operator.and',
        'function operator.and (p: Bool, q: Callable) -> Bool',
            (p, q) => !p? false: lazy_bool(q(), 'operator.and', 'q')
    ),
    '||': f (
        'operator.or',
        'function operator.or (p: Bool, q: Callable) -> Bool',
            (p, q) => p? true: lazy_bool(q(), 'operator.or', 'q')
    ),
    '!': f (
        'operator.not',
        'function operator.not (p: Bool) -> Bool',
            p => !p
    ),
    '&': f (
        'operator.intersect',
        'function operator.intersect (A: Type, B: Type) -> Type',
            (A, B) => Ins(A, B)
    ),
    '|': f (
        'operator.union',
        'function operator.union (A: Type, B: Type) -> Type',
            (A, B) => Uni(A, B)
    ),
    '~': f (
        'operator.complement',
        'function operator.complement (A: Type) -> Type',
            A => Not(A)
    ),
    '\\': f (
        'operator.difference',
        'function operator.difference (A: Type, B: Type) -> Type',
            (A, B) => Ins(A, Not(B))
    ),
    'not': f (
        'operator.keyword_not',
        'function operator.keyword_not (p: Bool) -> Bool',
            p => !p,
        'function operator.keyword_not (A: Type) -> Type',
            A => Not(A)
    ),
    /* Arithmetic */
    '+': f (
        'operator.plus',
        `function operator.plus (s1: StructOperand<'+'>, s2: StructOperand<'+'>) -> Object`,
            (s1, s2) => apply_operator('+', s1, s2),
        'function operator.plus (a: Iterable, b: Iterable) -> Iterable',
            (a, b) => {
                return (function* ()  {
                    for (let I of a) { yield I }
                    for (let I of b) { yield I }
                })()
            },
        'function operator.plus (a: List, b: List) -> List',
            (a, b) => [...a, ...b],
        'function operator.plus (a: String, b: String) -> String',
            (a, b) => a + b,
        'function operator.plus (x: Number, y: Number) -> MayNotNumber',
            (x, y) => x + y
    ),
    '-': f (
        'operator.minus',
        `function operator.minus (s: StructOperand<'negate'>) -> Object`,
            s => apply_unary('negate', s),
        `function operator.minus (s1: StructOperand<'-'>, s2: StructOperand<'-'>) -> Object`,
            (s1, s2) => apply_operator('-', s1, s2),
        'function operator.minus (x: Number) -> Number',
            x => -x,
        'function operator.minus (x: Number, y: Number) -> MayNotNumber',
            (x, y) => x - y
    ),
    '*': f (
        'operator.times',
        `function operator.times (s1: StructOperand<'*'>, s2: StructOperand<'*'>) -> Object`,
            (s1, s2) => apply_operator('*', s1, s2),
        'function operator.times (x: Number, y: Number) -> MayNotNumber',
            (x, y) => x * y
    ),
    '/': f (
        'operator.divide',
        `function operator.divide (s1: StructOperand<'/'>, s2: StructOperand<'/'>) -> Object`,
            (s1, s2) => apply_operator('/', s1, s2),
        'function operator.divide (x: Number, y: Number) -> MayNotNumber',
            (x, y) => x / y
    ),
    '%': f (
        'operator.modulo',
        `function operator.modulo (s1: StructOperand<'%'>, s2: StructOperand<'%'>) -> Object`,
            (s1, s2) => apply_operator('%', s1, s2),
        'function operator.modulo (x: Number, y: Number) -> MayNotNumber',
            (x, y) => x % y
    ),
    '^': f (
        'operator.power',
        `function operator.power (s1: StructOperand<'^'>, s2: StructOperand<'^'>) -> Object`,
            (s1, s2) => apply_operator('^', s1, s2),
        'function operator.power (x: Number, y: Number) -> MayNotNumber',
            (x, y) => Math.pow(x, y)
    )
}

let operator_is = operators['is']

Object.freeze(operators)

function get_operator (name) {
    assert(has(name, operators))
    return operators[name]
}
