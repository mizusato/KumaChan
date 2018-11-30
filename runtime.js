'use strict';


/**
 *  Exception Definition
 */


class RuntimeException extends Error {    
    static gen_err_msg (err_type, func_name, err_msg) {
        return `${func_name}: ${err_type}: ${err_msg}`
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
class NameConflict extends RuntimeException {}


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
const CopyFlag = {
    reference: '&',
    value: '*',
    default: '',
    native: '~'
}
const CopyFlagValue = fold(
    Object.keys(CopyFlag), {},
    (key,v) => (v[CopyFlag[key]] = key, v)
)


/**
 *  Object Type Definition
 * 
 *  Object ┬ Hash
 *         ┴ Simple ┬ List
 *                  ┼ Atomic
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


const AtomicNameSet = new Set()


function AtomicObject(name) {    
    NameConflict.assert(
        !AtomicNameSet.has(name),
        'AtomicObject()', 'atomic object name ${name} is in use'
    )
    return pour({}, { name: name, maker: AtomicObject })
}


SetMakerConcept(AtomicObject)


const SimpleObject = $u(PrimitiveObject, ListObject, AtomicObject)


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


const NullScope = AtomicObject('NullScope')


SetEquivalent(NullScope, $1(NullScope))


const NotNullScope = $n(HashObject, Struct({
    config: Struct({
        context: $u( NullScope, $(x => NotNullScope.contains(x)) )
    })
}))


function Scope (context) {
    return pour(HashObject(), {
        config: {
            context: context
        }
    })
}


SetEquivalent(Scope, $u(NullScope, NotNullScope))


const G = Scope(NullScope)
const K = G.data
K.global = G


const NaObject = K['N/A'] = AtomicObject('N/A')


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


function ConceptObject (concept_name, f) {
    check(ConceptObject, arguments, { f: Function })
    return pour(HashObject(), {
        config: {
            name: concept_name,
            contains: ConceptFunctionInstance(`${concept_name}::Checker`, f)
        }
    })
}


SetEquivalent(ConceptObject, $u(
    $f(AnyConcept, BoolConcept),
    $n(HashObject, Struct({
        config: Struct({
            name: StringObject,
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
    $( proto => forall(proto.order, key => proto.parameters.has(key)) )
)


FunctionPrototype.represent = function repr_prototype (prototype) {
    check(repr_prototype, arguments, { prototype: FunctionPrototype })
    function repr_parameter (key, parameter) {
        check(repr_parameter, arguments, { key: Str, parameter: Parameter })
        let type = parameter.constraint.config.name
        let flags = CopyFlag[parameter.pass_policy]
        return `${type} ${flags}${key}`
    }
    function opt (string, non_empty_callback) {
        check(opt, arguments, { string: Str, non_empty_callback: Function })
        return string && non_empty_callback(string) || ''
    }
    let order = prototype.order
    let parameters = prototype.parameters
    let retval_constraint = prototype.return_value
    let necessary = Enum.apply({}, order)
    return '(' + join(filter([
        join(map(
            order,
            key => repr_parameter(key, parameters[key])
        ), ', '),
        opt(join(map(
            filter(parameters, key => key.is_not(necessary)),
            (key, val) => repr_parameter(key, val)
        ), ', '), s => `[${s}]`),
    ], x => x), ', ') + ') -> ' + `${retval_constraint.config.name}`
}


FunctionPrototype.parse = function parse_prototype (string) {
    const pattern = /\((.*)\) -> (.*)/
    const sub_pattern = /(.*), *[(.*)]/
    check(parse_prototype, arguments, { string: Regex(pattern) })
    let str = {
        parameters: string.match(pattern)[1].trim(),
        return_value: string.match(pattern)[2].trim()
    }
    let has_optional = str.parameters.match(sub_pattern)
    str.parameters = {
        necessary: has_optional? has_optional[1].trim(): str.parameters,
        all: str.parameters
    }
    function check_concept (string) {
        if ( K.has(string) && K[string].is(ConceptObject) ) {
            return K[string]
        } else {
            throw Error('prototype parsing error: invalid constraint')
        }
    }
    function parse_parameter (string) {
        const pattern = /([^ ]*) *([\*\&\~]?)(.+)/
        check(parse_parameter, arguments, { string: Regex(pattern) })
        let str = {
            constraint: string.match(pattern)[1].trim(),
            pass_policy: string.match(pattern)[2].trim()
        }
        let name = string.match(pattern)[3]
        return { key: name, value: mapval(str, function (s, item) {
            return ({
                constraint: () => check_concept(s),
                pass_policy: () => CopyFlagValue[s] || 'default'
            })[item]()
        }) }
    }
    let trim_all = x => map_lazy(x, y => y.trim())
    return {
        order: map(map(
            trim_all(str.parameters.necessary.split(',')),
            s => parse_parameter(s)
        ), p => p.key),
        parameters: fold(map(
            trim_all(str.parameters.all.split(',')),
            s => parse_parameter(s)
        ), {}, (e,v) => (v[e.key]=e.value, v)),
        return_value: check_concept(str.return_value)
    }
}


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
    /*
      [TODO]
    let _ = suppose_items([
        () => a,
        () => b
    ])
    if ( _.is(Failed) ) { return _ }
    */
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
        context: Scope,
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
                return CheckReturnValue(
                    proto, f(Scope(context), arg),
                    this.return_value_promised, debug_name
                )
            },
            toString: function () {
                let proto_repr = FunctionPrototype.represent(this.prototype)
                return `${this.name} ${proto_repr}`
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
        name, G, ConceptFunctionPrototype, function (scope, argument) {
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


function PortEquivalent(hash_object, concept, name) {
    check(
        PortEquivalent, arguments,
        { hash_object: HashObject, concept: Concept, name: Str }
    )
    pour(hash_object.config, {
        name: name,
        contains: ConceptFunctionInstance(
            `${name}::Checker`, x => x.is(concept)
        )
    })
}


PortEquivalent(AnyConcept, ObjectObject, 'Any')
PortEquivalent(BoolConcept, BoolObject, 'Bool')
BoolConcept.config.contains.return_value_promised = true


function PortConcept(concept, name) {
    check(PortConcept, arguments, { concept: Concept, name: Str })
    return ConceptObject(name, x => x.is(concept))
}


const ConceptConcept = K.Concept = PortConcept(ConceptObject, 'Concept')
const NumberConcept = K.Number = PortConcept(NumberObject, 'Number')
const StringConcept = K.String = PortConcept(StringObject, 'String')
const PrimitiveConcept = K.Primitive = PortConcept(PrimitiveObject, 'Primitive')
const NonPrimitiveConcept = K.NonPrimitive = PortConcept(NonPrimitiveObject, 'NonPrimitive')
const ListConcept = K.List = PortConcept(ListObject, 'List')
const AtomicConcept = K.Atomic = PortConcept(AtomicObject, 'Atomic')
const SimpleConcept = K.Simple = PortConcept(SimpleObject, 'Simple')
const HashConcept = K.Hash = PortConcept(HashObject, 'Hash')
const ObjectConcept = K.Object = K.Any


const VoidValue = K.VoidValue = AtomicObject('VoidValue')
const VoidObject = () => VoidValue
SetEquivalent(VoidObject, $1(VoidValue))
const VoidConcept = K.Void = PortConcept(VoidObject, 'Void')


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
            },
            toString: function() {
                return join(map(this.instances, I => I.toString()), '\n')
            }
        })
    }
}


const FunctionInstanceConcept = K.FunctionInstance = PortConcept(FunctionInstanceObject, 'FunctionInstance')
const FunctionConcept = K.Function = PortConcept(FunctionObject, 'Function')


function CreateInstance (name_and_proto, js_function) {
    let name = name_and_proto.split(' ')[0]
    let prototype = name_and_proto.slice(name.length, name_and_proto.length)
    return FunctionInstanceObject(
        name, G, FunctionPrototype.parse(prototype), js_function
    )
}


pour(K, {
    is: FunctionObject('is', [
        CreateInstance(
            'Any::is (Any ~self, Concept ~concept) -> Bool',
            (s, a) => a.concept.config.contains.apply(a.self)
        )
    ])
})


pour(K, {
    ref_copy: FunctionObject('ref_copy', [
        CreateInstance (
            'NonPrimitive::ref_copy (NonPrimitive ~self) -> NonPrimitive',
            (s, a) => a.self
        )
    ]),
    val_copy: FunctionObject('val_copy', [
        CreateInstance (
            'Primitive::val_copy (Primitive ~self) -> Primitive',
            (s, a) => a.self
        ),
        CreateInstance (
            'List::val_copy (List ~self) -> List',
            (s, a) => map(a.self, e => K.copy.apply(e))
        )
    ]),
    copy: FunctionObject('copy', [
        CreateInstance (
            'Any::copy (Any ~self) -> Any',
            (s, a) => a.self
        )
    ])
})


pour(K, {
    has_data: FunctionObject('has_data', [
        CreateInstance (
            'Hash::has_data (Hash ~self, String ~key) -> Bool',
            (s, a) => a.self.data.has(a.key)
        )
    ]),
    get_data: FunctionObject('get_data', [
        CreateInstance (
            'Hash::get_data (Hash ~self, String ~key) -> Any',
            (s, a) => KeyError.assert(
                a.self.data.has(a.key), 'Hash::get_data', `'${a.key}'`
            ) && a.self.data[a.key]
        )
    ]),
    find_data: FunctionObject('find_data', [
        CreateInstance (
            'Hash::find_data (Hash ~self, String ~key) -> Any',
            (s, a) => a.self.data.has(a.key) && a.self.data[a.key] || NaObject
        )
    ]),
    set_data: FunctionObject('set_data', [
        CreateInstance (
            'Hash::set_data (Hash ~self, String ~key, Any ~value) -> Void',
            (s, a) => (a.self.data[a.key] = a.value, VoidValue)
        )
    ])
})
