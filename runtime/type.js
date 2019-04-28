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
    return type[Checker](value)
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
 *  ES6 Raw Types
 */
let ES = {
    Undefined: $(x => typeof x == 'undefined'),
    Null: $(x => x === null),
    Boolean: $(x => typeof x == 'boolean'),
    Number: $(x => typeof x == 'number'),
    String: $(x => typeof x == 'string'),
    Symbol: $(x => typeof x == 'symbol'),
    Function: $(x => typeof x == 'function'),
    Object: $(x => typeof x == 'object' && x !== null),
}


/**
 *  Basic Types
 */
let Types = {
    Type: Type,
    Bool: ES.Boolean,
    Number: ES.Number,
    String: ES.String,
    Int: $(x => Number.isInteger(x) && assert(Number.isSafeInteger(x))),
    List: $(x => x instanceof Array),
    Hash: Ins(ES.Object, $(x => get_proto(x) === Object.prototype)),
    Any: Any
}


/**
 *  TypeTemplate: Implementation of Generics
 *
 *  A type template is a function-like type object, which can be inflated by
 *    a fixed number of arguments (types) and returns a new type.
 *  The arguments and returned type will be cached. If the same arguments
 *    are provided in the next inflation, the cached type will be returned.
 *  For any object x and type template TT, x ∈ TT if and only if there exists
 *    a cached type T of TT that satisfies x ∈ T.
 */
class TypeTemplate {
    constructor (arity, inflater) {
        assert(is(arity, Types.Int))
        assert(is(inflater, ES.Function))
        this.arity = arity
        this.inflater = inflater
        this.inflate = this.inflate.bind(this)
        this.cache = []
        this[Checker] = (x => exists(this.cache, item => is(x, item.type)))
        Object.freeze(this)
    }
    inflate (args) {
        let cached = find(this.cache, item => equal(item.args, args))
        if (cached != NotFound) {
            return cached.type
        } else {
            let arity = this.arity
            let quantity_ok = (args.length == arity)
            ensure(quantity_ok, 'arg_wrong_quantity', arity, args.length)
            for (let i=0; i<arity; i++) {
                let is_type = is(args[i], Type)
                let inflater_info = this.inflater[WrapperInfo]
                ensure(is_type, 'arg_not_type', is_type || (
                    inflater_info?
                        inflater_info.proto.parameters[i].name:
                        `#${i}`
                ))
            }
            let type = this.inflater.apply(null, args)
            ensure(is(type, Type), 'retval_not_type')
            this.cache.push({
                args: copy(args),
                type: type
            })
            return type
        }
    }
    of (...types) {
        return this.inflate(types)
    }
    get [Symbol.toStringTag]() {
        return 'TypeTemplate'
    }
}

Types.TypeTemplate = $(x => x instanceof TypeTemplate)

/* shorthand */
let template = (f => new TypeTemplate(f.length, f))


/**
 *  Basic Generic Types
 */

/* TypedList<T> = List & { l | for all element in l, element ∈ T } */
Types.TypedList = template (
    T => Ins(Types.List, $(
        l => forall(l, e => is(e, T))
    ))
)

/* TypedHash<T> = Hash & { h | for all element in h, element ∈ T } */
Types.TypedHash = template (
    T => Ins(Types.Hash, $(
        h => forall(get_keys(h), k => is(h[k], T))
    ))
)


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
 *  Enum: An encapsulation of TypedHash<Singleton>
 */
class Enum {
    constructor (item_names, desc) {
        assert(is(item_names, Types.TypedList.of(Types.String)))
        this.items = {}
        this.values = []
        foreach(item_names, name => {
            let item = create_value(`${name} (${desc})`)
            this.items[name] = item
            this.values.push(item)
        })
        Object.freeze(this.items)
        Object.freeze(this.values)
        this[Checker] = (x => exists(this.values, v => v === x))
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Enum'
    }
}

Types.Enum = $(x => x instanceof Enum)


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
 *  Schema: Data format constraints on Hash.
 */
class Schema {
    constructor (table, defaults = null, requirement = (x => true)) {
        assert(is(table, Types.TypedHash.of(Type)))
        assert(is(defaults, Uni(ES.Null, Types.Hash)))
        assert(is(requirement, ES.Function))
        this.table = copy(table)
        Object.freeze(table)
        this.requirement = requirement
        if (defaults != null) {
            this.defaults = copy(defaults)
            Object.freeze(this.defaults)
            for (let key of get_keys(this.defaults)) {
                assert(has(key, this.table))
                let defval = this.defaults[key]
                let type = this.table[key]
                ensure(is(defval, type), 'schema_invalid_default', key)
            }
        } else {
            this.defaults = null
        }
        this[Checker] = (x => {
            if(!is(x, Types.Hash)) { return false }
            let defaults = this.defaults || {}
            for (let key of get_keys(this.table)) {
                let required_type = this.table[key]
                if (has(key, x)) {
                    if(!is(x[key], required_type)) {
                        return false
                    }
                } else {
                    if(!has(key, defaults)) {
                        return false
                    }
                }
            }
            let result = this.requirement(x)
            assert(is(result, Types.Bool))
            return result
        })
        Object.freeze(this)
    }
    patch (hash) {
        // apply default values to hash
        assert(is(hash, Types.Hash))
        if (this.defaults == null) {
            return hash
        }
        let new_hash = copy(hash)
        for (let key of get_keys(this.defaults)) {
            if (!has(key, new_hash)) {
                new_hash[key] = this.defaults[key]
            }
        }
        return new_hash
    }
    get [Symbol.toStringTag]() {
        return 'Schema'
    }
}

Types.Schema = $(x => x instanceof Schema)

// shorthand
let struct = ((table, def, req) => new Schema(table, def, req))
