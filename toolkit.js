'use strict';


function assert (bool) {
    if(!bool) {
        throw Error('Assertion Error')
    }
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
const Str = $(x => typeof x == 'string')
const Bool = $(x => typeof x == 'boolean')
const Hash = $(x => x instanceof Object)
const Optional = (concept, defval) => pour({defval: defval},$u(concept, NoValue))
const SetEquivalent = (target, concept) => target.contains = concept.contains
const SetMakerConcept = (maker) => maker.contains = (x => x.maker === maker)

const NumStr = $n(Str, $(x => !Number.isNaN(Number(x))) )

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
function Struct (hash) {
    check(Struct, arguments, { hash: HashOf(Concept) })
    return $(x => forall(Object.keys(hash), key => x.has(key) && hash[key].contains(x[key])))
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


function list (iterable) {
    check(list, arguments, { iterable: Iterable })
    var result = []
    for ( let element of iterable ) {
        result.push(element)
    }
    return result
}


function filter (to_be_filtered, f) {
    check(filter, arguments, { to_be_filtered: Hash })
    if (to_be_filtered.is(Iterable)) {
        let iterable = to_be_filtered
        let result = []
        for ( let element of iterable ) {
            if ( f(element) ) {
                result.push(element)
            }
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


function forall (iterable, f) {
    check(forall, arguments, { iterable: Iterable, f: Function })
    return fold(iterable, true, (e,v) => v && f(e))
}


function exists (iterable, f) {
    check(exists, arguments, { iterable: Iterable, f: Function })
    return fold(iterable, false, (e,v) => v || f(e))
}


Object.prototype.has = function (prop) { return this.hasOwnProperty(prop) }
Array.prototype.has = function (index) { return typeof this[index] != 'undefined' }
Object.prototype.has_no = function (prop) { return !this.has(prop) }
Array.prototype.has_no = function (index) { return !this.has(index) }
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
