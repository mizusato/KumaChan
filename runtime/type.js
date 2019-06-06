/**
 *  Type System
 *
 *  Types are runtime-manipulatable objects in this language. This idea is
 *    inspired by the Set Theory.
 *  A type object T contains a [Checker] function that takes an argument x,
 *    and returns a boolean value which indicates whether x ∈ T.
 *  In other words, a type object is no more than an encapsulation of a
 *    single-argument boolean-value function.
 */
const Checker = Symbol('Checker')
const ValueName = Symbol('ValueName')


/**
 *  SimpleType: The most simplist type object, containing no extra data.
 */
class SimpleType {
    constructor (checker) {
        assert(typeof checker == 'function')
        this[Checker] = checker
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'SimpleType'
    }
}

/* shorthand */
let $ = f => new SimpleType(f)
let Any = $(x => true)


/**
 *  Type: Definition of type objects. Type is also a type.
 */
let Type = $(x => (
    (x !== null)
    && typeof x == 'object'
    && typeof x[Checker] == 'function'
    && assert(Object.isFrozen(x))
))


/**
 *  Shorthand for Type Checking
 */
function is (value, type) {
    assert(Type[Checker](type))
    let r = type[Checker](value)
    assert(typeof r == 'boolean')
    return r
}


/**
 *  Basic Operators for Types
 */
function union (types) {
    // (∪ T), for T in types
    assert(forall(types, t => is(t, Type)))
    return new SimpleType(x => exists(types, t => t[Checker](x)))
}

function intersect (types) {
    // (∩ T), for T in types
    assert(forall(types, t => is(t, Type)))
    return new SimpleType(x => forall(types, t => t[Checker](x)))
}

function complement (type) {
    // (∁ T)
    assert(is(type, Type))
    return new SimpleType(x => !type[Checker](x))
}

/* shorthand */
let Uni = ((...args) => union(args))      // (A,B,...) => A ∪ B ∪ ...
let Ins = ((...args) => intersect(args))  // (A,B,...) => A ∩ B ∩ ...
let Not = (arg => complement(arg))        //  A => ∁ A


/**
 *  ES6 Raw Types (Modified from Original Definition)
 */
let ES = {
    Undefined: $(x => typeof x == 'undefined'),
    Null: $(x => x === null),
    Boolean: $(x => typeof x == 'boolean'),
    Number: $(x => (
        typeof x == 'number'
            && !Number.isNaN(x)
            && Number.isFinite(x)
    )),
    NaN: $(x => Number.isNaN(x)),
    Infinite: $(x => !Number.isFinite(x) && !Number.isNaN(x)),
    String: $(x => typeof x == 'string'),
    Symbol: $(x => typeof x == 'symbol'),
    Function: $(x => typeof x == 'function'),
    Object: $(x => typeof x == 'object' && x !== null),
    Iterable: $(x => (
        typeof x == 'object'
            && x !== null
            && typeof x[Symbol.iterator] == 'function'
    ))
}


/**
 *  Basic Types
 */
let Types = {
    Type: Type,
    Bool: ES.Boolean,
    Number: ES.Number,
    NaN: ES.NaN,
    Infinite: ES.Infinite,
    GeneralNumber: $(x => typeof x == 'number'),
    String: ES.String,
    Int: $(x => Number.isInteger(x) && Number.isSafeInteger(x)),
    Primitive: Uni(ES.Number, ES.String, ES.Boolean),
    List: $(x => x instanceof Array),
    Hash: Ins(ES.Object, $(x => get_proto(x) === Object.prototype)),
    ES_Object: Uni(Ins(ES.Object, $(x => {
        let p = get_proto(x)
        let p_ok = (p !== Object.prototype && p !== null)
        if (!p_ok) { return false }
        if (is(x, Type)) { return false }
        return forall(
            [Array, Error, Function, Instance, Structure],
            T => !(x instanceof T)
        )
    })), $(x => x === Function.prototype)),
    ES_Symbol: ES.Symbol,
    ES_Key: Uni(ES.String, ES.Symbol),
    ES_Class: Ins(ES.Function, $(
        f => is(f.prototype, ES.Object) || f === Function
    )),
    ES_Iterable: ES.Iterable,
    Any: Any
}


/**
 *  Typed Container Types for Internal Use (Don't Export them to Built-In)
 */
let TypedList = {
    of: T => assert(is(T, Type)) && Ins(Types.List, $(
        l => forall(l, e => is(e, T))
    ))
}
let TypedHash = {
    of: T => assert(is(T, Type)) && Ins(Types.Hash, $(
        h => forall(get_keys(h), k => is(h[k], T))
    ))
}


/**
 *  Singleton Object
 *
 *  A sinlgeton object S is a type, defined by S = { x | x === S },
 *    i.e. (x ∈ S) if and only if (x === S).
 *  Singleton objects are used to create special values and enum values,
 *    such as Nil, Void and FoobarEnum.Foo.
 */

function create_value (name) {
    let value = Object.create(null)
    value[ValueName] = name
    value[Checker] = (x => x === value)
    Object.freeze(value)
    return value
}

Types.Singleton = $(x => typeof x[ValueName] == 'string')

let Nil = create_value('Nil')
let Void = create_value('Void')
Types.Nil = Nil
Types.Void = Void


/**
 *  Finite Set: Types that only contains a specific list of objects.
 */
class Finite {
    constructor (objects) {
        assert(is(objects, Types.List))
        this.objects = copy(objects)
        Object.freeze(this.objects)
        this[Checker] = (x => exists(this.objects, object => object === x))
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Finite'
    }
}

Types.Finite = $(x => x instanceof Finite)

// shorthand
let one_of = ((...objects) => new Finite(objects))

/**
 *  HashFormat: Data format constraints on Hash. (for Internal Use)
 */
class HashFormat {
    constructor (table, requirement = (x => true)) {
        assert(is(table, TypedHash.of(Type)))
        assert(is(requirement, ES.Function))
        this.table = copy(table)
        Object.freeze(table)
        this.requirement = requirement
        this[Checker] = (x => {
            if (!is(x, Types.Hash)) { return false }
            for (let key of get_keys(this.table)) {
                let required_type = this.table[key]
                if (has(key, x) && is(x[key], required_type)) {
                    continue
                } else {
                    return false
                }
            }
            let result = this.requirement(x)
            assert(is(result, Types.Bool))
            return result
        })
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'HashFormat'
    }
}

// shorthand
let format = ((table, def, req) => new HashFormat(table, def, req))
