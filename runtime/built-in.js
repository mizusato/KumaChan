function lazy_bool (arg, desc, name) {
    if (typeof arg != 'boolean') {
        (new ErrorProducer(CallError, desc)).throw(
            MSG.arg_invalid(name)
        )
    }
    return arg
}

let str = f (
    'str',
    'local str (Number x) -> String',
        x => Number.prototype.toString.call(x),
    'local str (String s) -> String',
        s => s
)

let format_err = new ErrorProducer(FormatError, 'str_format()')

let str_format = f (
    'str_format',
    'local str_format (String s, Hash h) -> String',
        (s, h) => {
            h = DeRef(h)
            return s.replace(/\$\{([^}]+)\}/g, (match, p1) => {
                let key = p1
                let ok = has(key, h)
                format_err.assert(ok, ok || MSG.format_invalid_key(key))
                return str(Im(h[key]))
            })
        },
    'local str_format (String s, List l) -> String',
        (s, l) => {
            l = DeRef(l)
            let used = 0
            let result = s.replace(/\$\{(\d+)\}/g, (match, p1) => {
                let index = parseInt(p1) - 1
                let ok = (0 <= index && index < l.length)
                format_err.assert(ok, ok || MSG.format_invalid_index(index))
                used += 1
                return str(Im(l[index]))
            })
            let ok = (used == l.length)
            format_err.assert(ok, ok || MSG.format_not_all_converted)
            return result
        }
)


let operators = {
    /* Pull, Push, Derive, Otherwise */
    '<<': f (
        'operator_pull',
        'local pull (Callable f, Any *x) -> Any',
            (f, x) => f(x),
        'local pull (Hash &l, Hash *r) -> Hash',
            (l, r) => Object.assign(l, r),
        'local pull (String s, Any *x) -> String',
            (s, x) => str_format(s, x)
    ),
    '>>': f (
        'operator_push',
        'local push (Any *l, Any *r) -> Any',
            (l, r) => operators['<<'](r, l)
    ),
    '=>': f (
        'operator_derive',
        'local derive (Bool p, Callable ok) -> Any',
            (p, ok) => p? ok(): Nil
    ),
    'or': f (
        'operator_otherwise',
        'local otherwise (Any *x, Callable fallback) -> Any',
            (x, fallback) => (x !== Nil)? x: fallback()
    ),
    /* Comparsion */
    '<': f (
        'operator_less_than',
        'local less_than (String a, String b) -> Bool',
            (a, b) => a < b,
        'local less_than (Number x, Number y) -> Bool',
            (x, y) => x < y
    ),
    '>': f (
        'operator_greater_than',
        'local greater_than (Any *l, Any *r) -> Bool',
            (l, r) => operators['<'](r, l)
    ),
    '<=': f (
        'operator_less_than_or_equal',
        'local less_than_or_equal (Any *l, Any *r) -> Bool',
            (l, r) => !operators['<'](r, l)
    ),
    '>=': f (
        'operator_greater_than_or_equal',
        'local greater_than_or_equal (Any *l, Any *r) -> Bool',
            (l, r) => !operators['<'](l, r)
    ),
    '==': f (
        'operator_equal',
        'local equal (Any *l, Any *r) -> Bool',
            (l, r) => {
                l = IsRef(l)? DeRef(l): l
                r = IsRef(r)? DeRef(r): r
                return (l === r)
            }
    ),
    '!=': f (
        'operator_not_equal',
        'local not_equal (Any *l, Any *r) -> Bool',
            (l, r) => !operators['=='](l, r)
    ),
    /* Logic */
    '&&': f (
        'operator_and',
        'local and (Bool p, Callable q) -> Bool',
        (p, q) => !p? false: lazy_bool(q(), 'operator_and', 'q')
    ),
    '||': f (
        'operator_or',
        'local or (Bool p, Callable q) -> Bool',
        (p, q) => p? true: lazy_bool(q(), 'operator_or', 'q')
    ),
    '!': f (
        'operator_not',
        'local not (Bool p) -> Bool',
        p => !p
    ),
    '&': f (
        'operator_intersect',
        'local intersect (Abstract A, Abstract B) -> Concept',
        (A, B) => Ins(A, B)
    ),
    '|': f (
        'operator_union',
        'local union (Abstract A, Abstract B) -> Concept',
        (A, B) => Uni(A, B)
    ),
    '~': f (
        'operator_complement',
        'local complement (Abstract A) -> Concept',
        A => Not(A)
    ),
    '\\': f (
        'operator_difference',
        'local difference (Abstract A, Abstract B) -> Concept',
        (A, B) => Ins(A, Not(B))
    ),
    /* Arithmetic */
    '+': f (
        'operator_plus',
        'local plus (List *a, List *b) -> List',
            (a, b) => [...a, ...b],
        'local plus (String a, String b) -> String',
            (a, b) => a + b,
        'local plus (Number x, Number y) -> Number',
            (x, y) => x + y
    ),
    '-': f (
        'operator_minus',
        'local minus (Number x) -> Number',
            x => -x,
        'local minus (Number x, Number y) -> Number',
            (x, y) => x - y
    ),
    '*': f (
        'operator_times',
        'local times (Number x, Number y) -> Number',
            (x, y) => x * y
    ),
    '/': f (
        'operator_divide',
        'local divide (Number x, Number y) -> Number',
            (x, y) => x / y
    ),
    '%': f (
        'operator_modulo',
        'local modulo (Number x, Number y) -> Number',
            (x, y) => x % y
    ),
    '^': f (
        'operator_power',
        'local power (Number x, Number y) -> Number',
            (x, y) => Math.pow(x, y)
    )
}


function call_operator (name) {
    assert(has(name, operators))
    return operators[name]
}


let wrapped_is = fun (
    'local is (Any *x, Abstract A) -> Bool',
        (x, A) => is(x, A)
)


let helpers = scope => ({
    c: call,
    m: (obj, name, args) => call_method(scope, obj, name, args),
    o: call_operator,
    is: wrapped_is,
    id: var_lookup(scope),
    dl: var_declare(scope),
    rt: var_assign(scope),
    v: Void
})
