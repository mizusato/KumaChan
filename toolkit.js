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
const $1 = y => $(x => x === y)
const $f = (...elements) => $(x => exists(elements, e => e === x))

const Any = $(() => true)
const Void = $(() => false)
const Otherwise = Any  // for transform()

const NoValue = $(x => typeof x == 'undefined')
const Num = $(x => typeof x == 'number')
const Int = $n( Num, $(x => Number.isInteger(x)) )
const UnsignedInt = $n( Int, $(x => x >= 0) )
const Str = $(x => typeof x == 'string')
const Bool = $(x => typeof x == 'boolean')
const Hash = $(x => x instanceof Object && !(x instanceof Array))
const Optional = concept => $u(NoValue, concept)
const SetEquivalent = (target, concept) => target.contains = concept.contains
const MadeBy = (maker) => $(x => x.maker === maker)
const SetMakerConcept = (maker) => maker.contains = (x => x.maker === maker)

const NumStr = $n(Str, $(x => !Number.isNaN(parseInt(x))) )
const Regex = regex => $n( Str, $(s => s.match(regex)) )

const Break = { contains: x => x === this }  // for fold()
function BreakWith (value) {
    return { value: value, maker: BreakWith }
}
SetMakerConcept(BreakWith)

const NotFound = { contains: x => x === this }  // for find()
const Nothing = { contains: x => x === this }  // for insert() and added()
const Iterable = $(
    x => typeof x != 'undefined' && typeof x[Symbol.iterator] == 'function'
)
const ArrayOf = (
    concept => $(
        array => Array.contains(array) && forall(array, x=>concept.contains(x))
    )
)
const HashOf = (
    concept => $(
        hash => Hash.contains(hash)
            && forall(Object.keys(hash), key=>concept.contains(hash[key]))
    )
)
function Enum (...str_list) {
    assert( ArrayOf(Str).contains(str_list) )
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
            key => x.has(key)
                && concept_of[key].contains(x[key])
        )
    )
}


Object.prototype.is = function (concept) { return concept.contains(this) }
Object.prototype.is_not = function (concept) { return !this.is(concept) }
Function.prototype.contains = function (obj) { return obj instanceof this }


const once = (function () {
    let cache = {}
    function once (caller, value) {
        check(once, arguments, { caller: Function, value: Any })
        if ( cache.has(caller.name) ) {
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
    assert(items.is(Iterable))
    for ( let item of items ) {
        assert( $u(Result, Function).contains(item) )
        let result = (item.is(Result))? item: item()
        if ( result.is(Failed) ) {
            return result
        } else {
            continue
        }
    }
    return OK
}


function check(callee, args, concept_table) {
    assert(Function.contains(callee))
    assert(Hash.contains(concept_table))
    var parameters = callee.get_parameters()
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
        if (NoValue.contains(argument) && concept.has('defval')) {
            args[parameter_to_index[parameter]] = concept.defval
        }
        */
    }
}


function check_hash(callee, argument, constraint) {
    check(
        check_hash, arguments,
        { callee: Function, argument: Hash, constraint: Hash }
    )
    for ( let name of Object.keys(constraint) ) {
        if (!constraint[name].contains(argument[name])) {
            throw Error(`Invalid argument ${name} in function ${callee.name}`)
        }
        if (NoValue.contains(argument[name]) && constraint[name].has('defval')) {
            argument[name] = constraint[name].defval
        }
    }
}


function* map_lazy (to_be_mapped, f) {
    check(map_lazy, arguments, { to_be_mapped: $u(Object,Str), f: Function })
    if( to_be_mapped.is(Iterable) ) {
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
    check(map, arguments, { to_be_mapped: $u(Object,Str), f: Function })
    return list(map_lazy(to_be_mapped, f))
}


function mapkey (hash, f) {
    check(mapkey, arguments, { hash: Hash, f: Function })
    var result = {}
    for ( let key of Object.keys(hash) ) {
        let new_key = f(key, hash[key])
        result[new_key] = hash[key]
        /*
        if ( !result.has(new_key) ) {
            result[new_key] = hash[key]
        } else {
            throw Error('mapkey(): Key conflict detected')
        } 
        */
    }
    return result
}


function mapval (hash, f) {
    check(mapval, arguments, { hash: Hash, f: Function })
    var result = {}
    for ( let key of Object.keys(hash) ) {
        let value = hash[key]
        result[key] = f(value, key)
    }
    return result
}


function *rev (array) {
    check(rev, arguments, { array: Array })
    for ( let i=array.length-1; i>=0; i-- ) {
        yield array[i]
    }
}


function *cat (...iterables) {
    assert(iterables.is(Iterable))
    for( let iterable of iterables ) {
        assert(iterable.is(Iterable))
	for ( let element of iterable ) {
	    yield element	
	}
    }
}


function *lazy (f) {
    assert(Function.contains(f))
    let iterable = f()
    assert(iterable.is(Iterable))
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
        f: Function
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
        to_be_filtered: Object,
        f: Function
    })
    if (to_be_filtered.is(Iterable)) {
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
        { iterable: Iterable, initial: Any, f: Function }
    )
    var value = initial
    var index = 0
    for ( let element of iterable ) {
        let new_value = f(element, value, index)
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
        situations: ArrayOf(Struct({
            when_it_is: Concept,
            use: Function
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
        next_of: Function,
        terminate_when: Concept
    })
    let current = initial
    while ( !terminate_when.contains(current) ) {
        yield current
        current = next_of(current)
    }
}


function forall (to_be_checked, f) {
    check(forall, arguments, { to_be_checked: Object, f: Function })
    if (to_be_checked.is(Iterable)) {
        let iterable = to_be_checked
        return Boolean(fold(iterable, true, (e,v) => v && f(e)))
    } else {
        let hash = to_be_checked
        return Boolean(fold(Object.keys(hash), true, (k,v) => v && f(hash[k])))
    }
}


function exists (to_be_checked, f) {
    check(exists, arguments, { to_be_checked: Object, f: Function })
    return !forall(to_be_checked, x => !f(x))
}


function find (container, f) {
    check(find, arguments, { container: Object, f: Function })
    if (container.is(Iterable)) {
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
    assert(iterable.is(Iterable))
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
    assert(iterable.is(Iterable))
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
    assert(iterable.is(Iterable))
    assert(f.is(Function))
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


Object.prototype.transform_by = function (f) { return f(this) }
Object.prototype.has = function (prop) { return this.hasOwnProperty(prop) }
Object.prototype.has_ = Object.prototype.has
Array.prototype.has = function (index) { return typeof this[index] != 'undefined' }
Array.prototype.added = function (element) {
    let r = list(this)
    if (element != Nothing) {
        r.push(element)
    }
    return r
}
Array.prototype.added_front = function (element) {
    if (element != Nothing) {
        return list(cat([element], this))
    } else {
        return list(this)
    }
}
Object.prototype.has_no = function (prop) { return !this.has(prop) }
Array.prototype.has_no = function (index) { return !this.has(index) }
String.prototype.realCharAt = function (index) {
    return take_at(this, index, '')
}
String.prototype.genuineLength = function () {
    return fold(this, 0, (e,v) => v+1)
}
Function.prototype.get_parameters = function () {
    /* https://stackoverflow.com/questions/1007981/how-to-get-thistion-parameter-names-values-dynamically */
    var STRIP_COMMENTS = /(\/\/.*$)|(\/\*[\s\S]*?\*\/)|(\s*=[^,\)]*(('(?:\\'|[^'\r\n])*')|("(?:\\"|[^"\r\n])*"))|(\s*=[^,\)]*))/mg;
    var ARGUMENT_NAMES = /([^\s,]+)/g;
    var str = this.toString().replace(STRIP_COMMENTS, '')
    return (
        str
            .slice(str.indexOf('(')+1, str.indexOf(')'))
            .match(ARGUMENT_NAMES)
    ) || []
}
