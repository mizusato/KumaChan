'use strict';


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

// --------------------------------------------------------------------------- 

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


// --------------------------------------------------------------------------- 


const K = HashObject()
K.global = K


// --------------------------------------------------------------------------- 


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


// --------------------------------------------------------------------------- 


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


SetMakerConcept(String)


function NumberObject (number = 0) {
    assert(Num.contains(number))
    var object = {
        is_hash_object: true,
        is_atomic: true,
        default_copy_policy: 'value',
        number: number
    }
    return object
}


SetMakerConcept(Number)


// --------------------------------------------------------------------------- 


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


// --------------------------------------------------------------------------- 


const AnyConcept = K.Any = HashObject({ is_atomic: true })


function ConceptObject ( concept_function_instance ) {
    check()
    var object = HashObject()
    object.config.contains = concept_function_instance
}


ConceptObject = $u(
    $1(AnyConcept),
    $n(HashObject, Struct({
        config: Struct({
            contains: ConceptFunctionInstance
        })
    }))
)


const Parameter = Struct({
    constraint: ConceptObject,
    pass_policy: CopyPolicy
})


const FunctionPrototype = $n(Struct({
    parameters: HashOf(Parameter),
    order: ArrayOf(Str),
    return_value: ConceptObject
}), d => forall(d.order, name => parameters.has(name)) )


function FunctionInstanceObject (context, prototype, js_function) {
    check(FunctionInstance, arguments, {
        context: HashObject,
        prototype: FunctionPrototype,
        js_function: Function
    })
    return pour(HashObject(), {
        context: context,
        prototype: prototype,
        js_function: js_function
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
    return_value: BoolObject
}


function ConceptFunctionInstance (js_concept_function) {
    return FunctionInstance(
        K, ConceptFunctionPrototype, function (context, argument) {
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
                return_value: BoolObject
            }),
            $(proto => proto.parameter[proto.order[0]].is(Struct({
                constraint: $1(AnyConcept),
                pass_policy: $1('reference')
            })))
        )
    }))
)


function FunctionObject (function_instances) {

}
