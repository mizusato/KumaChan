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
class RangeError extends RuntimeError {}
class NameConflict extends RuntimeError {}
class InvalidFunctionInstance extends RuntimeError {}
class DataViewOutOfDate extends RuntimeError {}


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
 *  Enumerations Definition
 */


const CopyPolicy = Enum('dirty', 'immutable', 'value')
const CopyAction = {
    dirty: x => assert(x.is(MutableObject)) && x,
    immutable: x => Im(x),
    value: x => K.val_copy.apply(x) // TODO: instances of val_copy() must be pure
}
const CopyFlag = {
    dirty: '&',
    value: '*',
    immutable: ''
}
const CopyFlagValue = fold(
    Object.keys(CopyFlag), {},
    (key, v) => (v[CopyFlag[key]] = key, v)
)


const EffectRange = Enum('global', 'nearby', 'local')
const Restrictions = Enum('pure')
const RestrictionChecker = {
    pure: f => f.is(PureFunctionInstance)
}


/**
 *  Object Type Definition
 * 
 *  Object (Any) ┬ Compound ┬ Hash
 *               │          ┴ List
 *               ┴ Atomic ┬ FunctionInstance
 *                        ┼ Function
 *                        ┼ Singleton
 *                        ┴ Primitive ┬ String
 *                                    ┼ Number
 *                                    ┴ Bool
 *
 *  Note: atomic objects are immutable.
 */


/* Primitive Definition */


const StringObject = $(x => typeof x == 'string')
const NumberObject = $(x => typeof x == 'number')
const BoolObject = $(x => typeof x == 'boolean')
const PrimitiveObject = $u(StringObject, NumberObject, BoolObject)


/* Singleton Definition */


const SingletonNameSet = new Set()


function SingletonObject (name) {
    let err = ErrorProducer(NameConflict, 'SingletonObject()')
    err.if(SingletonNameSet.has(name), 'singleton name ${name} is in use')
    let singleton = {}
    return pour(singleton, {
        data: {
            name: name,
            checker: ConceptChecker(
                `Singleton(${name})`,
                a => a.object === singleton
            )
        },
        contains: x => x === singleton,
        maker: SingletonObject
    })
}


SetMakerConcept(SingletonObject)


/* Atomic Definition */


const AtomicObject = $u(
    SigletonObject, PrimitiveObject,
    $(x => x.is(FunctionInstanceObject)),
    $(x => x.is(FunctionObject))
)


/* Hash Definition */


function HashObject (hash = {}) {
    assert(hash.is(Hash))
    return {
        data: hash,
        config: { immutable: false },
        maker: HashObject
    }
}


SetMakerConcept(HashObject)


HashObject.immutable = hash => pour(
    HashObject(hash),
    { config: { immutable: true } }
)


const ImHashObject = $n(HashObject, $(x => x.config.immutable))
const MutHashObject = $n(HashObject, $(x => !x.config.immutable))


/* List Definition */


function ListObject (list = []) {
    assert(list.is(Array))
    return {
        data: list,
        config: { immutable: false },
        maker: ListObject
    }
}


SetMakerConcept(ListObject)


ListObject.immutable = list => pour(
    ListObject(list),
    { config: { immutable: true } }
)


const ImListObject = $n(ListObject, $(x => x.config.immutable))
const MutListObject = $n(ListObject, $(x => !x.config.immutable))


const ListObjectOf = concept => $n(
    ListObject,
    $(l => l.data.is(ArrayOf(concept)))
)


/* Compound Definition */


const CompoundObject = $u(HashObject, ListObject)


/* Object (Any) Definition */


const ObjectObject = $u(CompoundObject, AtomicObject)


/**
 * Mutable and Immutable Definition
 */


const MutableObject = $n(CompoundObject, $(x => !x.config.immutable))
const ImmutableObject = $_(MutableObject)


function Im (object) {
    if ( object.is(MutableObject) ) {
        return {
            data: object.data,
            config: pour(pour({}, object.config), {
                immutable: true
            }),
            maker: object.maker
        }
    } else {
        return object
    }
}


/**
 *  Scope Definition
 */


const NullScope = SingletonObject('NullScope')


SetEquivalent(NullScope, $1(NullScope))


const NotNullScope = $n(HashObject, Struct({
    context: $u( NullScope, $(x => x.is(NotNullScope)) )
}))


function Scope (context, data = {}) {
    assert(context.is(Scope))
    assert(data.is(Hash))
    return pour(HashObject(), {
        data: data,
        context: context
    })
}


SetEquivalent(Scope, $u(NullScope, NotNullScope))


/**
 *  Global Object Definition
 */


const G = Scope(NullScope)
const K = G.data
K.scope = G


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


const AnyConcept = HashObject()
const BoolConcept = HashObject()


function ConceptObject (concept_name, f) {
    check(ConceptObject, arguments, { f: Function })
    return HashObject.immutable({
        name: concept_name,
        checker: ConceptChecker(`${concept_name}`, f)
    })
}


function UnionConceptObject (concept1, concept2) {
    check(UnionConceptObject, arguments, {
        concept1: ConceptObject,
        concept2: ConceptObject
    })
    let name = `(${concept1.data.name} | ${concept2.data.name})`
    let f = x => exists(
        map_lazy([concept1, concept2], c => c.data.checker),
        f => f.apply(x) === true
    )
    return HashObject.immutable({
        name: name,
        checker: ConceptChecker(`${name}`, f)
    })
}


function IntersectConceptObject (concept1, concept2) {
    check(IntersectConceptObject, arguments, {
        concept1: ConceptObject,
        concept2: ConceptObject
    })
    let name = `(${concept1.data.name} & ${concept2.data.name})`
    let f = x => forall(
        map_lazy([concept1, concept2], c => c.data.checker),
        f => f.apply(x) === true
    )
    return HashObject.immutable({
        name: name,
        checker: ConceptChecker(`${name}`, f)
    })
}


function ComplementConceptObject (concept) {
    check(ComplementConceptObject, arguments, {
        concept: ConceptObject,
    })
    let name = `!${concept.data.name}`
    let f = x => concept.data.checker.apply(x) === false
    return HashObject.immutable({
        name: name,
        checker: ConceptChecker(`${name}`, f)
    })
}


//  ConceptObject = {Any, Bool} 
//                    | ( Immutable
//                          & (Hash | Singleton | Function)
//                          & ${ data: ${ ... } } )
SetEquivalent(ConceptObject, $u(
    $f(AnyConcept, BoolConcept),
    $n(
        ImmutableObject,
        $u(HashObject, SingletonObject, FunctionObject),
        Struct({
            data: Struct({
                name: StringObject,
                checker: ConceptChecker
            })
        })
    )
))


const Parameter = Struct({
    constraint: ConceptObject,
    pass_policy: CopyPolicy
})


const FunctionPrototype = $n(
    Struct({
        effect_range: EffectRange,
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
        let type = parameter.constraint.data.name
        let flags = CopyFlag[parameter.pass_policy]
        return `${type} ${flags}${key}`
    }
    function opt (string, non_empty_callback) {
        check(opt, arguments, { string: Str, non_empty_callback: Function })
        return string && non_empty_callback(string) || ''
    }
    let effect = prototype.effect_range
    let order = prototype.order
    let parameters = prototype.parameters
    let retval_constraint = prototype.return_value
    let necessary = Enum.apply({}, order)
    return effect + ' ' + '(' + join(filter([
        join(map(
            order,
            key => repr_parameter(key, parameters[key])
        ), ', '),
        opt(join(map(
            filter(parameters, key => key.is_not(necessary)),
            (key, val) => repr_parameter(key, val)
        ), ', '), s => `[${s}]`),
    ], x => x), ', ') + ') -> ' + `${retval_constraint.data.name}`
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
        const pattern = /([^ ]*) *([\*\&]?)(.+)/
        check(parse_parameter, arguments, { string: Regex(pattern) })
        let str = {
            constraint: string.match(pattern)[1].trim(),
            pass_policy: string.match(pattern)[2].trim()
        }
        let name = string.match(pattern)[3]
        return { key: name, value: mapval(str, function (s, item) {
            return ({
                constraint: () => check_concept(s),
                pass_policy: () => CopyFlagValue[s]
            })[item]()
        }) }
    }
    let trim_all = x => map_lazy(x, y => y.trim())
    return {
        effect_range: string.split(' ')[0],
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
                        || p.constraint.data.checker.apply(arg[key]),
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
            prototype.return_value.data.checker.apply(value),
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
    return {
        name: name || '[Anonymous]',
        context: context,
        prototype: prototype,
        js_function: js_function,
        maker: FunctionInstanceObject,
        __proto__: once(FunctionInstanceObject, {
            apply: function (...args) {
                assert(args.is(ArrayOf(ObjectObject)))
                return this.call(fold(args, {}, (e, v, i) => (v[i] = e, v)) )
            },
            call: function (argument) {
                assert(argument.is(HashOf(ObjectObject)))
                let err = ErrorProducer(InvalidArgument, `${this.name}()`)
                let err_r = ErrorProducer(InvalidReturnValue, `${this.name}()`)
                let proto = this.prototype
                let context = this.context
                let f = this.js_function
                let p = FunctionPrototype
                err.if_failed(p.check_argument(proto, argument))
                let normalized_argument = p.normalize_argument(proto, argument)
                let scope = Scope(context, {
                    argument: HashObject(normalized_argument)
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
    }
}


SetMakerConcept(FunctionInstanceObject)


const PureFunctionInstance = $n(Struct({
    effect_range: 'local',
}), $(f => forall(f.prototype.parameters, p => p.pass_policy != 'dirty')) )


const ConceptFunctionPrototype = {
    effect_range: 'local',
    parameters: {
        object: {
            constraint: AnyConcept,
            pass_policy: 'immutable'
        }
    },
    order: ['object'],
    return_value: BoolConcept
}


function ConceptChecker (name, f) {
    check(ConceptChecker, arguments, { name: Str, f: Function })
    return FunctionInstanceObject(
        `${name}.checker`, G, ConceptFunctionPrototype, function (scope) {
            return f(scope.data.argument.data.object)
        }
    )
}


SetEquivalent(
    ConceptChecker,
    $n(FunctionInstanceObject, Struct({
        prototype: $n(
            Struct({
                effect_range: 'local',
                order: $(array => array.length == 1),
                return_value: $1(BoolConcept)
            }),
            $(proto => proto.parameters[proto.order[0]].is(Struct({
                constraint: $1(AnyConcept),
                pass_policy: $1('immutable')
            })))
        )
    }))
)


function PortEquivalent(hash_object, concept, name) {
    check(
        PortEquivalent, arguments,
        { hash_object: HashObject, concept: Concept, name: Str }
    )
    pour(hash_object.data, {
        name: name,
        checker: ConceptChecker(
            `${name}`, x => x.is(concept)
        )
    })
    hash_object.config.immutable = true
}


PortEquivalent(AnyConcept, ObjectObject, 'Any')
PortEquivalent(BoolConcept, BoolObject, 'Bool')
BoolConcept.data.checker.return_value_promised = true


function CreateInstance (name_and_proto, js_function) {
    let name = name_and_proto.split(' ')[0]
    let prototype = name_and_proto.slice(name.length, name_and_proto.length)
    return FunctionInstanceObject(
        name, G, FunctionPrototype.parse(prototype), function (scope) {
            return js_function (scope.data.argument.data)
        }
    )
}


/**
 *  Function Definition
 */


function FunctionObject (name, instances, restrictions, equivalent_concept) {
    check(FunctionObject, arguments, {
        name: Str,
        instances: $n(
            ArrayOf(FunctionInstanceObject),
            $(array => array.length > 0)
        ),
        restrictions: Optional(ArrayOf(Restriction)),
        equivalent_concept: Optional(ConceptObject)
    })
    restrictions = restrictions || []
    concept_name = (
        equivalent_concept?
        equivalent_concept.data.name:
        'no equivalent concept'
    )
    concept_checker = (
        equivalent_concept?
        equivalent_concept.data.checker:
        'no concept checker'
    )
    assert(forall(instances, I => forall(
        restrictions,
        r => RestrictionChecker[r](I)
    )))
    return {
        name: name,
        instances: instances,
        restrictions: restrictions,
        maker: FunctionObject,
        data: {
            name: concept_name,
            checker: concept_checker
        }
        __proto__: once(FunctionObject, {
            has_restriction: function (restriction) {
                return this.restrictions.indexOf(restriction) != -1
            },
            is_valid_instance (instance) {
                return forall(
                    this.restrictions,
                    r => RestrictionChecker[r](instance)
                )
            },
            check_valid_instance (instance) {
                return need(map_lazy(this.restrictions, r => suppose(
                    RestrictionChecker[r](instance),
                    `restriction ${r} not satisfied`
                )))
            },
            added: function (instance) {
                assert(instance.is(FunctionInstanceObject))
                let err = ErrorProducer(InvalidFunctionInstance, 'Function')
                err.if_failed(check_valid_instance(instance))
                let new_list = map(this.instances, x=>x)
                new_list.push(instance)
                return FunctionObject(
                    this.name,
                    new_list,
                    this.restrictions
                )
            },            
            apply: function (...args) {
                assert(args.is(ArrayOf(ObjectObject)))
                return this.call(fold(args, {}, (e, v, i) => (v[i] = e, v)) )
            },
            call: function (argument) {
                assert(argument.is(HashOf(ObjectObject)))
                for(let instance of rev(this.instances)) {
                    let p = FunctionPrototype
                    let check = p.check_argument(instance.prototype, argument)
                    if ( check === OK ) {
                        return instance.call(argument)
                    }
                }
                let err = ErrorProducer(NoMatchingPattern, `${this.data.name}()`)
                let msg = 'invalid call: matching function prototype not found'
                msg += '\n' + 'available instances are:' + '\n'
                msg += this.toString()
                err.throw(msg)
            },
            has_method_of: function (object) {
                return exists(
                    map_lazy(this.instances, I => I.prototype),
                    p => (p.order.length > 0)
                        && (p.parameters[p.order[0]]
                            .constraint.data.checker.apply(object))
                )
            },
            toString: function () {
                return join(map(this.instances, I => I.toString()), '\n')
            }
        })
    }
}


SetMakerConcept(FunctionObject)


const HasMethod = (...names) => $(
    x => assert(x.is(ObjectObject))
        && forall(names, name => K.has(name) && K[name].is(FunctionObject)
                  && K[name].has_method_of(x))
)


/**
 *  Port Native Concepts
 */


function PortConcept(concept, name) {
    check(PortConcept, arguments, { concept: Concept, name: Str })
    return ConceptObject(name, x => x.is(concept))
}


const SingletonConcept = FunctionObject('Singleton', [
    CreateInstance (
        'local Singleton (String name) -> Singleton',
        a => SingletonObject(a.name)
    )
], [], PortConcept(SingletonObject, 'Singleton'))


pour(K, {
    /* concept */
    Concept: PortConcept(ConceptObject, 'Concept'),
    /* special */
    Any: AnyConcept,
    Bool: BoolConcept,
    /* primitive */
    Number: PortConcept(NumberObject, 'Number'),
    Int: PortConcept(Int, 'Int')
    UnsignedInt: PortConcept(UnsignedInt, 'UnsignedInt')
    String: PortConcept(StringObject, 'String'),
    Primitive: PortConcept(PrimitiveObject, 'Primitive'),
    /* non-primitive atomic */
    FunctionInstance: PortConcept(FunctionInstanceObject, 'FunctionInstance'),
    Function: PortConcept(FunctionObject, 'Function'),
    Singleton: SingletonConcept,
    'N/A': SingletonObject('N/A'),
    'Void': SingletonObject('Void'),
    Atomic: PortConcept(AtomicObject, 'Atomic'),
    /* compound */
    List: PortConcept(ListObject, 'List'),
    Hash: PortConcept(HashObject, 'Hash'),
    Compound: PortConcept(CompoundObject, 'Compound'),
    /* mutable or not */
    ImHash: PortConcept(ImHashObject, 'ImHash'),
    MutHash: PortConcept(MutHashObject, 'MutHash'),
    ImList: PortConcept(ImListObject, 'ImList'),
    MutList: PortConcept(MutListObject, 'MutList'),
    Immutable: PortConcept(ImmutableObject, 'Immutable')
    Mutable: PortConcept(MutableObject, 'Mutable')
})


/* concept alias */


pour(K, {
    Object: K.Any,
    Index: K.UnsignedInt,
    Size: K.UnsignedInt
})


/**
 *  Fundamental Functions Definition
 */


pour(K, {
    is: FunctionObject('is', [
        CreateInstance(
            'local Any::is (Any self, Concept concept) -> Bool',
            a => a.concept.data.checker.apply(a.self)
        )
    ]),
    union: FunctionObject('union', [
        CreateInstance(
            'local union (Concept concept1, Concept concept2) -> Concept',
            a => UnionConceptObject(a.concept1, a.concept2)            
        )
    ]),
    intersect: FunctionObject('intersect', [
        CreateInstance(
            'local intersect (Concept concept1, Concept concept2) -> Concept',
            a => IntersectConceptObject(a.concept1, a.concept2)
        )
    ]),
    complement: FunctionObject('complement', [
        CreateInstance(
            'local complement (Concept concept) -> Concept',
            a => ComplementConceptObject(a.concept)
        )
    ])
})


pour(K, {
    dirty_copy: FunctionObject('dirty_copy', [
        CreateInstance (
            'local Mutable::dirty_copy (Mutable &self) -> Mutable',
            a => a.self
        )
    ]),
    immutable_copy: FunctionObject('immutable_copy', [
        CreateInstance (
            'local Any::immutable_copy (Any self) -> Immutable',
            a => a.self
        )
    ]),
    copy: FunctionObject('copy', [
        CreateInstance (
            'local Any::copy (Any &self) -> Object',
            a => a.self.is(MutableObject)?
                CopyAction.dirty(a.self):
                CopyAction.immutable(a.self)
        )
    ])
    val_copy: FunctionObject('val_copy', [
        CreateInstance (
            'local Primitive::val_copy (Primitive self) -> Primitive',
            a => a.self
        ),
        CreateInstance (
            'local List::val_copy (List self) -> List',
            a => ListObject(
                map(a.self.data, e => CopyAction.immutable(e))
            )
        )
    ], ['pure']),
})


pour(K, {
    plus: FunctionObject('plus', [
        CreateInstance (
            'local plus (Number p, Number q) -> Number',
            a => a.p + a.q
        ),
        CreateInstance (
            'local plus (String s1, String s2) -> String',
            a => a.s1 + a.s2
        )
    ]),
    minus: FunctionObject('minus', [
        CreateInstance (
            'local minus (Number p, Number q) -> Number',
            a => a.p - a.q
        ),
        CreateInstance (
            'local minus (Number x) -> Number',
            a => -a.x
        ),
        CreateInstance(
            'local minus (Concept concept1, Concept concept2) -> Concept',
            a => IntersectConceptObject(
                a.concept1, ComplementConceptObject(a.concept2)
            )
        )
    ]),
    multiply: FunctionObject('multiply', [
        CreateInstance (
            'local multiply (Number p, Number q) -> Number',
            a => a.p * a.q
        )
    ]),
    divide: FunctionObject('divide', [
        CreateInstance (
            'local divide (Number p, Number q) -> Number',
            a => a.p / a.q
        )
    ]),
    mod: FunctionObject('mod', [
        CreateInstance (
            'local mod (Number p, Number q) -> Number',
            a => a.p % a.q
        )
    ]),
    pow: FunctionObject('pow', [
        CreateInstance (
            'local pow (Number p, Number q) -> Number',
            a => Math.pow(a.p, a.q)
        )
    ]),
    is_finite: FunctionObject('is_finite', [
        CreateInstance (
            'local Number::is_finite (Number self) -> Bool',
            a => Number.isFinite(a.self)
        )
    ]),
    is_NaN: FunctionObject('is_NaN', [
        CreateInstance (
            'local Number::is_NaN (Number self) -> Bool',
            a => Number.isNaN(a.self)
        )
    ]),
    floor: FunctionObject('floor', [
        CreateInstance (
            'local floor (Number x) -> Number',
            a => Math.floor(a.x)
        )
    ]),
    ceil: FunctionObject('ceil', [
        CreateInstance (
            'local ceil (Number x) -> Number',
            a => Math.ceil(a.x)
        )
    ]),
    round: FunctionObject('round', [
        CreateInstance (
            'local round (Number x) -> Number',
            a => Math.round(a.x)
        )
    ])
})


const HasSlice = $( x => Im(x).is(HasMethod('at', 'length')) )


function SliceObject (object, start, end) {
    check(SliceObject, arguments, {
        object: HasSlice, start: Int, end: Int
    })
    let err = ErrorProducer(KeyError, 'Slice::Creator')
    let length = K.length.apply(object)
    let normalize = index => (index < 0)? length+index: index
    start = normalize(start)
    end = normalize(end)
    if ( end == 0 ) { end = length }
    // todo end == infinity
    err.if(start > end, 'start position greater than end position')
    err.unless(0 <= start && start < length, 'invalid start position')
    err.unless(0 <= end && end <= length, 'invalid end position')
    return pour(HashObject(), {
        data: {
            object: object,
            start: start,
            end: end,
            length: length
        },
        config: {
            name: 'Slice',
            comment: `[${start}, ${end}) of [0, ${length})`,
            immutable: true
        }
    })
}


SetEquivalent(SliceObject, $n(ImHashObject, Struct({
    data: Struct({
        object: $n(Immutable, HasSlice),
        start: UnsignedInt,
        end: UnsignedInt,
        length: UnsignedInt
    })
})))


K.Slice = PortConcept(SliceObject, 'Slice')
K.HasSlice = PortConcept(HasSlice, 'HasSlice')


const UCS2Char = $n(Str, $(x => x.length == 1))
K.Char = PortConcept(UCS2Char, 'Char')


const ListOperations = {
    at: a => a.index < a.self.data.length
                && a.self.data[a.index]
                || ErrorProducer(KeyError, 'List::at').throw(`${a.index}`),
    append: a => (a.self.data.push(a.element), K.Void)
}


pour(K, {
    at: FunctionObject('at', [
        CreateInstance (
            'local ImList::at (ImList self, Index index) -> Immutable',
            ListOperations.at
        ),
        CreateInstance (
            'local MutList::at (MutList &self, Index index) -> Object',
            ListOperations.at
        ),
        CreateInstance (
            'local String::at (String self, Index index) -> Char',
            a => a.index < a.self.length
                && a.self[a.index]
                || ErrorProducer(RangeError, 'String::at').throw(`${a.index}`)
        ),
        CreateInstance (
            'Slice::at (Slice self, Index index) -> Immutable',
            function (a) {
                let current_obj_length = K.length.apply(a.self.data.object)
                let slice_length = K.length.apply(a.self)
                let e_range = ErrorProducer(RangeError, 'Slice::at')
                let e_invalid = ErrorProducer(DataViewOutOfDate, 'Slice::at')
                e_range.if(a.index >= slice_length,`${a.index}`)
                e_invalid.if(
                    a.self.data.length != current_obj_length,
                    'length of object changed'
                )
                return K.at.apply(a.self.data.object, a.self.data.start+a.index)
            }
        )
    ]),
    real_at: FunctionObject('real_at', [
        CreateInstance (
            'local String::real_at (String self, Index index) -> String',
            a => a.self.realCharAt(a.index)
              || ErrorProducer(RangeError, 'String::real_at').throw(`${a.index}`)
        )
    ]),
    length: FunctionObject('length', [
        CreateInstance (
            'local List::length (List self) -> Size',
            a => a.self.data.length
        ),
        CreateInstance (
            'local String::length (String self) -> Size',
            a => a.self.length
        ),
        CreateInstance (
            'local Slice::length (Slice self) -> Size',
            a => (a.self.data.end - a.self.data.start)
        )
    ]),
    genuine_length: FunctionObject('genuine_length', [
        CreateInstance (
            'local String::genuine_length (String self) -> Size',
            a => a.self.genuineLength()
        )
    ]),
    slice: FunctionObject('slice', [
        CreateInstance (
            'local HasSlice::slice (HasSlice self, Int start, Int end) -> Slice',
            a => SliceObject(a.self, a.start, a.end)
        )
    ]),
    append: FunctionObject('append', [
        CreateInstance (
            'local MutList::append (MutList &self, Immutable element) -> Void',
            a => ListOperations.append
        ),
        CreateInstance (
            'local MutList::append (MutList &self, Mutable &element) -> Void',
            a => ListOperations.append
        ),
    ])
})


const HashOperations = {
    get: a => a.self.data.has(a.key)
                && a.self.data[a.key]
                || ErrorProducer(KeyError, 'Hash::get_data').throw(`${a.key}`),
    find: a => a.self.data.has(a.key) && a.self.data[a.key] || K['N/A'],
    set: a => (a.self.data[a.key] = a.value, K.Void)
}


pour(K, {
    has: FunctionObject('has', [
        CreateInstance (
            'Hash::has (Hash self, String key) -> Bool',
            a => a.self.data.has(a.key)
        )
    ]),
    get: FunctionObject('get', [
        CreateInstance (
            'ImHash::get (ImHash self, String key) -> Immutable',
            HashOperations.get
        ),
        CreateInstance (
            'MutHash::get (MutHash &self, String key) -> Object',
            HashOperations.get
        )
    ]),
    find: FunctionObject('find', [
        CreateInstance (
            'ImHash::find (ImHash self, String key) -> Immutable',
            HashOperations.find
        ),
        CreateInstance (
            'MutHash::find (MutHash &self, String key) -> Object',
            HashOperations.find
        )
    ]),
    set: FunctionObject('set', [
        CreateInstance (
            'MutHash::set (MutHash &self, String key, Immutable value) -> Void',
            HashOperations.set
        ),
        CreateInstance (
            'MutHash::set (MutHash &self, String key, Mutable &value) -> Void',
            HashOperations.set
        )
    ])
})
