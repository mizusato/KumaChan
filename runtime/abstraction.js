/**
 *  Abstraction (Concept, Category, Singleton, Enum, Schema, Type)
 */


/**
 *  Abstraction Mechanics
 *
 *  An object A that satisfies (typeof A[Checker] == 'function')
 *    is called an "abstraction object" or "abstract set".
 *  The function A[Checker]() should take one argument x
 *    and return a boolean value which indicates whether x ∈ A.
 *  In this programming language, data types are implemented by
 *    abstraction objects. A data type is just an abstraction object,
 *    therefore data type checking is performed at runtime,
 *    by calling the [Checker]() function of the abstraction object.
 */

var Type;

function is (value, abstraction) {
    return abstraction[Checker](value)
}

function has (key, object) {
    assert(typeof key == 'string' || typeof key == 'symbol')
    return Object.prototype.hasOwnProperty.call(object, key)
}


/**
 *  Concept Object
 *
 *  The so-called "concept objects" are the simplest kind of
 *    abstraction objects, which only contains a [Checker]() function.
 *  In other words, concept objects do not contain any extra information
 *    except its abstraction information.
 *  Any abstraction object can do operations such as intersect,
 *    union, complement with others, and produce a new concept object.
 */

class Concept {
    constructor (checker) {
        assert(typeof checker == 'function')
        this[Checker] = checker
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Concept'
    }
}

function union (abstracts) {
    // (∪ A_i), for A_i in abstracts
    assert(forall(abstracts, a => typeof a[Checker] == 'function'))
    return new Concept(x => exists(abstracts, a => a[Checker](x)))
}

function intersect (abstracts) {
    // (∩ A_i), for A_i in abstracts
    assert(forall(abstracts, a => typeof a[Checker] == 'function'))
    return new Concept(x => forall(abstracts, a => a[Checker](x)))
}

function complement (abstraction) {
    // (∁ A)
    assert(typeof abstraction[Checker] == 'function')
    return new Concept(x => !abstraction[Checker](x))
}

let $ = (f => new Concept(f))             // create concept from f(x)
let Uni = ((...args) => union(args))      // (A,B,...) => A ∪ B ∪ ...
let Ins = ((...args) => intersect(args))  // (A,B,...) => A ∩ B ∩ ...
let Not = (arg => complement(arg))        //  A => ∁ A
let Any = $(x => true)


/**
 *  Category Object
 *
 *  A collection of abstraction objects can be integrated into
 *    a "category object", which is also an abstraction object.
 */

class Category {
    constructor (precondition, branches) {
        let concept = Ins(precondition, union(Object.values(branches)))
        this[Checker] = concept[Checker]
        pour(this, mapval(branches, A => {
            if (A instanceof Category) {
                return new Category(
                    Ins(precondition, A[BranchInfo].precondition),
                    A[BranchInfo].branches
                )
            } else {
                return Ins(precondition, A)
            }
        }))
        this[BranchInfo] = Object.freeze({ precondition, branches })
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Category'
    }
    static get_branch (object, category) {
        assert(is(object, category[BranchInfo].precondition))
        let branches = category[BranchInfo].branches
        for (let b of Object.keys(branches)) {
            if (is(object, branches[b])) {
                if (branches[b] instanceof Category) {
                    return Array.concat(
                        [b],
                        Category.get_branch(object, branches[b])
                    )
                } else {
                    return [b]
                }
            }
        }
        assert(false)
    }
}

let category = ((a, b) => new Category(a, b))


/**
 *  Singleton Object
 *
 *  The so-called "singleton object" is just a kind of abstraction
 *    in this language. If S is a singleton object, it means that
 *    S = { x | x === S }, i.e. (x ∈ S) if and only if (x === S).
 *  The singleton object mechanics is used to create special values,
 *    such as Nil and Void, which are available by default.
 */

class Singleton {
    constructor (description) {
        assert(typeof description == 'string')
        this.description = description
        this[Checker] = (x => x === this)
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return `Singleton<${this.description}>`
    }
}

let Nil = new Singleton('Nil')
let Void = new Singleton('Void')


/**
 *  Enumeration Object
 *
 *  An enumeration is just a set of string.
 */

class Enum {
    constructor (str_list) {
        assert(is(str_list, list_of(Type.String)))
        let item_set = new Set(str_list)
        this[Checker] = (x => item_set.has(x))
        this.item_list = Object.freeze(list(map(str_list, x => x)))
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Enum'
    }
}

let one_of = ((...items) => new Enum(items))


/**
 *  Schema Object
 *
 *  A schema is an abstraction of Hash Objects with specified structure.
 */

class Schema {
    constructor (table, defaults = null, requirement = (x => true)) {
        assert(forall(Object.values(table), v => is(v, Type.Abstract)))
        assert(is(defaults, Uni(Type.Null, Type.Container.Hash)))
        assert(is(requirement, Type.Function))
        this.table = table
        this.requirement = requirement
        if (defaults != null) {
            this.defaults = Object.freeze(defaults)
            let err = new ErrorProducer(SchemaError)
            for (let key of Object.keys(this.defaults)) {
                assert(has(key, this.table))
                err.assert(
                    is(this.defaults[key], this.table[key]),
                    MSG.schema_invalid_default(key)
                )
            }
        }
        this[Checker] = (x => {
            if(!is(x, Type.Container.Hash)) { return false }
            if (this.defaults == null) {
                return (forall(
                    Object.keys(this.table),
                    k => is(x[k], this.table[k])
                ) && this.requirement(x))
            } else {
                for (let key of Object.keys(this.table)) {
                    if (has(key, x)) {
                        if(!is(x[key], this.table[key])) {
                            return false
                        }
                    } else if(has(key, this.defaults)) {
                        x[key] = this.defaults[key]
                    } else {
                        return false
                    }
                }
                return this.requirement(x)
            }
        })
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Schema'
    }
}

let struct = ((table, def, req) => new Schema(table, def, req))


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
    Object: $(x => typeof x == 'object' && x !== null)
}


/**
 *  Wrapped Types
 */

Type = {
    Undefined: ES.Undefined,
    Null: ES.Null,
    Symbol: ES.Symbol,
    Bool: ES.Boolean,
    Number: ES.Number,
    String: ES.String,
    Function: category(ES.Function, {
        Bare: $(f => !has(WrapperInfo, f)),
        Wrapped: category(
            $(f => has(WrapperInfo, f)),
            {
                Sole: $(f => has('context', f[WrapperInfo])),
                Overload: $(f => has('functions', f[WrapperInfo])),
                Binding: $(f => has('original', f[WrapperInfo]))
            }
        )
    }),
    Abstract: category(
        $(
            x => typeof x == 'object'
                && typeof x[Checker] == 'function'
                && Object.getPrototypeOf(x) !== Object.prototype
        ),
        {
            Concept: $(x => x instanceof Concept),
            Category: $(x => x instanceof Category),
            Singleton: $(x => x instanceof Singleton),
            Enum: $(x => x instanceof Enum),
            Schema: $(x => x instanceof Schema),
            Class: $(x => x instanceof Class),
            Signature: $(x => x instanceof Signature),
            Interface: $(x => x instanceof Interface)
        }
    ),
    Container: category(ES.Object, {
        List: $(x => x instanceof Array),
        Hash: $(x => Object.getPrototypeOf(x) === Object.prototype)
    }),
    Instance: $(x => x instanceof Instance)
}

let list_of = (A => Ins(
    Type.Container.List,
    $(l => forall(l, e => is(e, A)))
))

let hash_of = (A => Ins(
    Type.Container.Hash,
    $(h => forall(Object.keys(h), k => is(h[k], A)))
))

let UserlandTypeRename = {
    Sole: 'Function',
    Binding: 'Function'
}

function get_type(object) {
    let type = 'Raw'
    for (let T of Object.keys(Type)) {
        if (is(object, Type[T])) {
            if (Type[T] instanceof Category) {
                let l = Category.get_branch(object, Type[T])
                type = l[l.length-1]
                break
            } else {
                type = T
                break
            }
        }
    }
    return UserlandTypeRename[type] || type
}

