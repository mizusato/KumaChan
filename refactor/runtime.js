(function() {

    function assert (value) {
        if(!value) { throw Error('Assertion Error') }
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

    const Checker = Symbol()
    
    class Concept {
        constructor (checker) {
            assert(typeof checker == 'function')
            this[Checker] = checker
        }
        static union (abstracts) {
            assert(forall(abstracts, a => is(a, Abstract)))
            return new Concept(x => exists(abstracts, a => a[Checker](x)))
        }
        static intersect (abstracts) {
            assert(forall(abstracts, a => is(a, Abstract)))
            return new Concept(x => forall(abstracts, a => a[Checker](x)))
        }
        static complement (abstraction) {
            assert(is(abstraction, Abstract))
            return new Concept(x => !abstraction[Checker](x))
        }
    }

    const Abstract = new Concept(
        x => typeof x == 'object' && typeof x[Checker] == 'function'
    )
    
    function is (value, abstraction) {
        return abstraction[Checker](value)
    }

    function has(key, object) {
        return Object.prototype.hasOwnProperty.call(object, key)
    }

    let $ = (f => new Concept(f))
    let Uni = ((...concepts) => Concept.union(concepts))
    let Ins = ((...concepts) => Concept.intersect(concepts))
    let Not = (concept => Concept.complement(concept))

    class Category {
        constructor (abstracts) {
            this[Checker] = (Concept.union(Object.values(abstracts)))[Checker]
            pour(this, abstracts)
        }
    }

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

    let DataKey = Symbol('WrapperData')
    let Type = {
        Any: $(x => true),
        Undefined: ES.Undefined,
        Null: ES.Null,
        Boolean: ES.Boolean,
        Number: ES.Number,
        String: ES.String,
        Function: new Category({
            Wrapped: Ins(ES.Function, $(x => has(DataKey, x))),
            Simple: Ins(ES.Function, $(x => !has(DataKey, x)))
        }),
        Abstract: Abstract,
        List: $(x => x instanceof Array),
        Hash: Ins(ES.Object, $(
            x => Object.getPrototypeOf(x) === Object.prototype
        ))
    }
    
    
    let export_object = { Concept, is, has, Category, Type }
    let export_name = 'KumaChan'
    let global_scope = (typeof window == 'undefined')? global: window
    global_scope[export_name] = export_object
    
})()
