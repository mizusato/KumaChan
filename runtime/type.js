/**
 *  Type System
 *
 *  Types are runtime-manipulatable objects in this language. This idea is
 *    inspired by the Set Theory.
 *  A type object T contains a [Checker] function that takes an argument x,
 *    and returns a boolean value which indicates whether x ∈ T.
 *  In other words, a type object is no more than an encapsulation of a
 *    single-parameter boolean-valued function.
 */
const Checker = Symbol('Checker')
const ValueName = Symbol('ValueName')


/**
 *  Simplist type object containing no extra data.
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
/* universe set Ω */
let Any = $(x => true)
/* empty set ∅ */
let Never = $(x => false)


/**
 *  `Type` is the definition of type object,
 *     which is also a type object.
 */
let Type = $(x => (
    (x !== null)
    && typeof x == 'object'
    && typeof x[Checker] == 'function'
    && assert(Object.isFrozen(x))
))


/**
 *  Definition of value ∈ type
 */
function is (value, type) {
    assert(Type[Checker](type))
    let r = type[Checker](value)
    assert(typeof r == 'boolean')
    return r
}


/**
 *  Definition of intersect ⋂, union ⋃, and complement ∁
 *
 *  An instance of CompoundType is called compound type, meanwhile
 *    a type that isn't a compound type is called an atomic type.
 *  The propose to define this class is to provide a mechanics to
 *    determine if two types are logical equivalent, such as
 *    ∁(Ω, A ∪ B) ⇔ ∁(Ω, A) ∩ ∁(Ω, B) but ⇎ ∁(Ω, A) ∪ ∁(Ω, B),
 *  Above equivalency checking is possible because through extracting
 *    all atomic types depended by the compound type, we can treat
 *    the compond type as a propositional formula and evaluate it
 *    over a truth table of all atomic types depended by it.
 */
class CompoundType {
    constructor (op, args) {
        /**
         *  Opcode Definition:
         *    0. intersect: ⋂ T, for T in args
         *    1. union: ⋃ T, for T in args
         *    2. complement: ∁(Ω, T), where T = args[0]
         *    3. wrap: equivalence of T, where T = args[0]
         */
        assert(exists([0,1,2,3], v => op == v))
        assert(args instanceof Array)
        assert(forall(args, arg => is(arg, Type)))
        if (op == 2 || op == 3) { assert(args.length == 1) }
        this.op = op
        this.args = copy(args)
        if (this.op == 0) {
            // ∀ T, T ∩ Ω = T
            this.args = this.args.filter(T => !is_any(T))
            // ∀ T, T ∩ ∅ = ∅
            if (exists(this.args, T => is_never(T))) {
                this.op = 3
                this.args = [Never]
            }
        }
        if (this.op == 1) {
            // ∀ T, T ∪ ∅ = T
            this.args = this.args.filter(T => !is_never(T))
            // ∀ T, T ∪ Ω = Ω
            if (exists(this.args, T => is_any(T))) {
                this.op = 3
                this.args = [Any]
            }
        }
        if (this.op == 2) {
            let arg = this.args[0]
            if (is_any(arg)) {
                // ∁(Ω, Ω) = ∅
                this.op = 3
                this.args = [Never]
            } else if (is_never(arg)) {
                // C(Ω, ∅) = Ω
                this.op = 3
                this.args = [Any]
            }
        }
        this.atomic_args = extract_atomic_args(this.args)
        Object.freeze(this.args)
        Object.freeze(this.atomic_args)
        if (this.op == 0) {
            this[Checker] = x => forall(this.args, T => T[Checker](x))
        } else if (this.op == 1) {
            this[Checker] = x => exists(this.args, T => T[Checker](x))
        } else if (this.op == 2) {
            this[Checker] = x => !(this.args[0][Checker](x))
        } else {
            this[Checker] = this.args[0][Checker]
        }
        Object.freeze(this)
    }
    evaluate (value_map) {
        // evaluate the type using a truth table of its atomic arguments
        assert(value_map instanceof Map)
        if (this.op == 0) {
            // intersect
            return forall(this.args, T => evaluate_type(T, value_map))
        } else if (this.op == 1) {
            // union
            return exists(this.args, T => evaluate_type(T, value_map))
        } else if (this.op == 2) {
            // complement
            return !(evaluate_type(this.args[0], value_map))
        } else if (this.op == 3) {
            // just wrap it
            return evaluate_type(this.args[0], value_map)
        }
    }
}


function is_any (T) {
    return (
        (T === Any)
        || (T instanceof CompoundType && T.op == 3 && T.args[0] === Any)
    )
}


function is_never (T) {
    return (
        (T === Never)
        || (T instanceof CompoundType && T.op == 3 && T.args[0] === Never)
    )
}


function extract_atomic_args (args) {
    return new Set((function* () {
        for (let T of args) {
            assert(is(T, Type))
            if (T instanceof CompoundType) {
                for (let arg of T.atomic_args) {
                    yield arg
                }
            } else {
                yield T
            }
        }
    })())
}


function evaluate_type (T, value_map) {
    if (T instanceof CompoundType) {
        // evaluate recursively
        return T.evaluate(value_map)
    } else {
        // read from the truth table
        assert(value_map.has(T))
        return Boolean(value_map.get(T))
    }
}


function type_equivalent (T1, T2) {
    assert(is(T1, Type) && is(T2, Type))
    // optimization
    if (T1 === T2) {
        return true
    }
    // wrap atomic types
    if (!(T1 instanceof CompoundType)) {
        T1 = new CompoundType(3, [T1])
    }
    if (!(T2 instanceof CompoundType)) {
        T2 = new CompoundType(3, [T2])
    }
    // check if T1 and T2 have same dependencies
    if (!set_equal(T1.atomic_args, T2.atomic_args)) {
        return false
    }
    // prepare the truth table
    let args = list(T1.atomic_args)
    let L = args.length
    let N = Math.pow(2, L)
    assert(Number.isSafeInteger(N))
    let value_map = new Map()
    for (let arg of args) {
        value_map.set(arg, false)
    }
    // iterate over the truth table
    let i = 0
    let j = 0
    while (i < N) {
        if (T1.evaluate(value_map) != T2.evaluate(value_map)) {
            return false
        }
        // do a binary addition
        for (j = 0; j < L; j++) {
            if (value_map.get(args[j]) == false) {
                value_map.set(args[j], true)
                break
            } else {
                value_map.set(args[j], false)
            }
        }
        i += 1
    }
    return true
}


/**
 *  Basic Type Operators
 */
function intersect (types) {
    // (⋂ T), for T in types
    return new CompoundType(0, types)
}

function union (types) {
    // (⋃ T), for T in types
    return new CompoundType(1, types)
}

function complement (type) {
    // (∁ T)
    return new CompoundType(2, [type])
}

/* shorthand */
let Uni = ((...args) => union(args))      // (A,B,...) => A ∪ B ∪ ...
let Ins = ((...args) => intersect(args))  // (A,B,...) => A ∩ B ∩ ...
let Not = (arg => complement(arg))        //  A => ∁(Ω, A)


/**
 *  Singleton Type
 *
 *  A singleton type S is defined by S = { x | x === S },
 *    i.e. (x ∈ S) if and only if (x === S).
 *  Singleton types may be used as unique special values or enum values,
 *    such as Nil, Void and FoobarEnum.Foo.
 */
function create_value (name) {
    assert(typeof name == 'string')
    // use a null prototype to disable == conversion with primitive values
    let value = Object.create(null)
    value[ValueName] = name
    value[Checker] = (x => x === value)
    Object.freeze(value)
    return value
}

let Singleton = $(x => typeof x[ValueName] == 'string')
let Nil = create_value('Nil')
let Void = create_value('Void')


/**
 *  Collection type that only contains a specific list of objects.
 */
class FiniteSetType {
    constructor (objects) {
        assert(objects instanceof Array)
        assert(objects.length > 0)
        this.objects = copy(objects)
        Object.freeze(this.objects)
        this[Checker] = (x => exists(this.objects, object => object === x))
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'FiniteSetType'
    }
}

// shorthand
let one_of = ((...objects) => new FiniteSetType(objects))


/**
 *  ES6 Raw Types (Modified from Original Definition)
 */
let ES = {
    Undefined: $(x => typeof x == 'undefined'),
    Null: $(x => x === null),
    Boolean: $(x => typeof x == 'boolean'),
    Number: $(x => (
        typeof x == 'number'
            && !Number.isNaN(x)    // NaN is called "not a number"
            && Number.isFinite(x)  // Inifinite breaks the number system
    )),
    NaN: $(x => Number.isNaN(x)),
    Infinite: $(x => !Number.isFinite(x) && !Number.isNaN(x)),
    String: $(x => typeof x == 'string'),
    Symbol: $(x => typeof x == 'symbol'),
    Function: $(x => typeof x == 'function'),
    Object: $(x => typeof x == 'object' && x !== null),
    Iterable: $(x => (
        typeof x == 'string'
            || (
                typeof x == 'object'
                    && x !== null
                    && typeof x[Symbol.iterator] == 'function'
            )
    ))
}


/**
 *  Basic Types
 */
let Types = {
    /* Special Types */
    Type, Any, Never, Nil, Void,
    Object: Any,
    /* Primitive Value Types */
    String: ES.String,
    Bool: ES.Boolean,
    Number: ES.Number,
    NaN: ES.NaN,
    Infinite: ES.Infinite,
    Int: $(x => Number.isSafeInteger(x)),
    GeneralNumber: $(x => typeof x == 'number'),
    Primitive: Uni(ES.Number, ES.String, ES.Boolean),
    /* Basic Container Types */
    List: $(x => x instanceof Array),
    Hash: Ins(ES.Object, $(x => get_proto(x) === Object.prototype)),
    /* ECMAScript Compatible Types */
    ES_Object: Uni(Ins(ES.Object, $(x => {
        // objects that aren't in control of our runtime
        let p = get_proto(x)
        if (get_proto(x) === Object.prototype) { return false }
        if (is(x, Type)) { return false }
        return forall(
            [Array, Error, Function, Instance, Struct],
            T => !(x instanceof T)
        )
    })), $(x => x === Function.prototype)),
    ES_Symbol: ES.Symbol,
    ES_Key: Uni(ES.String, ES.Symbol),
    ES_Class: Ins(ES.Function, $(
        f => is(f.prototype, ES.Object) || f === Function
    )),
    ES_Iterable: ES.Iterable
    // this list will keep growing until the init of runtime is finished
}


/**
 *  Typed Container/Structure Types for Internal Use
 *
 *  These types are fragile and cost a lot,
 *    therefore should be only used by internal runtime code.
 */
let TypedList = {
    of: T => assert(is(T, Type)) && Ins(Types.List, $(
        l => forall(l, e => is(e, T))
    ))
}

let TypedHash = {
    of: T => assert(is(T, Type)) && Ins(Types.Hash, $(
        h => forall(Object.keys(h), k => is(h[k], T))
    ))
}

class HashFormat {
    constructor (table, requirement = (x => true)) {
        assert(is(table, TypedHash.of(Type)))
        assert(is(requirement, ES.Function))
        this.table = copy(table)
        Object.freeze(table)
        this.requirement = requirement
        this[Checker] = (x => {
            if (!is(x, Types.Hash)) { return false }
            for (let key of Object.keys(this.table)) {
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



class ObjectRegistry {
    constructor () {
        let self = {
            map: new Map(),   // value --> id
            next_id: 0
        }
        this.get_id = this.get_id.bind(self)
        Object.freeze(this)
    }
    get_id (value) {
        if (this.map.has(value)) {
            return this.map.get(value)
        } else {
            if (is(value, Type)) {
                for (let item of this.map) {
                    let [T, id] = item
                    if (is(T, Type) && type_equivalent(T, value)) {
                        this.map.set(value, id)
                        return id
                    }
                }
            }
            let id = this.next_id
            this.map.set(value, id)
            this.next_id = id + 1
            assert(is(this.next_id, Types.Int))
            return id
        }
    }
}


class VectorMapCache {
    constructor () {
        let self = {
            registry: new ObjectRegistry(),
            groups: new Map(),
            hash: this.hash
        }
        this.set = this.set.bind(self)
        this.find = this.find.bind(self)
        Object.freeze(this)
    }
    hash (id_vector) {
        assert(!(this instanceof VectorMapCache))
        assert(is(id_vector, TypedList.of(Types.Int)))
        let L = id_vector.length
        let value = L % 10
        let offset = 10
        for (let i = 0; i < L && i < 10; i += 1) {
            value += offset * (id_vector[i] % 10)
            offset *= 10
        }
        return value
    }
    set (vector, value) {
        assert(is(vector, Types.List))
        let id_vector = vector.map(element => this.registry.get_id(element))
        let h = this.hash(id_vector)
        if (this.groups.has(h)) {
            let group = this.groups.get(h)
            assert(is(group, Types.List))
            for (let item of group) {
                let [iv, _] = item
                assert(!equal(iv, id_vector))
            }
            group.push([id_vector, value])
        } else {
            this.groups.set(h, [[id_vector, value]])
        }
    }
    find (vector) {
        assert(is(vector, Types.List))
        let id_vector = vector.map(element => this.registry.get_id(element))
        let h = this.hash(id_vector)
        if (this.groups.has(h)) {
            let group = this.groups.get(h)
            assert(is(group, Types.List))
            for (let item of group) {
                let [iv, value] = item
                if (equal(iv, id_vector)) {
                    return value
                }
            }
            return NotFound
        } else {
            return NotFound
        }
    }
}
