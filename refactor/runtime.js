(function() {

    function assert (value) {
        if(!value) { throw new Error('Assertion Error') }
        return value
    }

    function pour (o1, o2) {
        return Object.assign(o1, o2)
    }
    
    function list (iterable) {
        let result = []
        for (let I of iterable) {
            result.push(I)
        }
        return result
    }
    
    function *map (iterable, f) {
        let index = 0
        for (let I of iterable) {
            yield f(I, index)
            index += 1
        }
    }

    function mapkey (object, f) {
        let mapped = {}
        for (let key of Object.keys(object)) {
            let value = object[key]
            mapped[f(key, value)] = value
        }
        return mapped
    }

    function mapval (object, f) {
        let mapped = {}
        for (let key of Object.keys(object)) {
            mapped[key] = f(object[key], key)
        }
        return mapped
    }
    
    function *filter (iterable, f) {
        let index = 0
        for (let I of iterable) {
            if (f(I, index)) {
                yield I
            }
            index += 1
        }
    }

    let NotFound = { tip: 'Object Not Found' }
    
    function find (iterable, f) {
        for (let I of iterable) {
            if (f(I)) {
                return I
            }
        }
        return NotFound
    }
    
    function *iterate (initial, next_of, terminate) {
        let value = initial
        while (!terminate(value)) {
            yield value
            value = next_of(value)
        }
    }
    
    function fold (iterable, initial, f, terminate) {
        let index = 0
        let value = initial
        for (let I of iterable) {
            value = f(I, value, index)
            if (terminate && terminate(value)) {
                break
            }
        }
        return value
    }
    
    function forall (iterable, f) {
        return fold(iterable, true, ((e,v) => v && f(e)), (v => v == false))
    }
    
    function exists (iterable, f) {
        return fold(iterable, false, ((e,v) => v || f(e)), (v => v == true))
    }

    function chain (functions) {
        return ( x => fold(functions, x, (f, v) => f(v)) )
    }

    let Checker = Symbol('Checker')
    let DataKey = Symbol('WrapperData')

    function is (value, abstraction) {
        return abstraction[Checker](value)
    }

    function has(key, object) {
        return Object.prototype.hasOwnProperty.call(object, key)
    }
    
    class Concept {
        constructor (checker) {
            assert(typeof checker == 'function')
            this[Checker] = checker
        }
        get [Symbol.toStringTag]() {
            return 'Concept'
        }
    }

    function union (abstracts) {
        assert(forall(abstracts, a => typeof a[Checker] == 'function'))
        return new Concept(x => exists(abstracts, a => a[Checker](x)))
    }
    
    function intersect (abstracts) {
        assert(forall(abstracts, a => typeof a[Checker] == 'function'))
        return new Concept(x => forall(abstracts, a => a[Checker](x)))
    }
    
    function complement (abstraction) {
        assert(typeof abstraction[Checker] == 'function')
        return new Concept(x => !abstraction[Checker](x))
    }

    let $ = (f => new Concept(f))
    let Uni = ((...args) => union(args))
    let Ins = ((...args) => intersect(args))
    let Not = (arg => complement(arg))

    class Category {
        constructor (abstraction, branches) {
            this[Checker] = (
                (abstraction == null)?
                    (union(Object.values(branches)))[Checker]:
                    abstraction[Checker]
            )
            pour(this, branches)
        }
        get [Symbol.toStringTag]() {
            return 'Category'
        }
    }
    
    let category = ((a, b) => new Category(a, b))
    
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

    let Type = {
        Any: $(x => true),
        Undefined: ES.Undefined,
        Null: ES.Null,
        Symbol: ES.Symbol,
        Bool: ES.Boolean,
        Number: category(ES.Number, {
            Int: $(
                x => Number.isInteger(x) && assert(Number.isSafeInteger(x))
            ),
            Safe: $(x => Number.isSafeInteger(x)),
            Finite: $(x => Number.isFinite(x)),
            NaN: $(x => Number.isNaN(x))
        }),
        String: ES.String,
        Function: category(ES.Function, {
            Wrapped: Ins(ES.Function, $(x => has(DataKey, x))),
            Simple: Ins(ES.Function, $(x => !has(DataKey, x)))
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
                Schema: $(x => x instanceof Schema)
            }
        ),
        Container: category(null, {
            List: $(x => x instanceof Array),
            Hash: Ins(ES.Object, $(
                x => Object.getPrototypeOf(x) === Object.prototype
            ))
        }),
        Instance: $(x => x instanceof Instance),
        Reference: $(x => x instanceof Reference)
    }
    
    let NonSolid = Uni(Type.Container, Type.Instance)
    let Solid = Not(NonSolid)
    
    class Reference {
        constructor (object) {
            assert(is(object, NonSolid))
            this.point_to = object
        }
    }
    
    let ImRef = (x => is(x, Solid) && x || new Reference(x))

    class Singleton {
        constructor (description) {
            assert(typeof description == 'string')
            this.description = description
            this[Checker] = (x => x === this)
        }
        get [Symbol.toStringTag]() {
            return `Singleton<${this.description}>`
        }
    }

    let Nil = new Singleton('Nil')
    let Void = new Singleton('Void')
    let Done = new Singleton('Done')

    let list_of = (A => Ins(
        Type.Container.List,
        $(l => forall(l, e => is(e, A)))
    ))
    let StringList = list_of(Type.String)

    class Enum {
        constructor (str_list) {
            assert(is(str_list, StringList))
            let item_set = new Set(str_list)
            this[Checker] = (x => item_set.has(x))
            this.item_list = list(map(str_list, x => x))
        }
    }

    let one_of = ((...items) => new Enum(items))

    class Schema {
        constructor (table, requirement = (x => true)) {
            assert(forall(Object.values(table), v => is(v, Type.Abstract)))
            assert(is(requirement, Type.Function))
            this[Checker] = (x => (
                is(x, Type.Hash)
                    && forall(Object.keys(table), k => is(x[k], table[k]))
                    && requirement(x)
            ))
        }
    }

    let struct = (table => new Schema(table))

    let PassPolicy = one_of('immutable', 'natural', 'dirty')
    let EffectRange = one_of('local', 'upper', 'global')

    let Parameter = struct({
        name: Type.String,
        pass_policy: PassPolicy,
        constraint: Type.Abstract
    })

    let ParameterList = list_of(Parameter)

    let Prototype = struct({
        affect: EffectRange,
        value: Type.Abstract,
        parameters: ParameterList
    })

    class Scope {
        constructor (context, affect = 'local', data = {}) {
            assert(context === null || context instanceof Scope)
            assert(is(affect, EffectRange))
            assert(is(data, Type.Hash))
            this.context = context
            this.affect = affect
            this.data = data
        }
        has (variable) {
            return has(variable, this.data)
        }
        declare (variable, initial_value) {
            assert(!this.has(variable))
            this.data[variable] = initial_value
        }
        assign (variable, new_value) {
            assert(this.has(variable))
            this.data[variable] = new_value
        }
        lookup (variable) {
            let info = this.find(variable)
            if (info == NotFound) {
                return NotFound
            } else {
                if (info.is_mutable) {
                    return info.object
                } else {
                    return ImRef(info.object)
                }
            }
        }
        find (variable) {
            let affect = this.affect
            let mutable_depth = 0
            if (affect == 'local') {
                mutable_depth = 0
            } else if (affect == 'upper') {
                mutable_depth = 1
                let upper = this.context
                while (upper != null && upper.affect == 'upper') {
                    mutable_depth += 1
                    upper = upper.context
                }
            } else if (affect == 'global') {
                mutable_depth = Infinity
            }
            let scope_chain = iterate(
                this, scope => scope.context, scope => scope == null
            )
            return find(map(scope_chain, (scope, depth) => ({
                scope: scope,
                depth: depth,
                is_mutable: depth <= mutable_depth,
                object: scope.has(variable)? scope.data[variable]: NotFound
            })), info => info.object != NotFound)
        }
    }
    
    let Symbols = { DataKey, Checker }
    let export_object = {
        Concept, is, has, $, Uni, Ins, Not, Category, Singleton, Type, Symbols,
        Nil, Void, Done, struct
    }
    let export_name = 'KumaChan'
    let global_scope = (typeof window == 'undefined')? global: window
    global_scope[export_name] = export_object
    
})()
