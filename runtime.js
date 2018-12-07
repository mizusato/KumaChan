'use strict';


/**
 *  Exception Definition
 */


class RuntimeError extends Error {}
class InvalidOperation extends RuntimeError {}
class InvalidArgument extends RuntimeError {}
class InvalidReturnValue extends RuntimeError {}
class NoMatchingPattern extends RuntimeError {}
class KeyError extends RuntimeError {}
class NameConflict extends RuntimeError {}


function ErrorProducer (err_class, f_name) {
    check(ErrorProducer, arguments, {
        err_class: $(x => x.prototype instanceof Error),
        f_name: Str
    })
    return {
        if: function (bool, err_msg) {
            check(this.if, arguments, { bool: Bool, err_msg: Str })
            if ( bool ) {
                let err_type = err_class.name.replace(
                    /([a-z])([A-Z])/, '$1 $2'
                )
                throw new err_class(`${f_name}: ${err_type}: ${err_msg}`)
            }
        },
        unless: function (bool, err_msg) {
            check(this.unless, arguments, { bool: Bool, err_msg: Str })
            return this.if(!bool, err_msg)
        },
        throw: function (err_msg) {
            return this.if(true, err_msg)
        },
        if_failed: function (result) {
            check(this.if_failed, arguments, { result: Result })
            if ( result.is(Failed) ) {
                this.if(true, result.message)
            }
        }
    }
}


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


function ListObject () {
    return pour([], { maker: ListObject })
}


SetMakerConcept(ListObject)


const AtomicNameSet = new Set()


function AtomicObject (name) {
    let err = ErrorProducer(NameConflict, 'AtomicObject()')
    err.if(AtomicNameSet.has(name), 'atomic object name ${name} is in use')
    return pour({}, { name: name, maker: AtomicObject })
}


SetMakerConcept(AtomicObject)


const SimpleObject = $u(PrimitiveObject, ListObject, AtomicObject)


function HashObject () {
    return {
        data: {},
        config: {},
    }
}


pour(HashObject, {
    create_by_data: data => pour(HashObject(), { data: data }),
    create_by_config: config => pour(HashObject(), { config: config })
})


SetEquivalent(HashObject, Struct({
    data: Hash,
    config: Hash
}))


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


function Scope (context, data = {}) {
    assert(Scope.contains(context))
    assert(Hash.contains(data))
    return pour(HashObject(), {
        data: data,
        config: {
            context: context
        }
    })
}


SetEquivalent(Scope, $u(NullScope, NotNullScope))


const G = Scope(NullScope)
const K = G.data
const scope = G
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
    return HashObject.create_by_config({
        name: concept_name,
        contains: ConceptFunctionInstance(`${concept_name}::Checker`, f)
    })
}


function UnionConceptObject (concept1, concept2) {
    check(UnionConceptObject, arguments, {
        concept1: ConceptObject,
        concept2: ConceptObject
    })
    let name = `(${concept1.config.name} | ${concept2.config.name})`
    let f = x => exists(
        map_lazy([concept1, concept2], c => c.config.contains),
        f => f.apply(x) === true
    )
    return HashObject.create_by_config({
        name: name,
        contains: ConceptFunctionInstance(`${name}::Checker`, f)
    })
}


function IntersectConceptObject (concept1, concept2) {
    check(IntersectConceptObject, arguments, {
        concept1: ConceptObject,
        concept2: ConceptObject
    })
    let name = `(${concept1.config.name} & ${concept2.config.name})`
    let f = x => forall(
        map_lazy([concept1, concept2], c => c.config.contains),
        f => f.apply(x) === true
    )
    return HashObject.create_by_config({
        name: name,
        contains: ConceptFunctionInstance(`${name}::Checker`, f)
    })
}


function ComplementConceptObject (concept) {
    check(ComplementConceptObject, arguments, {
        concept: ConceptObject,
    })
    let name = `!${concept.config.name}`
    let f = x => concept.config.contains.apply(x) === false
    return HashObject.create_by_config({
        name: name,
        contains: ConceptFunctionInstance(`${name}::Checker`, f)
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


FunctionPrototype.check_argument = function (prototype, argument) {
    check( FunctionPrototype.check_argument, arguments, {
        prototype: FunctionPrototype, argument: Hash
    })
    let proto = prototype
    let parameters = proto.parameters
    let order = proto.order
    return need (
        cat(
            map_lazy(Object.keys(argument), key => suppose(
                !(key.is(NumStr) && order.has_no(key)),
                `redundant argument ${key}`
            )),
            map_lazy(Object.keys(argument), key => suppose(
                !(key.is(NumStr) && argument.has(order[key])),
                `conflict argument ${key}`
            )),
            map_lazy(order, (key, index) => suppose(
                argument.has(index) || argument.has(key),
                `missing argument ${key}`
            )),
            lazy(function () {
                let arg = mapkey(
                    argument,
                    key => key.is(NumStr)? order[key]: key
                )
                return map_lazy(parameters, (key, p) => suppose(
                    !arg.has(key)
                        || (p.constraint === AnyConcept
                            && ObjectObject.contains(arg[key]))
                        || p.constraint.config.contains.apply(arg[key]),
                    `illegal argument '${key}'`
                ))
            })
        )
    )
}


FunctionPrototype.normalize_argument = function (prototype, argument) {
    check( FunctionPrototype.normalize_argument, arguments, {
        prototype: FunctionPrototype, argument: Hash
    })
    return mapval(
        mapkey(argument, key => key.is(NumStr)? prototype.order[key]: key),
        (val, key) => CopyAction[prototype.parameters[key].pass_policy](val)
    )
}


FunctionPrototype.check_return_value = function (prototype, value) {
    check( FunctionPrototype.check_return_value, arguments, {
        prototype: FunctionPrototype, value: Any
    })
    if (prototype.return_value !== AnyConcept) {
        return suppose(
            prototype.return_value.config.contains.apply(value),
            `invalid return value ${value}`
        )
    } else {
        return OK
    }
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
                let err = ErrorProducer(InvalidArgument, `${this.name}()`)
                let err_r = ErrorProducer(InvalidReturnValue, `${this.name}()`)
                let proto = this.prototype
                let context = this.context
                let f = this.js_function
                let p = FunctionPrototype
                err.if_failed(p.check_argument(proto, argument))
                let normalized_argument = p.normalize_argument(proto, argument)
                let scope = Scope(context, {
                    argument: HashObject.create_by_data(normalized_argument)
                })
                scope.data.scope = scope
                pour(scope.data, normalized_argument)
                let value = f(scope)
                if (!this.return_value_promised) {
                    err_r.if_failed(p.check_return_value(proto, value))
                }
                return value
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
        name, G, ConceptFunctionPrototype, function (scope) {
            return f(scope.data.argument.data.object)
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
const IntConcept = K.Int = PortConcept(Int, 'Int')
const UnsignedIntConcept = K.UnsignedInt = PortConcept(UnsignedInt, 'UnsignedInt')
const IndexConcept = K.Index = K.UnsignedInt
const SizeConcept = K.Size = K.UnsignedInt
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
    return pour(HashObject(), {
        name: name,
        instances: instances,
        maker: FunctionObject,
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
                    let p = FunctionPrototype
                    let check = p.check_argument(instance.prototype, argument)
                    if ( check === OK ) {
                        return instance.call(argument)
                    }
                }
                let err = ErrorProducer(NoMatchingPattern, `${this.name}()`)
                let msg = 'invalid call: matching function prototype not found'
                err.throw(msg)
            },
            has_method_of: function (object) {
                return exists(
                    map_lazy(this.instances, I => I.prototype),
                    p => (p.order.length > 0)
                        && (p.parameters[p.order[0]]
                            .constraint.config.contains.apply(object))
                )
            },
            toString: function () {
                return join(map(this.instances, I => I.toString()), '\n')
            }
        })
    })
}


SetMakerConcept(FunctionObject)


const HasMethod = (...names) => $(
    x => assert(x.is(ObjectObject))
        && forall(names, name => K.has(name) && K[name].is(FunctionObject)
                  && K[name].has_method_of(x))
)


const FunctionInstanceConcept = K.FunctionInstance = PortConcept(FunctionInstanceObject, 'FunctionInstance')
const FunctionConcept = K.Function = PortConcept(FunctionObject, 'Function')


function CreateInstance (name_and_proto, js_function) {
    let name = name_and_proto.split(' ')[0]
    let prototype = name_and_proto.slice(name.length, name_and_proto.length)
    return FunctionInstanceObject(
        name, G, FunctionPrototype.parse(prototype), function (scope) {
            return js_function (scope.data.argument.data)
        }
    )
}


pour(K, {
    is: FunctionObject('is', [
        CreateInstance(
            'Any::is (Any ~self, Concept ~concept) -> Bool',
            a => a.concept.config.contains.apply(a.self)
        )
    ]),
    union: FunctionObject('union', [
        CreateInstance(
            'union (Concept ~concept1, Concept ~concept2) -> Concept',
            a => UnionConceptObject(a.concept1, a.concept2)            
        )
    ]),
    intersect: FunctionObject('intersect', [
        CreateInstance(
            'intersect (Concept ~concept1, Concept ~concept2) -> Concept',
            a => IntersectConceptObject(a.concept1, a.concept2)
        )
    ]),
    complement: FunctionObject('complement', [
        CreateInstance(
            'complement (Concept ~concept) -> Concept',
            a => ComplementConceptObject(a.concept)
        )
    ]),
    difference: FunctionObject('difference', [
        CreateInstance(
            'difference (Concept ~concept1, Concept ~concept2) -> Concept',
            a => IntersectConceptObject(
                a.concept1, ComplementConceptObject(a.concept2)
            )
        )
    ])
})


pour(K, {
    ref_copy: FunctionObject('ref_copy', [
        CreateInstance (
            'NonPrimitive::ref_copy (NonPrimitive ~self) -> NonPrimitive',
            a => a.self
        )
    ]),
    val_copy: FunctionObject('val_copy', [
        CreateInstance (
            'Primitive::val_copy (Primitive ~self) -> Primitive',
            a => a.self
        ),
        CreateInstance (
            'List::val_copy (List ~self) -> List',
            a => map(a.self, e => K.copy.apply(e))
        )
    ]),
    copy: FunctionObject('copy', [
        CreateInstance (
            'Any::copy (Any ~self) -> Any',
            a => a.self
        )
    ])
})


pour(K, {
    plus: FunctionObject('plus', [
        CreateInstance (
            'plus (Number ~p, Number ~q) -> Number',
            a => a.p + a.q
        ),
        CreateInstance (
            'plus (String ~string1, String string2) -> String',
            a => a.string1 + a.string2
        )
    ]),
    minus: FunctionObject('minus', [
        CreateInstance (
            'minus (Number ~p, Number ~q) -> Number',
            a => a.p - a.q
        ),
        CreateInstance (
            'minus (Number ~x) -> Number',
            a => -a.x
        )
    ]),
    multiply: FunctionObject('multiply', [
        CreateInstance (
            'multiply (Number ~p, Number ~q) -> Number',
            a => a.p * a.q
        )
    ]),
    divide: FunctionObject('divide', [
        CreateInstance (
            'divide (Number ~p, Number ~q) -> Number',
            a => a.p / a.q
        )
    ]),
    pow: FunctionObject('pow', [
        CreateInstance (
            'pow (Number ~p, Number ~q) -> Number',
            a => Math.pow(a.p, a.q)
        )
    ]),
    is_finite: FunctionObject('is_finite', [
        CreateInstance (
            'Number::is_finite (Number ~self) -> Bool',
            a => Number.isFinite(a.self)
        )
    ]),
    is_NaN: FunctionObject('is_NaN', [
        CreateInstance (
            'Number::is_NaN (Number ~self) -> Bool',
            a => Number.isNaN(a.self)
        )
    ]),
    exp: FunctionObject('exp', [
        CreateInstance (
            'exp (Number ~x) -> Number',
            a => Math.exp(a.x)
        )
    ]),
    log: FunctionObject('log', [
        CreateInstance (
            'log (Number ~x) -> Number',
            a => Math.log(a.x)
        )
    ]),
    floor: FunctionObject('floor', [
        CreateInstance (
            'floor (Number ~x) -> Number',
            a => Math.floor(a.x)
        )
    ]),
    ceil: FunctionObject('ceil', [
        CreateInstance (
            'ceil (Number ~x) -> Number',
            a => Math.ceil(a.x)
        )
    ]),
    round: FunctionObject('round', [
        CreateInstance (
            'round (Number ~x) -> Number',
            a => Math.round(a.x)
        )
    ]),
    sin: FunctionObject('sin', [
        CreateInstance (
            'sin (Number ~x) -> Number',
            a => Math.sin(a.x)
        )
    ]),
    atan2: FunctionObject('atan2', [
        CreateInstance (
            'atan2 (Number ~y, Number ~x) -> Number',
            a => Math.atan2(a.y, a.x)
        )
    ])
})


const HasSlice = HasMethod('at', 'length')


function SliceObject (object, start, end) {
    check(SliceObject, arguments, {
        object: HasSlice, start: Int, end: Int
    })
    let err = ErrorProducer(KeyError, 'Slice::Creator')
    let length = K.length.apply(object)
    let normalize = index => (index < 0)? length+index: index
    start = normalize(start)
    end = normalize(end)
    // todo end == infinity
    err.if(start > end, 'start position greater than end position')
    err.unless(0 <= start && start < length, 'invalid start position')
    err.unless(0 <= end && end <= length, 'invalid end position')
    return pour(HashObject(), {
        data: {
            object: object,
            start: start,
            end: end
        },
        config: {
            name: 'Slice'
        }
    })
}


SetEquivalent(SliceObject, $n(HashObject, Struct({
    data: Struct({
        object: HasSlice,
        start: UnsignedInt,
        end: UnsignedInt
    })
})))


const SliceConcept = K.Slice = PortConcept(SliceObject, 'Slice')
const HasSliceConcept = K.HasSlice = PortConcept(HasSlice, 'HasIndex')


const UCS2Char = $n(Str, $(x => x.length == 1))
const CharConcept = K.Char = PortConcept(UCS2Char, 'Char')


pour(K, {
    at: FunctionObject('at', [
        CreateInstance (
            'List::at (List ~self, Index ~index) -> Any',
            a => a.index < a.self.length
                && a.self[a.index]
                || ErrorProducer(KeyError, 'List::at').throw(`${a.index}`)
        ),
        CreateInstance (
            'String::at (String ~self, Index ~index) -> Char',
            a => a.index < a.self.length
                && a.self[a.index]
                || ErrorProducer(KeyError, 'String::at').throw(`${a.index}`)
        ),
        CreateInstance (
            'Slice::at (Slice ~self, Index ~index) -> Any',
            a => a.index < K.length.apply(a.self)
                && K.at.apply(a.self.data.object, a.self.data.start+a.index)
                || ErrorProducer(KeyError, 'Slice::at').throw(`${a.index}`)
        )
    ]),
    real_at: FunctionObject('real_at', [
        CreateInstance (
            'String::real_at (String ~self, Index ~index) -> String',
            a => a.self.realCharAt(a.index)
              || ErrorProducer(KeyError, 'String::real_at').throw(`${a.index}`)
        )
    ]),
    length: FunctionObject('length', [
        CreateInstance (
            'List::length (List ~self) -> Size',
            a => a.self.length
        ),
        CreateInstance (
            'String::length (String ~self) -> Size',
            a => a.self.length
        ),
        CreateInstance (
            'Slice::length (Slice ~self) -> Size',
            a => (a.self.data.end - a.self.data.start)
        )
    ]),
    genuine_length: FunctionObject('genuine_length', [
        CreateInstance (
            'String::genuine_length (String ~self) -> Size',
            a => a.self.genuineLength()
        )
    ]),
    slice: FunctionObject('slice', [
        CreateInstance (
            'HasSlice::slice (HasSlice ~self, Int ~start, Int ~end) -> Slice',
            a => SliceObject(a.self, a.start, a.end)
        )
    ]),
    append: FunctionObject('append', [
        CreateInstance (
            'List::append (List ~self, Any ~element) -> Void',
            a => (a.self.push(a.element), VoidValue)
        )
    ])
})


pour(K, {
    has_data: FunctionObject('has_data', [
        CreateInstance (
            'Hash::has_data (Hash ~self, String ~key) -> Bool',
            a => a.self.data.has(a.key)
        )
    ]),
    get_data: FunctionObject('get_data', [
        CreateInstance (
            'Hash::get_data (Hash ~self, String ~key) -> Any',
            a => a.self.data.has(a.key)
                && a.self.data[a.key]
                || ErrorProducer(KeyError, 'Hash::get_data').throw(`${a.key}`)
        )
    ]),
    find_data: FunctionObject('find_data', [
        CreateInstance (
            'Hash::find_data (Hash ~self, String ~key) -> Any',
            a => a.self.data.has(a.key) && a.self.data[a.key] || NaObject
        )
    ]),
    set_data: FunctionObject('set_data', [
        CreateInstance (
            'Hash::set_data (Hash ~self, String ~key, Any ~value) -> Void',
            a => (a.self.data[a.key] = a.value, VoidValue)
        )
    ])
})
