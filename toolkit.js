'use strict';


function list2hash (string_list, get_value = (() => true)) {
    var result = {}
    for ( let key of string_list ) {
	result[key] = get_value(key)
    }
    return result
}


function check_type(type_table) {
    for ( let item of type_table ) {
	var argument = item[0]
	var type_list = item[1]
	if ( !(typeof argument in list2hash(type_list)) ) {
	    throw Error('Invalid argument')
	}
    }
}


function check_iterable(argument) {
    if ( !argument.is_iterable() ) {
	throw Error('Invalid argument')
    }
}


function map (to_be_mapped, f) {
    check_type([
	[to_be_mapped, ['object', 'string']],
	[f, ['function']]
    ])
    if( to_be_mapped.is_iterable() ) {
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
    check_type([
        [hash, ['object']],
        [f, ['function']]
    ])
    var result = {}
    for ( let key of Object.keys(hash) ) {
        let value = hash[key]
        result[key] = f(value)
    }
    return result
}


function* cat (...iterables) {
    check_type([[iterables, ['object']]])
    check_iterable(iterables)
    for( let iterable of iterables ) {
	check_type([[iterable, ['object', 'string']]])
	check_iterable(iterable)
	for ( let element of iterable ) {
	    yield element	
	}
    }
}


Object.prototype.is_iterable = function () {
    return (typeof this[Symbol.iterator] == 'function')
}
