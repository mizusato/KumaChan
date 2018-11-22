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

    static intersect (A, B) {
        return new Concept(x => A.contains(x) && B.contains(x))
    }

    static union (A, B) {
        return new Concept(x => A.contains(x) || B.contains(x))
    }

    static complement (A) {
        return new Concept(x => !A.contains(x))
    }

    static contains (x) {
        return (typeof x.contains == 'function')
    }
}


var $ = f => new Concept(f)
var $I = x => x
var $n = Concept.intersect
var $u = Concept.union
var $_ = Concept.complement
var Any = $(() => true)
var Void = $(() => false)
var Empty = $(x => typeof x == 'undefined')
var Num = $(x => typeof x == 'number')
var Str = $(x => typeof x == 'string')
var Bool = $(x => typeof x == 'boolean')
var Hash = $(x => x instanceof Object)
var NA = {}
NA.contains = x => x === NA
var Iterable = $(function (x) {
    return typeof x != 'undefined' && typeof x[Symbol.iterator] == 'function'
})
var Optional = function (concept) { return $u(concept, Empty) }
var ArrayOf = function (concept) {
    return $(function (array) {
        return (
            (array instanceof Array)
                && forall(array, x=>concept.contains(x))
        )
    })
}
var Enum = function (...str_list) {
    assert(str_list.is(ArrayOf(Str)) )
    var set = new Set(str_list)
    return $(function (item) {
        return Str.contains(item) && set.has(item)
    })
}
Object.prototype.is = function (concept) { return concept.contains(this) }
Object.prototype.is_not = function (concept) { return !this.is(concept) }
Function.prototype.contains = function (obj) { return obj instanceof this }
Array.prototype.contains = function (obj) { return this.indexOf(obj) != -1 }


function check(callee, args, concept_table) {
    var parameters = callee.get_parameters()
    var parameter_to_index = {}
    var index = 0
    for ( let parameter of parameters ) {
        parameter_to_index[parameter] = index
        index++
    }
    if ( concept_table.is(Hash) ) {
        for ( let parameter of Object.keys(concept_table) ) {
	    var argument = args[parameter_to_index[parameter]]
            var concept = concept_table[parameter]
            if (!concept.contains(argument)) {
	        throw Error(
                    `Invalid argument '${parameter}' in function ${callee.name}`
                )
            }
        }
    } else {
        let concept = concept_table
        for ( let i=0; i<callee.arguments.length; i++ ) {
            let argument = arguments[i]
            if (argument.is_not(concept)) {
	        throw Error(
                    `Invalid argument #${i} in function ${callee.name}`
                )
            }
        }
    }
}


function* map_lazy (to_be_mapped, f) {
    check(map, arguments, { to_be_mapped: Object, f: Function })
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
        let new_key = f(key, value)
        if ( !result.has(new_key) ) {
            result[new_key] = hash[key]
        } else {
            throw Error('mapkey(): Key conflict detected')
        }        
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


function pick (hash, key, concept, f) {
    check(
        pick, arguments,
        { hash: Hash, key: Str, concept: Concept, f: Function }
    )
    if (hash.is(NA)) return NA;
    if ( hash.has(key) && concept.contains(hash[key]) ) {
        return f(hash[key])
    } else {
        return NA
    }
}


function take (...operation_list) {    
    assert(operation_list.is(ArrayOf(Function)) )
    for ( let f of operation_list ) {
        let result = f()
        if ( result !== NA ) {
            return result
        }
    }
    throw Error('take(): All operations produce N/A result')
}


function pour (target, source) {
    check(pour, arguments, { target: Hash, source: Hash })
    for ( let key of Object.keys(source) ) {
        target[key] = source[key]
    }
}


function fold (iterable, initial, f) {
    check(
        fold, arguments,
        { iterable: Iterable, initial: Any, f: Function }
    )
    var value = initial
    for ( let element of iterable ) {
        value = f(element, value)
    }
    return value
}


function forall (iterable, f) {
    check(forall, arguments, { iterable: Iterable, f: Function })
    return fold(iterable, true, (e,v) => v && f(e))
}


function exists (iterable, f) {
    check(forall, arguments, { iterable: Iterable, f: Function })
    return fold(iterable, false, (e,v) => v || f(e))
}


Object.prototype.has = function (prop) { return this.hasOwnProperty(prop) }
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
