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
}


var I = x => x
var $n = Concept.intersect
var $u = Concept.union
var $_ = Concept.complement
var Any = new Concept(() => true)
var None = new Concept(() => false)
var Num = new Concept(x => typeof x == 'number')
var Str = new Concept(x => typeof x == 'string')
var Hash = new Concept(x => x instanceof Object)
var Iterable = new Concept(x => typeof x[Symbol.iterator] == 'function')
Object.prototype.is = function (concept) { return concept.contains(this) }
Object.prototype.is_not = function (concept) { return !this.is(concept) }
Function.prototype.contains = function (obj) { return obj instanceof this }
Array.prototype.contains = function (obj) { return this.indexOf(obj) != -1 }


function check(callee, concept_table) {
    var parameters = callee.get_parameters()
    var parameter_to_index = {}
    var index = 0
    for ( let parameter of parameters ) {
        parameter_to_index[paramter] = index
        index++
    }
    if ( concept_table.is(Hash) ) {
        for ( let parameter of Object.keys(type_table) ) {
	    var argument = callee.arguments[parameter_to_index[parameter]]
            var concept = concept_table[parameter]
            if (argument.is_not(concept)) {
	        throw Error(
                    `Invalid argument ${parameter} in function ${callee.name}`
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


function map (to_be_mapped, f) {
    check(map, { to_be_mapped: Object, f: Function })
    if( to_be_mapped.is(Iterable) ) {
	let iterable = to_be_mapped
	let result = []
	let index = 0
	for ( let I of iterable ) {
	    result.push(f(I, index))
	    index += 1
	}
	return result
    } else {
	let hash = to_be_mapped
	let result = []
	for ( let key of Object.keys(hash) ) {
	    result.push(f(key, hash[key]))
	}
	return result
    }
}


function mapval (hash, f) {
    check(mapval, { hash: Hash, f: Function })
    var result = {}
    for ( let key of Object.keys(hash) ) {
        let value = hash[key]
        result[key] = f(value)
    }
    return result
}


function* cat (...iterables) {
    // TODO
    //check(cat, Iterable)
    for( let iterable of iterables ) {
	check_type([[iterable, ['object', 'string']]])
	check_iterable(iterable)
	for ( let element of iterable ) {
	    yield element	
	}
    }
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


