'use strict';


/**
 * Exceptions Definition
 */


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
class InvalidArgument extends RuntimeException {}
class InvalidReturnValue extends RuntimeException {}
class KeyError extends RuntimeException {}


/**
 * Enumerations Definition
 */


const CopyPolicy = Enum('reference', 'value')


/**
 * Basic Object Definition
 */


function HashObject (argument = {}) {
    assert(Hash.contains(argument))
    check_hash(HashObject, argument, { is_atomic: Optional(Bool, false) })
    var object = {
        is_hash_object: true,
        is_atomic: argument.is_atomic,
        default_copy_policy: 'reference',
        manufacturer: HashObject
    }
    if (!argument.is_atomic) {
        object.data = {}
        object.config = {}
    }
    return object
}


HashObject.contains = (x => x.is_hash_object === true)
const Atomic = $n(HashObject, $(x => x.is_atomic))
const NonAtomic = $n(HashObject, $(x => !x.is_atomic))


/**
 * Global Object Definition
 */


const K = HashObject()
K.global = K

// TODO: primitive type refactor

/**
 * Boolean Value Definition
 */


const True = K.data.True = HashObject({ is_atomic: true })
const False = K.data.False = HashObject({ is_atomic: true })
const Unknown = K.data.Unknown = HashObject({ is_atomic: true })


function BoolObject(bool = false) {
    assert(Bool.contains(bool))
    return bool && True || False
}


SetEquivalent(BoolObject, $f(True, False))


function ExtBoolObject(value = 'unknown') {
    assert($u(Bool, $1('unknown')).contains(value))
    if (value == 'unknown') {
        return Unknown
    } else {
        return BoolObject(value)
    }
}


SetEquivalent(ExtBoolObject, $f(True, False, Unknown))


/**
 * String & Number Definition
 */


function StringObject (string = '') {
    assert(Str.contains(string))
    var object = {
        is_hash_object: true,
        is_atomic: true,
        default_copy_policy: 'value',
        string: string,
        manufacturer: StringObject
    }
    return object
}


function SetMakerConcept (manufacturer) {
    SetEquivalent(
        manufacturer, $n(HashObject, $(x => x.made_by(manufacturer)))
    )
}


SetMakerConcept(StringObject)


function NumberObject (number = 0) {
    assert(Num.contains(number))
    var object = {
        is_hash_object: true,
        is_atomic: true,
        default_copy_policy: 'value',
        number: number,
        manufacturer: NumberObject
    }
    return object
}


SetMakerConcept(NumberObject)


/**
 * List Definition
 */


function ListObject (list = []) {
    assert(ArrayOf(HashObject).contains(list))
    var object = {
        is_hash_object: true,
        is_atomic: true,
        default_copy_policy: 'value',
        list: list,
        manufacturer: ListObject
    }
    return object
}


SetMakerConcept(ListObject)


/**
 * Concept Definition
 */


const AnyConcept = K.Any = HashObject({ is_atomic: true })
const BoolConcept = K.Bool = HashObject()


function ConceptObject ( f ) {
    check(ConceptObject, arguments, { f: Function })
    return pour(HashObject(), {
        config: {
            contains: ConceptFunctionInstance('', f)
        }
    })
}


SetEquivalent(ConceptObject, $u(
    $f(AnyConcept, BoolConcept),
    $n(NonAtomic, Struct({
        config: Struct({
            contains: ConceptFunctionInstance
        })
    }))
))


const Parameter = Struct({
    constraint: ConceptObject,
    pass_policy: CopyPolicy
})


const FunctionPrototype = $n(
    Struct({
        parameters: HashOf(Parameter),
        order: ArrayOf(Str),
        return_value: ConceptObject
    }),
    $(proto => forall(proto.order, name => proto.parameters.has(name)))
)


function FunctionInstanceObject (name, context, prototype, js_function) {
    check(FunctionInstanceObject, arguments, {
        name: Str,
        context: HashObject,
        prototype: FunctionPrototype,
        js_function: Function
    })
    return pour(HashObject(), {
        name: name,
        context: context,
        prototype: prototype,
        js_function: js_function,
        manufacturer: FunctionInstanceObject,
        call: function (argument) {
            let f_name = `${this.name}()`
            let proto = this.prototype
            argument = mapkey(argument, function (key) {
                let n = Number(key)
                if (!Number.isNaN(n)) {
                    let str_name = proto.order[n]
                    InvalidArgument.assert(
                        typeof str_name != 'undefined',
                        f_name, `redundant argument ${key}`
                    )
                    InvalidArgument.assert(
                        !argument.has(str_name),
                        f_name, `missing argument ${key}`
                    )
                    return str_name
                } else {
                    return key
                }
            })
            map(proto.order, function (name) {
                let parameter = proto.parameters[name]
                if ( parameter.constraint !== AnyConcept ) {
                    InvalidArgument.assert(
                        True === parameter.constraint.config.contains.call({
                            '0': argument[name]
                        }),
                        f_name, `invalid argument ${name}`
                    )
                }
                // TODO: copy value
            })
            let return_value = js_function(this.context, argument)
            if (proto.return_value !== AnyConcept
                && !this.return_value_promised) {
                InvalidReturnValue.assert(
                    True === proto.return_value.config.contains.call({
                        '0': return_value
                    }),
                    f_name, `invalid return value`
                )
            }
            return return_value
        }
    })
}


SetMakerConcept(FunctionInstanceObject)


const ConceptFunctionPrototype = {
    parameters: {
        object: {
            constraint: AnyConcept,
            pass_policy: 'reference'            
        }
    },
    order: ['object'],
    return_value: BoolConcept
}


function ConceptFunctionInstance (name, js_concept_function) {
    return FunctionInstanceObject(
        name, K, ConceptFunctionPrototype, function (context, argument) {
            return BoolObject(js_concept_function(argument.object))
        }
    )
}


SetEquivalent(
    ConceptFunctionInstance,
    $n(FunctionInstanceObject, Struct({
        prototype: $n(
            Struct({
                order: $(array => array.length == 1),
                return_value: $1(BoolConcept)
            }),
            $(proto => proto.parameters[proto.order[0]].is(Struct({
                constraint: $1(AnyConcept),
                pass_policy: $1('reference')
            })))
        )
    }))
)


const ConceptConcept = K.Concept = ConceptObject(x => x.is(ConceptObject))


BoolConcept.config.contains = ConceptFunctionInstance(
    'Bool', x => BoolObject.contains(x)
)
BoolConcept.config.contains.return_value_promised = true


const NumberConcept = K.Number = pour(HashObject(), {
    config: {
        contains: ConceptFunctionInstance(
            'Number', x => NumberObject.contains(x)
        )
    }
})


const StringConcept = K.String = pour(HashObject(), {
    config: {
        contains: ConceptFunctionInstance(
            'String', x => StringObject.contains(x)
        )
    }
})




function FunctionObject (function_instances) {

}
