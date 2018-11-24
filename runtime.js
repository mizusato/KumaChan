'use strict';


/**
 *  Exception Definition
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
        } else {
            return true
        }
    }
}


class InvalidOperation extends RuntimeException {}
class InvalidArgument extends RuntimeException {}
class InvalidReturnValue extends RuntimeException {}
class KeyError extends RuntimeException {}


/**
 *  Enumeration Definition
 */


const CopyPolicy = Enum('reference', 'value')


/**
 *  Object Type Definition
 * 
 *  Object ┬ Hash
 *         ┴ Simple ┬ List
 *                  ┴ Primitive ┬ String
 *                              ┼ Number
 *                              ┴ Bool
 *
 *  Note: non-hash object is called simple object.
 */


const StringObject = $(x => typeof x == 'string')
const NumberObject = $(x => typeof x == 'number')
const BoolObject = $(x => typeof x == 'boolean')
const PrimitiveObject = $u(StringObject, NumberObject, BoolObject)


function ListObject() {
    return pour([], { maker: ListObject })
}


SetMakerConcept(ListObject)


const SimpleObject = $u(PrimitiveObject, ListObject)


function HashObject () {
    return {
        data: {},
        config: {},
        maker: HashObject
    }
}


SetMakerConcept(HashObject)


const ObjectObject = $u(SimpleObject, HashObject)


/**
 *  Global Object Definition
 */


const G = HashObject()
const K = G.data
K.global = G


/**
 *  Concept & Function Instance Definition
 *  
 *  We have to deal AnyConcept and BoolConcept especially.
 *  It's because a Concept is defined by its checker function,
 *  so we have to build a Function Instance before building a concept,
 *  and it is necessary to build a Function Prototype before building
 *  this Function Instance, the prototype can be described as
 *  f: (object::Any) -> Bool, which requires the definition of Any and Bool.
 */


const AnyConcept = K.Any = HashObject()
const BoolConcept = K.Bool = HashObject()


function ConceptObject (f_name, f) {
    check(ConceptObject, arguments, { f: Function })
    return pour(HashObject(), {
        config: {
            contains: ConceptFunctionInstance(f_name, f)
        }
    })
}


SetEquivalent(ConceptObject, $u(
    $f(AnyConcept, BoolConcept),
    $n(HashObject, Struct({
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
        name: name || '[Anonymous]',
        context: context,
        prototype: prototype,
        js_function: js_function,
        maker: FunctionInstanceObject,
        __proto__: once(FunctionInstanceObject, {
            apply: function (...args) {
                return this.call(fold(args, {}, (e, v, i) => (v[i] = e, v)) )
            },
            call: function (argument) {
                let debug_name = `${this.name}()`
                let proto = this.prototype
                let parameters = proto.parameters
                let order = proto.order
                let context = this.context
                let f = this.js_function
                mapkey(argument, key => InvalidArgument.assert(
                    !(key.is(NumStr) && order.has_no(key)),
                    debug_name, `redundant argument ${key}`
                ))
                mapkey(argument, key => InvalidArgument.assert(
                    !(key.is(NumStr) && argument.has(order[key])),
                    debug_name, `conflict argument ${key}`
                ))
                map(order, (key, index) => InvalidArgument.assert(
                    argument.has(index) || argument.has(key),
                    debug_name, `missing argument ${key}`
                ))
                argument = mapkey(argument, key => key.is(NumStr)? order[key]: key)
                map(order, function (key) {
                    let parameter = parameters[key]
                    if ( parameter.constraint !== AnyConcept ) {
                        let contains = parameter.constraint.config.contains
                        InvalidArgument.assert(
                            contains.apply(argument[key]), debug_name,
                            `illegal value ${argument[key]} for argument ${key}`
                        )
                    }
                    // TODO: copy value
                })
                let value = f(this.context, argument)
                let value_set = proto.return_value
                if (value_set !== AnyConcept && !this.return_value_promised) {
                    let contains = value_set.config.contains
                    InvalidReturnValue.assert(
                        contains.apply(value),
                        debug_name, `invalid return value ${value}`
                    )
                }
                return value
            }
        })
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
        name, G, ConceptFunctionPrototype, function (context, argument) {
            return js_concept_function(argument.object)
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


function PortEquivalent(hash_object, concept, f_name) {
    check(
        PortEquivalent, arguments,
        { hash_object: HashObject, concept: Concept, f_name: Str }
    )
    hash_object.config.contains = (
        ConceptFunctionInstance(
            f_name, x => x.is(concept)
        )   
    )
}


PortEquivalent(AnyConcept, ObjectObject, 'Any_Checker')
PortEquivalent(BoolConcept, BoolObject, 'Bool_Checker')
BoolConcept.config.contains.return_value_promised = true


function PortConcept(concept, f_name) {
    check(PortConcept, arguments, { concept: Concept, f_name: Str })
    var r = HashObject()
    PortEquivalent(r, concept, f_name)
    return r
}


const ConceptConcept = K.Concept = PortConcept(ConceptObject, 'Concept_Checker')
const NumberConcept = K.Number = PortConcept(NumberObject, 'Number_Checker')
const StringConcept = K.String = PortConcept(StringObject, 'String_Checker')
const PrimitiveConcept = K.Primitive = PortConcept(PrimitiveObject, 'Primitive_Checker')
const ListConcept = K.List = PortConcept(ListObject, 'List_Checker')
const SimpleConcept = K.Simple = PortConcept(SimpleObject, 'Simple_Checker')
const HashConcept = K.Hash = PortConcept(HashObject, 'Hash_Checker')
const ObjectConcept = K.Object = K.Any


/**
 *  Function Definition
 */


function FunctionObject (function_instances) {

}
