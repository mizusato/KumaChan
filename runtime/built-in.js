let IndexType = Ins(Types.Int, $(x => x >= 0))

pour(Types, {
    Object: Types.Any,
    Callable: Uni(Types.ES_Function, Types.TypeTemplate, Types.Class),
    Iterable: $(x => typeof x[Symbol.iterator] == 'function'),
    Arity: template(fun(
        'function Arity (n: Int) -> Type',
            n => Ins(Types.Function, $(
                f => f[WrapperInfo].proto.parameters.length == n
            ))
    )),
    Index: IndexType,
    Size: IndexType
})

Object.freeze(Types)


let Global = new Scope(null, {
    Type: Type,
    Any: Types.Any,
    Object: Types.Object,
    Nil: Nil,
    Void: Void,
    Bool: Types.Bool,
    Number: Types.Number,
    NaN: Types.NaN,
    Infinite: Types.Infinite,
    MayNotNumber: Types.MayNotNumber,
    Int: Types.Int,
    String: Types.String,
    Function: Types.Function,
    Binding: Types.Binding,
    Overload: Types.Overload,
    Callable: Types.Callable,
    Arity: Types.Arity,
    List: Types.List,
    Hash: Types.Hash,
    Iterable: Types.Iterable,
    es: {
        undefined: undefined,
        null: null,
        Symbol: ES.Symbol
    },
    print: fun (
        'function print (s: String) -> Void',
        s => (console.log(s), Void)
    )
}, true)

let Eval = new Scope(Global)
let default_scopes = { Global, Eval }
Object.freeze(default_scopes)


function lazy_bool (arg, desc, name) {
    ensure(is(arg, Types.Bool), 'arg_require_bool', name)
    return arg
}

let get_data = f (
    'get_data',
    'function get_data (nil: Nil, k: Any, nf: Bool) -> Object',
        () => Nil,
    'function get_data (l: List, i: Index, nf: Bool) -> Object',
        (l, i, nf) => (i < l.length)? l[i]: (ensure(nf, 'index_error', i), Nil),
    'function get_data (h: Hash, k: String, nf: Bool) -> Object',
        (h, k, nf) => has(k, h)? h[k]: (ensure(nf, 'key_error', k), Nil)
)

let set_data = f (
    'set_data',
    'function set_data (nil: Nil, k: Any, v: Any) -> Void',
        () => Void,
    'function set_data (l: List, i: Index, v: Any) -> Void',
        (l, i, v) => {
            ensure(i < l.length, 'index_error', i)
            l[i] = v
            return Void
        },
    'function set_data (h: Hash, k: String, v: Any) -> Void',
        (h, k, v) => {
            h[k] = v
            return Void
        }
)


let for_loop = f (
    'for_loop',
    'function for_loop (h: Hash) -> Iterable',
        h => mapkv(h, (k, v) => ({ key: k, value: v })),
    'function for_loop (i: Iterable) -> Iterable',
        i => map(i, (e, i) => ({ key: i, value: e }))
)


let string_format = f (
    'string_format',
    'function string_format (s: String, h: Hash) -> String',
        (s, h) => {
            return s.replace(/\$\{([^}]+)\}/g, (match, p1) => {
                let key = p1
                let ok = has(key, h)
                ensure(ok, 'format_invalid_key', key)
                return call(operators['str'], [h[key]])
            })
        },
    'function string_format (s: String, l: List) -> String',
        (s, l) => {
            let used = 0
            let result = s.replace(/\$\{(\d+)\}/g, (match, p1) => {
                let index = parseInt(p1) - 1
                let ok = (0 <= index && index < l.length)
                ensure(ok, 'format_invalid_index', index)
                used += 1
                return call(operators['str'], [l[index]])
            })
            let ok = (used == l.length)
            ensure(ok, 'format_not_all_converted')
            return result
        }
)


let operators = {
    'is': f (
        'operator.is',
        'function operator.is (x: Any, T: Type) -> Bool',
            (x, A) => is(x, A)
    ),
    'str': f (
        'operator.str',
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
        'function operator.minus (x: Number) -> Number',
            x => -x,
        'function operator.minus (x: Number, y: Number) -> MayNotNumber',
            (x, y) => x - y
    ),
    '*': f (
        'operator.times',
        'function operator.times (x: Number, y: Number) -> MayNotNumber',
            (x, y) => x * y
    ),
    '/': f (
        'operator.divide',
        'function operator.divide (x: Number, y: Number) -> MayNotNumber',
            (x, y) => x / y
    ),
    '%': f (
        'operator.modulo',
        'function operator.modulo (x: Number, y: Number) -> MayNotNumber',
            (x, y) => x % y
    ),
    '^': f (
        'operator.power',
        'function operator.power (x: Number, y: Number) -> MayNotNumber',
            (x, y) => Math.pow(x, y)
    )
}

Object.freeze(operators)


function get_operator (name) {
    assert(has(name, operators))
    return operators[name]
}


function bind_method_call (scope) {
    return (obj, name, args, file, row, col) => {
        return call_method(scope, obj, name, args, file, row, col)
    }
}


function ensure_failed (e, name, args) {
    // TODO: record throw position
    e.type = 1
    e.info = { name, args }
    throw new UserlandEnsureFailed(name)
}

function try_failed (e, name) {
    // TODO: record throw position
    e.type = 2
    e.info = { name }
    throw new UserlandTryFailed(name)
}

function inject_ensure_args (scope, names, e) {
    assert(scope instanceof Scope)
    assert(is(names, TypedList.of(Types.String)))
    assert(is(e.args, Types.List))
    foreach(names, (name, i) => {
        if (i < e.args.length) {
            scope.declare(name, e.args[i])
        } else {
            scope.declare(name, Nil)
        }
    })
}

inject_desc(inject_ensure_args, 'inject_ensure_args')


let global_helpers = {
    a: Types.Any,
    c: call,
    o: get_operator,
    g: (o, k, nf, f, r, c) => call(get_data, [o, k, nf], f, r, c),
    s: set_data,
    cl: create_class,
    it: create_interface,
    ef: ensure_failed,
    tf: try_failed,
    ie: inject_ensure_args,
    f: for_loop,
    v: Void
}

Object.freeze(global_helpers)


let get_helpers = scope => ({
    m: bind_method_call(scope),
    id: inject_desc(scope.lookup.bind(scope), 'lookup_variable'),
    dl: inject_desc(scope.declare.bind(scope), 'declare_variable'),
    rt: inject_desc(scope.reset.bind(scope), 'reset_variable'),
    gv: f => get_vals(f, scope),
    w: (proto, vals, desc, raw) => wrap(scope, proto, vals, desc, raw),
    __: global_helpers
})
