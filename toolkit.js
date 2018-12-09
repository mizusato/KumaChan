'use strict';


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

const NoValue = $(x => typeof x == 'undefined')
const Num = $(x => typeof x == 'number')
const Int = $n( Num, $(x => Number.isInteger(x)) )
const UnsignedInt = $n( Int, $(x => x >= 0) )
const Str = $(x => typeof x == 'string')
const Bool = $(x => typeof x == 'boolean')
const Hash = $(x => x instanceof Object)
const Optional = concept => $u(NoValue, concept)
const SetEquivalent = (target, concept) => target.contains = concept.contains
const SetMakerConcept = (maker) => maker.contains = (x => x.maker === maker)

const NumStr = $n(Str, $(x => !Number.isNaN(parseInt(x))) )
const Regex = regex => $n( Str, $(s => s.match(regex)) )

const NA = { contains: x => x === this }
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
                `Invalid argument '${parameter}' in function ${callee.name}`
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
    check(map_lazy, arguments, { to_be_mapped: Object, f: Function })
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


function take_if (iterable, f) {
    check(take_if, arguments, { iterable: Iterable, f: Function })
    for ( let I of iterable ) {
        if (f(I)) {
            return I
        }
    }
    return NA
}


function map (to_be_mapped, f) {
    check(map, arguments, { to_be_mapped: Object, f: Function })
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


function filter (to_be_filtered, f) {
    check(filter, arguments, { to_be_filtered: Object })
    if (to_be_filtered.is(Iterable)) {
        let iterable = to_be_filtered
        let result = []
        let index = 0
        for ( let element of iterable ) {
            if ( f(element, index) ) {
                result.push(element)
            }
            index++
        }
        return result
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
        target[key] = source[key]
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
        value = f(element, value, index)
        index++
    }
    return value
}


function forall (to_be_checked, f) {
    check(forall, arguments, { to_be_checked: Object, f: Function })
    if (to_be_checked.is(Iterable)) {
        let iterable = to_be_checked
        return fold(iterable, true, (e,v) => v && f(e))
    } else {
        let hash = to_be_checked
        return fold(Object.keys(hash), true, (k,v) => v && f(hash[k]))
    }
}


function exists (to_be_checked, f) {
    check(exists, arguments, { to_be_checked: Object, f: Function })
    return !forall(to_be_checked, x => !f(x))
}


Object.prototype.has = function (prop) { return this.hasOwnProperty(prop) }
Array.prototype.has = function (index) { return typeof this[index] != 'undefined' }
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
