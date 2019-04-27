/**
 *  Initialize Global Scope
 */

let Global = new Scope(null)
let G = Global.data

pour(Types, {
    Callable: Uni(Types.ES_Function, Types.TypeTemplate, Types.Class),
    Iterable: $(x => typeof x[Symbbol.iterator] == 'function')
})

pour(Global.data, {
    Type: Type,
    Any: Types.Any,
    Object: Types.Any,
    Nil: Nil,
    Void: Void,
    Bool: Types.Bool,
    Number: Types.Number,
    Int: Types.Int,
    String: Types.String,
    Function: Types.Function,
    Binding: Types.Binding,
    Overload: Types.Overload,
    Callable: Types.Callable,
    List: Types.List,
    Hash: Types.Hash,
    Iterable: Types.Iterable,
    es: {
        undefined: undefined,
        null: null,
        Symbol: ES.Symbol
    },
})


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
    'function str (x: Number) -> String',
        x => Number.prototype.toString.call(x),
    'function str (s: String) -> String',
        s => s
)

let str_format = f (
    'str_format',
    'function str_format (s: String, h: Hash) -> String',
        (s, h) => {
            h = DeRef(h)
            return s.replace(/\$\{([^}]+)\}/g, (match, p1) => {
                let key = p1
                let ok = has(key, h)
                ensure(ok, 'format_invalid_key', key)
                format_err.assert(ok, ok || MSG.format_invalid_key(key))
                return str(Im(h[key]))
            })
        },
    'function str_format (s: String, l: List) -> String',
        (s, l) => {
            l = DeRef(l)
            let used = 0
            let result = s.replace(/\$\{(\d+)\}/g, (match, p1) => {
                let index = parseInt(p1) - 1
                let ok = (0 <= index && index < l.length)
                ensure(ok, 'format_invalid_index', index)
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
        'function pull (f: Callable, x: Any) -> Any',
            (f, x) => f(x),
        'function pull (l: Hash, r: Hash) -> Hash',
            (l, r) => Object.assign(l, r),
        'function pull (s: String, x: Any) -> String',
            (s, x) => str_format(s, x)
    ),
    '>>': f (
        'operator_push',
        'function push (l: Any, r: Any) -> Any',
            (l, r) => operators['<<'](r, l)
    ),
    '=>': f (
        'operator_derive',
        'function derive (p: Bool, ok: Callable) -> Any',
            (p, ok) => p? ok(): Nil
    ),
    'or': f (
        'operator_otherwise',
        'function otherwise (x: Any, fallback: Callable) -> Any',
            (x, fallback) => (x !== Nil)? x: fallback()
    ),
    /* Comparsion */
    '<': f (
        'operator_less_than',
        'function less_than (a: String, b: String) -> Bool',
            (a, b) => a < b,
        'function less_than (x: Number, y: Number) -> Bool',
            (x, y) => x < y
    ),
    '>': f (
        'operator_greater_than',
        'function greater_than (l: Any, r: Any) -> Bool',
            (l, r) => operators['<'](r, l)
    ),
    '<=': f (
        'operator_less_than_or_equal',
        'function less_than_or_equal (l: Any, r: Any) -> Bool',
            (l, r) => !operators['<'](r, l)
    ),
    '>=': f (
        'operator_greater_than_or_equal',
        'function greater_than_or_equal (l: Any, r: Any) -> Bool',
            (l, r) => !operators['<'](l, r)
    ),
    '==': f (
        'operator_equal',
        'function equal (l: Any, r: Any) -> Bool',
            (l, r) => {
                l = NoRef(l)
                r = NoRef(r)
                return (l === r)
            }
    ),
    '!=': f (
        'operator_not_equal',
        'function not_equal (l: Any, r: Any) -> Bool',
            (l, r) => !operators['=='](l, r)
    ),
    /* Logic */
    '&&': f (
        'operator_and',
        'function and (p: Bool, q: Callable) -> Bool',
            (p, q) => !p? false: lazy_bool(q(), 'operator_and', 'q')
    ),
    '||': f (
        'operator_or',
        'function or (p: Bool, q: Callable) -> Bool',
            (p, q) => p? true: lazy_bool(q(), 'operator_or', 'q')
    ),
    '!': f (
        'operator_not',
        'function not (p: Bool) -> Bool',
            p => !p
    ),
    '&': f (
        'operator_intersect',
        'function intersect (A: Type, B: Type) -> Type',
            (A, B) => Ins(A, B)
    ),
    '|': f (
        'operator_union',
        'function union (A: Type, B: Type) -> Type',
            (A, B) => Uni(A, B)
    ),
    '~': f (
        'operator_complement',
        'function complement (A: Type) -> Type',
            A => Not(A)
    ),
    '\\': f (
        'operator_difference',
        'function difference (A: Type, B: Type) -> Type',
            (A, B) => Ins(A, Not(B))
    ),
    'not': f (
        'operator_keyword_not',
        'function keyword_not (p: Bool) -> Bool',
            p => !p,
        'function keyword_not (A: Type) -> Type',
            A => Not(A)
    ),
    /* Arithmetic */
    '+': f (
        'operator_plus',
        'function plus (a: Hash, b: Hash) -> Hash',
            (a, b) => Object.assign({}, a, b),
        'function plus (a: Iterable, b: Iterable) -> Iterable',
            (a, b) => {
                return (function* ()  {
                    for (let I of a) { yield I }
                    for (let I of b) { yield I }
                })()
            },
        'function plus (a: List, b: List) -> List',
            (a, b) => [...a, ...b],
        'function plus (a: String, b: String) -> String',
            (a, b) => a + b,
        'function plus (x: Number, y: Number) -> Number',
            (x, y) => x + y
    ),
    '-': f (
        'operator_minus',
        'function minus (x: Number) -> Number',
            x => -x,
        'function minus (x: Number, y: Number) -> Number',
            (x, y) => x - y
    ),
    '*': f (
        'operator_times',
        'function times (x: Number, y: Number) -> Number',
            (x, y) => x * y
    ),
    '/': f (
        'operator_divide',
        'function divide (x: Number, y: Number) -> Number',
            (x, y) => x / y
    ),
    '%': f (
        'operator_modulo',
        'function modulo (x: Number, y: Number) -> Number',
            (x, y) => x % y
    ),
    '^': f (
        'operator_power',
        'function power (x: Number, y: Number) -> Number',
            (x, y) => Math.pow(x, y)
    )
}


function call_operator (name) {
    assert(has(name, operators))
    return operators[name]
}


let wrapped_is = fun (
    'function is (x: Any, T: Type) -> Bool',
        (x, A) => is(x, A)
)


function bind_method_call (scope) {
    return (obj, name, args, file, row, col) => {
        return call_method(scope, obj, name, args, file, row, col)
    }
}


let helpers = scope => ({
    c: call,
    m: bind_method_call(scope),
    o: call_operator,
    is: wrapped_is,
    id: scope.lookup.bind(scope),
    dl: scope.declare.bind(scope),
    rt: scope.reset.bind(scope),
    cl: create_class,
    it: create_interface,
    v: Void
})
