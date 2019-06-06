function lazy_bool (arg, desc, name) {
    ensure(is(arg, Types.Bool), 'arg_require_bool', name)
    return arg
}


function apply_operator (name, a, b) {
    assert(is(name, Types.String))
    if (is(a, Types.Structure)) {
        let s = get_common_schema(a, b)
        return call(s.get_operator(name), [a, b])
    } else {
        assert(is(a, Types.Instance))
        let c = get_common_class(a, b)
        return call(c.get_operator(name), [a, b])
    }
}


function apply_unary (name, a) {
    assert(is(name, Types.String))
    if (is(a, Types.Structure)) {
        return call(a.schema.get_operator(name), [a])
    } else {
        assert(is(a, Types.Instance))
        let C = find(a.class_.super_classes, C => C.defined_operator(name))
        assert(C !== NotFound)
        return call(C.get_operator(name), [a])
    }
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
        `function operator.str (o: Operand<'str'>) -> String`,
            o => apply_unary('str', o),
        'function operator.str (p: Bool) -> String',
            p => p? 'true': 'false',
        'function operator.str (x: Number) -> String',
            x => Number.prototype.toString.call(x),
        'function operator.str (s: String) -> String',
            s => s
    ),
    'len': f (
        'operator.len',
        `function operator.len (o: Operand<'len'>) -> Size`,
            o => apply_unary('len', o),
        'function operator.len (l: List) -> Size',
            l => l.length,
        'function operator.len (s: String) -> Size',
            s => s.length
    ),
    'iter': f (
        'operator.iter',
        `function operator.iter (o: Operand<'iter'>) -> Iterator`,
            o => apply_unary('iter', o),
        'function operator.iter (i: ES_Iterable) -> Iterator',
            l => l[Symbol.iterator]()
    ),
    'negate': f (
        'operator.negate',
        `function operator.negate (o: Operand<'negate'>) -> Object`,
            o => apply_unary('negate', o),
        'function operator.negate (x: Number) -> Number',
            x => -x
    ),
    /* Pull, Push, Derive, Otherwise */
    '<<': f (
        'operator.pull',
        `function operator.pull (a: Operand<'<<'>, b: Operand<'<<'>) -> Any`,
            (a, b) => apply_operator('<<', a, b),
        'function operator.pull (f: Callable, x: Any) -> Any',
            (f, x) => call(f, [x]),
        'function operator.pull (l: Hash, r: Hash) -> Hash',
            (l, r) => Object.assign(l, r),
        'function operator.pull (s: String, x: Any) -> String',
            (s, x) => call(string_format, [s, x])
    ),
    '>>': f (
        'operator.push',
        `function operator.push (a: Operand<'>>'>, b: Operand<'>>'>) -> Any`,
            (a, b) => apply_operator('>>', a, b),
        'function operator.push (x: Any, f: Callable) -> Any',
            (x, f) => call(f, [x])
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
        `function operator.less_than (a: Operand<'<'>, b: Operand<'<'>) -> Bool`,
            (a, b) => apply_operator('<', a, b),
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
        `function operator.equal (a: Operand<'=='>, b: Operand<'=='>) -> Bool`,
            (a, b) => apply_operator('==', a, b),
        'function operator.equal (p: Bool, q: Bool) -> Bool',
            (p, q) => (p === q),
        'function operator.equal (a: String, b: String) -> Bool',
            (a, b) => (a === b),
        'function operator.equal (x: Number, y: Number) -> Bool',
            (x, y) => (x === y)
    ),
    '!=': f (
        'operator.not_equal',
        'function operator.not_equal (l: Any, r: Any) -> Bool',
            (l, r) => !operators['=='](l, r)
    ),
    '~~': f (
        'operator.equivalent',
        'function operator.equivalent (l: Any, r: Any) -> Bool',
            (l, r) => (l === r)
    ),
    '!~': f (
        'operator.not_equivalent',
        'function operator.not_equivalent (l: Any, r: Any) -> Bool',
            (l, r) => (l !== r)
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
        'function operator.keyword_not (T: Type) -> Type',
            T => Not(T)
    ),
    /* Arithmetic */
    '+': f (
        'operator.plus',
        `function operator.plus (a: Operand<'+'>, b: Operand<'+'>) -> Object`,
            (a, b) => apply_operator('+', a, b),
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
        'function operator.plus (x: Number, y: Number) -> GeneralNumber',
            (x, y) => x + y
    ),
    '-': f (
        'operator.minus',
        `function operator.minus (a: Operand<'-'>, b: Operand<'-'>) -> Object`,
            (a, b) => apply_operator('-', a, b),
        'function operator.minus (x: Number, y: Number) -> GeneralNumber',
            (x, y) => x - y
    ),
    '*': f (
        'operator.times',
        `function operator.times (a: Operand<'*'>, b: Operand<'*'>) -> Object`,
            (a, b) => apply_operator('*', a, b),
        'function operator.times (x: Number, y: Number) -> GeneralNumber',
            (x, y) => x * y
    ),
    '/': f (
        'operator.divide',
        `function operator.divide (a: Operand<'/'>, b: Operand<'/'>) -> Object`,
            (a, b) => apply_operator('/', a, b),
        'function operator.divide (x: Number, y: Number) -> GeneralNumber',
            (x, y) => x / y
    ),
    '%': f (
        'operator.modulo',
        `function operator.modulo (a: Operand<'%'>, b: Operand<'%'>) -> Object`,
            (a, b) => apply_operator('%', a, b),
        'function operator.modulo (x: Number, y: Number) -> GeneralNumber',
            (x, y) => x % y
    ),
    '^': f (
        'operator.power',
        `function operator.power (a: Operand<'^'>, b: Operand<'^'>) -> Object`,
            (a, b) => apply_operator('^', a, b),
        'function operator.power (x: Number, y: Number) -> GeneralNumber',
            (x, y) => Math.pow(x, y)
    )
}

Object.freeze(operators)


let operator_is = operators['is']

function str (value) {
    return call(operators['str'], [value])
}

function iter (value) {
    return call(operators['iter'], [value])
}


function get_operator (name) {
    assert(has(name, operators))
    return operators[name]
}
