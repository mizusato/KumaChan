'use strict';


/**
 *  Exception Definition
 */


class RuntimeException extends Error {    
    static gen_err_msg (err_type, func_name, err_msg) {
        return `[Runtime Exception] ${func_name}: ${err_type}: ${err_msg}`
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
class NoMatchingPattern extends RuntimeException {}
class KeyError extends RuntimeException {}


/**
 *  Enumeration Definition
 */


const CopyPolicy = Enum('reference', 'value', 'default', 'native')
const CopyAction = {
    reference: x => K.ref_copy.apply(x),
    value: x => K.val_copy.apply(x),
    default: x => K.copy.apply(x),
    native: x => x
}


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
const NonPrimitiveObject = $_(PrimitiveObject)


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


function CheckArgument(prototype, argument, debug_name = '') {
    /**
     *  Check if argument is valid.
     *  If valid, returns normalized argument object.
     *  Normalize: Index Number -> Key; Execute Value Copy
     */
    check(
        CheckArgument, arguments,
        { prototype: FunctionPrototype, argument: Hash }
    )
    let proto = prototype
    let parameters = proto.parameters
    let order = proto.order
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
    argument = mapkey(
        argument,
        key => key.is(NumStr)? order[key]: key
    )
    map(parameters, function (key, parameter) {
        if ( argument.has(key) && parameter.constraint !== AnyConcept ) {
            let contains = parameter.constraint.config.contains
            InvalidArgument.assert(
                contains.apply(argument[key]),
                debug_name, `illegal argument ${key}`
            )            
        }
    })
    argument = mapval(
        argument,
        (val, key) => CopyAction[parameters[key].pass_policy](val)
    )
    return argument
}


function CheckReturnValue (prototype, value, promised = false, debug_name = '') {
    /**
     *  Check if return value is legal.
     *  If legal, returns the raw return value.
     */
    check(
        CheckReturnValue, arguments,
        { prototype: FunctionPrototype, value: Any }
    )
    let value_set = prototype.return_value
    if (value_set !== AnyConcept && !promised) {
        InvalidReturnValue.assert(
            value_set.config.contains.apply(value),
            debug_name, `invalid return value ${value}`
        )
    }
    return value
}


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
                assert(ArrayOf(ObjectObject).contains(args))
                return this.call(fold(args, {}, (e, v, i) => (v[i] = e, v)) )
            },
            call: function (argument) {
                assert(HashOf(ObjectObject).contains(argument))
                let debug_name = `${this.name}()`
                let proto = this.prototype
                let context = this.context
                let f = this.js_function
                let arg = CheckArgument(proto, argument, debug_name)
                arg.argument = pour(HashObject(), { data: arg })
                return CheckReturnValue(
                    proto, f(context, arg),
                    this.return_value_promised, debug_name
                )
            }
        })
    })
}


SetMakerConcept(FunctionInstanceObject)


const ConceptFunctionPrototype = {
    parameters: {
        object: {
            constraint: AnyConcept,
            pass_policy: 'native'
        }
    },
    order: ['object'],
    return_value: BoolConcept
}


function ConceptFunctionInstance (name, f) {
    check(ConceptFunctionInstance, arguments, { name: Str, f: Function })
    return FunctionInstanceObject(
        name, G, ConceptFunctionPrototype, function (context, argument) {
            return f(argument.object)
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
                pass_policy: $f('reference', 'native')
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
const NonPrimitiveConcept = K.NonPrimitive = PortConcept(NonPrimitiveObject, 'NonPrimitive_Checker')
const ListConcept = K.List = PortConcept(ListObject, 'List_Checker')
const SimpleConcept = K.Simple = PortConcept(SimpleObject, 'Simple_Checker')
const HashConcept = K.Hash = PortConcept(HashObject, 'Hash_Checker')
const ObjectConcept = K.Object = K.Any


const VoidValue = K.VoidValue = HashObject()
const VoidObject = () => VoidValue
SetEquivalent(VoidObject, $1(VoidValue))
const VoidConcept = K.Void = PortConcept(VoidObject, 'Void_Checker')


/**
 *  Function Definition
 */


function FunctionObject (name, instances) {
    check(FunctionObject, arguments, {
        name: Str,
        instances: $n(
            ArrayOf(FunctionInstanceObject),
            $(array => array.length > 0)
        )
    })
    return {
        name: name,
        instances: instances,
        __proto__: once(FunctionObject, {
            add: function (instance) {
                assert(FunctionInstanceObject.contains(instance))
                this.instances.push(instance)
            },
            apply: function (...args) {
                assert(ArrayOf(ObjectObject).contains(args))
                return this.call(fold(args, {}, (e, v, i) => (v[i] = e, v)) )
            },
            call: function (argument) {
                assert(HashOf(ObjectObject).contains(argument))
                for(let instance of rev(this.instances)) {
                    try {
                        let arg = CheckArgument(instance.prototype, argument)
                        return instance.call(arg)
                    } catch (err) {
                        if (err instanceof InvalidArgument) {
                            continue
                        } else {
                            throw err
                        }
                    }
                }
                NoMatchingPattern.assert(
                    false, `${this.name}()`,
                    'invalid call: cannot find matching function prototype'
                )
            }
        })
    }
}


const FunctionInstanceConcept = K.FunctionInstance = PortConcept(FunctionInstanceObject, 'FunctionInstance_Checker')
const FunctionConcept = K.Function = PortConcept(FunctionObject, 'Function_Checker')


K.ref_copy = FunctionObject('ref_copy', [
    FunctionInstanceObject('NonPrimitive::ref_copy', G, {
        parameters: {
            self: {
                constraint: NonPrimitiveConcept,
                pass_policy: 'native'
            }
        },
        order: ['self'],
        return_value: NonPrimitiveConcept
    }, (c,a) => a.self)
])


K.val_copy = FunctionObject('val_copy', [
    FunctionInstanceObject('Primitive::val_copy', G, {
        parameters: {
            self: {
                constraint: PrimitiveConcept,
                pass_policy: 'native'
            }
        },
        order: ['self'],
        return_value: PrimitiveConcept
    }, (c,a) => a.self),
    FunctionInstanceObject('List::val_copy', G, {
        parameters: {
            self: {
                constraint: ListConcept,
                pass_policy: 'native'
            }
        },
        order: ['self'],
        return_value: ListConcept
    }, (c,a) => map(a.self, e => K.copy.apply(e)) )
])


K.copy = FunctionObject('copy', [
    FunctionInstanceObject('Any::copy', G, {
        parameters: {
            self: {
                constraint: AnyConcept,
                pass_policy: 'native'
            }
        },
        order: ['self'],
        return_value: AnyConcept
    }, (c,a) => a.self)
])


K.is = FunctionObject('is', [
    FunctionInstanceObject('Any::is', G, {
        parameters: {
            self: {
                constraint: AnyConcept,
                pass_policy: 'native'
            },
            concept: {
                constraint: ConceptConcept,
                pass_policy: 'native'
            }
        },
        order: ['self', 'concept'],
        return_value: BoolConcept
    }, (c,a) => a.concept.config.contains.apply(a.self) )
])


K.has_data = FunctionObject('has_data', [
    FunctionInstanceObject('Hash::has_data', G, {
        parameters: {
            self: {
                constraint: HashConcept,
                pass_policy: 'native'
            },
            key: {
                constraint: StringConcept,
                pass_policy: 'native'
            }
        },
        order: ['self', 'key'],
        return_value: BoolConcept
    }, (c,a) => a.self.data.has(a.key))
])


K.get_data = FunctionObject('get_data', [
    FunctionInstanceObject('Hash::get_data', G, {
        parameters: {
            self: {
                constraint: HashConcept,
                pass_policy: 'native'
            },
            key: {
                constraint: StringConcept,
                pass_policy: 'native'
            }
        },
        order: ['self', 'key'],
        return_value: AnyConcept
    }, (c,a) => (
        KeyError.assert(a.self.data.has(a.key), 'Hash::get_data', `'${a.key}'`)
            && a.self.data[a.key]
    ))
])


K.set_data = FunctionObject('set_data', [
    FunctionInstanceObject('Hash::set_data', G, {
        parameters: {
            self: {
                constraint: HashConcept,
                pass_policy: 'native'
            },
            key: {
                constraint: StringConcept,
                pass_policy: 'native'
            },
            value: {
                constraint: AnyConcept,
                pass_policy: 'native'
            }
        },
        order: ['self', 'key', 'value'],
        return_value: VoidConcept
    }, (c,a) => (a.self.data[a.key] = a.value, VoidValue))
])
