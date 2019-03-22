'use strict';


const TAB = '\t'
const CR = '\r'
const LF = '\n'


function assert (bool) {
    if(!bool) {
        throw Error('Assertion Error')
    }
    return true
}


class Concept {
    constructor (contains) {
        assert(typeof contains == 'function')
        this.contains = contains
    }

    static intersect (...concepts) {
        return new Concept(function (x) {
            var result = true
            for(let concept of concepts) {
                if (!concept.contains(x)) {
                    result = false
                    return result
                }
            }
            return result
        })
    }

    static union (...concepts) {
        return new Concept(function (x) {
            var result = false
            for(let concept of concepts) {
                if (concept.contains(x)) {
                    result = true
                    return result
                }
            }
            return result
        })
    }

    static complement (A) {
        return new Concept(x => !A.contains(x))
    }

    static contains (x) {
        return (typeof x.contains == 'function')
    }
}


const $ = f => new Concept(f)
const $n = Concept.intersect
const $u = Concept.union
const $_ = Concept.complement
const $d = (x, y) => $n(x, $_(y))
const $1 = y => $(x => x === y)
const $f = (...elements) => $(x => exists(elements, e => e === x))

const Any = $(any => true)
const Void = $(any => false)
const Otherwise = Any  // for transform()

const NoValue = $(x => typeof x == 'undefined')
const Num = $(x => typeof x == 'number')
const Int = $n( Num, $(x => Number.isInteger(x)) )
const UnsignedInt = $n( Int, $(x => x >= 0) )
const Str = $(x => typeof x == 'string')
const Bool = $(x => typeof x == 'boolean')
const JsObject = $(x => x instanceof Object)
const Fun = $(object => object instanceof Function)
const List = $(object => object instanceof Array)
const Hash = $n( $(x => x instanceof Object), $_(List) )
const StrictHash = $n( $(x => x instanceof Object), $_($u(List, Fun)) )
const Optional = concept => $u(NoValue, concept)
const SetEquivalent = (target, concept) => target.contains = concept.contains
const MadeBy = (maker) => $n(JsObject, $(x => x.maker === maker))
const SetMakerConcept = (maker) => maker.contains = MadeBy(maker).contains

const NumStr = $n(Str, $(x => !Number.isNaN(parseInt(x))) )
const Regex = regex => $n( Str, $(s => s.match(regex)) )

const SingletonContains = function(x) { return x === this }
const Break = { contains: SingletonContains, _name: 'Break' }  // for fold()
function BreakWith (value) {
    return { value: value, maker: BreakWith }
}
SetMakerConcept(BreakWith)

// special value for find()
const NotFound = { contains: SingletonContains, _name: 'NotFound' }
// special value for insert() and added()
const Nothing = { contains: SingletonContains, _name: 'Nothing' }  
const Iterable = $(
    x => typeof x != 'undefined' && typeof x[Symbol.iterator] == 'function'
)
const ListOf = (
    concept => $(
        array => List.contains(array) && forall(array, x=>concept.contains(x))
    )
)
const HashOf = (
    concept => $(
        hash => Hash.contains(hash)
            && forall(Object.keys(hash), key=>concept.contains(hash[key]))
    )
)
function Enum (...str_list) {
    assert( ListOf(Str).contains(str_list) )
    var set = new Set(str_list)
    return $(function (item) {
        return Str.contains(item) && set.has(item)
    })
}
const one_of = Enum
function Struct (concept_of) {
    check(Struct, arguments, { concept_of: HashOf(Concept) })
    return $(
        x => forall(
            Object.keys(concept_of),
            key => has(x, key)
                && concept_of[key].contains(x[key])
        )
    )
}


function is (object, concept) {
    return concept.contains(object)
}


function is_not (object, concept) {
    return !is(object, concept)
}


const once = (function () {
    let cache = {}
    function once (caller, value) {
        check(once, arguments, { caller: Fun, value: Any })
        if ( has(cache, caller.name) ) {
            return cache[caller.name]
        } else {
            cache[caller.name] = value
            return value
        }
    }
    return once
})()


function Failed (message) {
    return {
        message: message,
        maker: Failed
    }
}


SetMakerConcept(Failed)


const OK = { name: 'OK', contains: x => x === OK }


const Result = $u(OK, Failed)


function suppose (bool, message) {
    check(suppose, arguments, { bool: Bool, message: Str })
    if (bool) {
        return OK
    } else {
        return Failed(message)
    }
}


function need (items) {
    assert(Iterable.contains(items))
    for ( let item of items ) {
        assert( $u(Result, Fun).contains(item) )
        let result = (Result.contains(item))? item: item()
        if ( Failed.contains(result) ) {
            return result
        } else {
            continue
        }
    }
    return OK
}


function check(callee, args, concept_table) {
    assert(Fun.contains(callee))
    assert(Hash.contains(concept_table))
    var parameters = get_function_parameters(callee)
    var parameter_to_index = {}
    var index = 0
    for ( let parameter of parameters ) {
        parameter_to_index[parameter] = index
        index++
    }
    for ( let parameter of Object.keys(concept_table) ) {
	var argument = args[parameter_to_index[parameter]]
        var concept = concept_table[parameter]
        if (!concept.contains(argument)) {
	    throw Error(
                `Invalid argument '${parameter}' in function ${callee.name}()`
            )
        }
        /*
        if (NoValue.contains(argument) && has(concept, 'defval')) {
            args[parameter_to_index[parameter]] = concept.defval
        }
        */
    }
}


function check_hash(callee, argument, constraint) {
    check(
        check_hash, arguments,
        { callee: Fun, argument: Hash, constraint: Hash }
    )
    for ( let name of Object.keys(constraint) ) {
        if (!constraint[name].contains(argument[name])) {
            throw Error(`Invalid argument ${name} in function ${callee.name}`)
        }
        if (NoValue.contains(argument[name]) && has(constraint[name], 'defval')) {
            argument[name] = constraint[name].defval
        }
    }
}


function* map_lazy (to_be_mapped, f) {
    check(map_lazy, arguments, { to_be_mapped: $u(JsObject,Str), f: Fun })
    if( Iterable.contains(to_be_mapped) ) {
        let iterable = to_be_mapped
        let index = 0
        for ( let I of iterable ) {
            yield f(I, index)
            index += 1
        }
    } else {
        let hash = to_be_mapped
        for ( let key of Object.keys(hash) ) {
            yield f(key, hash[key])
        }
    }
}


function map (to_be_mapped, f) {
    check(map, arguments, { to_be_mapped: $u(JsObject,Str), f: Fun })
    return list(map_lazy(to_be_mapped, f))
}


function mapkey (hash, f) {
    check(mapkey, arguments, { hash: Hash, f: Fun })
    var result = {}
    for ( let key of Object.keys(hash) ) {
        let new_key = f(key, hash[key])
        result[new_key] = hash[key]
        /*
        if ( !has(result, new_key) ) {
            result[new_key] = hash[key]
        } else {
            throw Error('mapkey(): Key conflict detected')
        } 
        */
    }
    return result
}


function mapval (hash, f) {
    check(mapval, arguments, { hash: Hash, f: Fun })
    var result = {}
    for ( let key of Object.keys(hash) ) {
        let value = hash[key]
        result[key] = f(value, key)
    }
    return result
}


function *rev (array) {
    check(rev, arguments, { array: List })
    for ( let i=array.length-1; i>=0; i-- ) {
        yield array[i]
    }
}


function *cat (...iterables) {
    for( let iterable of iterables ) {
        assert(Iterable.contains(iterable))
        for ( let element of iterable ) {
            yield element	
        }
    }
}


function concat (iterable_of_iterable) {
    check(concat, arguments, { iterable_of_iterable: Iterable })
    return cat.apply({}, iterable_of_iterable)
}


function *lazy (f) {
    assert(Fun.contains(f))
    let iterable = f()
    assert(Iterable.contains(iterable))
    for ( let I of iterable ) {
        yield I
    }
}


function* take_while (iterable, f) {
    let index = 0
    for ( let I of iterable ) {
        if (f(I, index)) {
            yield I
        } else {
            break
        }
        index++
    }
}


function* map_while (iterable, condition, f) {
    take_while(map_lazy(iterable, f), condition)
}


function take_at (iterable, index, default_value) {
    check(take_at, arguments, { iterable: Iterable })
    let count = 0;
    for (let I of iterable) {
        if (count == index) {
            return I
        }
        count++
    }
    return default_value
}


function list (iterable) {
    check(list, arguments, { iterable: Iterable })
    var result = []
    for ( let element of iterable ) {
        result.push(element)
    }
    return result
}


function join (iterable, separator) {
    check(join, arguments, { iterable: Iterable, separator: Str })
    return list(iterable).join(separator)
}


function join_lines (...lines) {
    return join(lines, LF);
}


function* filter_lazy (iterable, f) {
    check(filter_lazy, arguments, {
        iterable: Iterable,
        f: Fun
    })
    let index = 0
    let count = 0
    for ( let element of iterable ) {
        if ( f(element, index, count) ) {
            yield element
            count++
        }
        index++
    }
}


function filter (to_be_filtered, f) {
    check(filter, arguments, {
        to_be_filtered: JsObject,
        f: Fun
    })
    if (Iterable.contains(to_be_filtered)) {
        let iterable = to_be_filtered
        return list(filter_lazy(iterable, f))
    } else {
        let hash = to_be_filtered
        let result = {}
        for ( let key of Object.keys(hash) ) {
            if ( f(key, hash[key]) ) {
                result[key] = hash[key]
            }
        }
        return result
    }
}


function pour (target, source) {
    check(pour, arguments, { target: Hash, source: Hash })
    for ( let key of Object.keys(source) ) {
        if (key != 'contains') {
            target[key] = source[key]
        }
    }
    if ( source.__proto__ !== Object.prototype ) {
        target.__proto__ = source.__proto__
    }
    return target
}


function fold (iterable, initial, f) {
    check(
        fold, arguments,
        { iterable: Iterable, initial: Any, f: Fun }
    )
    var value = initial
    var index = 0
    for ( let element of iterable ) {
        let new_value = f(element, value, index)
        assert(new_value !== undefined)
        if (Break.contains(new_value)) {
            break
        } else if (BreakWith.contains(new_value)) {
            value = new_value.value
            break
        } else {
            value = new_value
            index++
        }
    }
    return value
}


function chain (...functions) {
    return ( x => fold(functions, x, (f,v) => f(v)) )
}


function transform (object, situations) {
    check(transform, arguments, {
        object: Any,
        situations: ListOf(Struct({
            when_it_is: Concept,
            use: Fun
        }))
    })
    for (let s of situations) {
        let concept = s.when_it_is
        let f = s.use
        if ( concept.contains(object) ) {
            return f(object)
        }
    }
    throw Error('transform(): cannot find proper transformation')
}


function* iterate (initial, next_of, terminate_when) {
    check(iterate, arguments, {
        initial: Any,
        next_of: Fun,
        terminate_when: Concept
    })
    let current = initial
    while ( !terminate_when.contains(current) ) {
        yield current
        current = next_of(current)
    }
}


function forall (to_be_checked, f) {
    check(forall, arguments, { to_be_checked: JsObject, f: Fun })
    if (Iterable.contains(to_be_checked)) {
        let iterable = to_be_checked
        return Boolean(fold(iterable, true, (e,v) => v && f(e)))
    } else {
        let hash = to_be_checked
        return Boolean(fold(Object.keys(hash), true, (k,v) => v && f(hash[k])))
    }
}


function exists (to_be_checked, f) {
    check(exists, arguments, { to_be_checked: JsObject, f: Fun })
    return !forall(to_be_checked, x => !f(x))
}


function find (container, f) {
    check(find, arguments, { container: JsObject, f: Fun })
    if (Iterable.contains(container)) {
        let iterable = container
        for ( let I of iterable ) {
            if (f(I)) {
                return I
                break
            }
        }
    } else {
        let hash = container
        for ( let key of Object.keys(hash) ) {
            if (f(hash[key], key)) {
                return { key: key, value: hash[key] }
                break
            }
        }
    }
    return NotFound
}


function* lookahead (iterable, empty_value) {
    assert(Iterable.contains(iterable))
    let it = iterable[Symbol.iterator]()
    let current = it.next()
    let next = it.next()
    let third = it.next()
    while ( !current.done ) {
        yield {
            current: current.value,
            next: next.done? empty_value: next.value,
            third: third.done? empty_value: third.value
        }
        current = next
        next = third
        third = it.next()
    }
}


function* lookaside (iterable, empty_value) {
    assert(Iterable.contains(iterable))
    let it = iterable[Symbol.iterator]()
    let left = { value: empty_value, done: false }
    let current = it.next()
    let right = it.next()
    while ( !current.done ) {
        yield {
            left: left.value,
            current: current.value,
            right: right.done? empty_value: right.value
        }
        left = current
        current = right
        right = it.next()
    }
}


function* insert (iterable, empty_value, f) {
    assert(Iterable.contains(iterable))
    assert(Fun.contains(f))
    for( let look of lookahead(iterable, empty_value)) {
        yield look.current
        let add = f(look.current, look.next)
        if (add != Nothing) {
            yield add
        }
    }
}


function* count (n) {
    let i = 0
    while (i < n) {
        yield i
        i += 1
    }
}


function apply_on (object, f) {
    return f(object)
}


function has (hash, key) {
    return Object.prototype.hasOwnProperty.call(hash, key)
}


function added (iterable, element) {
    check(added, arguments, { iterable: Iterable, element: Any })
    let r = list(iterable)
    if (element != Nothing) {
        r.push(element)
    }
    return r
}


function added_front (iterable, element) {
    check(added_front, arguments, { iterable: Iterable, element: Any })
    if (element != Nothing) {
        return list(cat([element], iterable))
    } else {
        return list(iterable)
    }
}


function get_function_parameters (f) {
    /* https://stackoverflow.com/questions/1007981/how-to-get-thistion-parameter-names-values-dynamically */
    var STRIP_COMMENTS = /(\/\/.*$)|(\/\*[\s\S]*?\*\/)|(\s*=[^,\)]*(('(?:\\'|[^'\r\n])*')|("(?:\\"|[^"\r\n])*"))|(\s*=[^,\)]*))/mg;
    var ARGUMENT_NAMES = /([^\s,]+)/g;
    var str = f.toString().replace(STRIP_COMMENTS, '')
    return (
        str
            .slice(str.indexOf('(')+1, str.indexOf(')'))
            .match(ARGUMENT_NAMES)
    ) || []
}
