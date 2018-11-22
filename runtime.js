'use strict';


var K = {}


class RuntimeException extends Error {    
    static gen_err_msg (err_type, func_name, err_msg) {
        return `[Runtime Exception] ${err_type}: ${func_name}: ${err_msg}`
    }    
    static assert (bool, function_name, error_message) {
        if (!bool) {
            throw new this(
                this.gen_err_msg(
                    (this.name == 'RuntimeException')?
                        'Error': this.name.replace(/([a-z])([A-Z])/, '$1 $2'),
                    function_name,
                    error_message
                )
            )
        }
    }
}


class InvalidOperation extends RuntimeException {}
class KeyError extends RuntimeException {}


const CopyPolicy = Enum('reference', 'value')


function HashObject (argument = {}) {
    assert(Hash.contains(argument))
    check_hash(HashObject, argument, { atomic: Optional(Bool, false) })
    var object = {
        is_hash_object: true,
        is_atomic: argument.atomic,
        copy_policy: 'reference'
    }
    if (!argument.atomic) {
        object.data = {}
    }
    return object
}


function StringObject (string = '') {
    assert(String.contains(string))
    var object = {
        is_hash_object: true,
        is_atomic: true,
        copy_policy: 'value'
    }
}

HashObject.contains = x => x.is_hash_object === true
