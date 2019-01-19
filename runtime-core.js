'use strict';


/**
 *  Pass Policy Definition
 * 
 *  - dirty: pass argument as a mutable reference
 *           (the argument must be mutable)
 *  - immutable: pass argument as an immutable reference
 *               (the argument can be mutable or immutable, but not raw)
 *  - natural: pass argument as is
 */


const PassPolicy = Enum('dirty', 'immutable', 'natural')
const PassAction = {
    dirty: x => assert(is(x, MutableObject)) && x,
    natural: x => x,
    immutable: x => ImRef(x)
}
const PassFlag = {
    dirty: '&',
    natural: '*',
    immutable: ''
}
const PassFlagValue = fold(
    Object.keys(PassFlag), {},
    (key, v) => (v[PassFlag[key]] = key, v)
)


/**
 *  Effect Range Definition
 * 
 *  The effect range of a function determines the range of scope chain
 *  that can be affected by the function, which indicates
 *  the magnitude of side-effect.
 *  
 *  |  value  | Local Scope | Upper Scope | Other Scope |
 *  |---------|-------------|-------------|-------------|
 *  |  global | full-access | full-access | full-access |
 *  |  upper  | full-access | full-access |  read-only  |
 *  |  local  | full-access |  read-only  |  read-only  |
 * 
 */


const EffectRange = Enum('global', 'upper', 'local')


/**
 *  Object Type Definition
 * 
 *  Object (Any) ┬ Compound ┬ Hash
 *               │          ┴ List
 *               ┼ Atomic ┬ Concept
 *               ┴ Raw    ┼ Iterator
 *                        ┼ Primitive ┬ String
 *                        │           ┼ Number
 *                        │           ┴ Bool
 *                        ┴ Functional ┬ Function
 *                                     ┴ Overload
 *
 *  Note: atomic object must be immutable.
 */


/* Primitive Definition */


const StringObject = $(x => typeof x == 'string')
const NumberObject = $(x => typeof x == 'number')
const BoolObject = $(x => typeof x == 'boolean')
const PrimitiveObject = $u(StringObject, NumberObject, BoolObject)


/* Atomic Definition */


const FunctionalObject = $u(
    $(x => is(x, FunctionObject)),
    $(x => is(x, OverloadObject))
)


const AtomicObject = $u(
    PrimitiveObject,
    FunctionalObject,
    $(x => is(x, ConceptObject)),
    $(x => is(x, IteratorObject))
)


/* Hash Definition */


const Config = {
    default: {
        immutable: false
    },
    from: object => pour({}, object.config),
    immutable_from: object => pour(Config.from(object), {
        immutable: true
    }),
    mutable_from: object => pour(Config.from(object), {
        immutable: false
    }),
    get_flags: Detail.Config.get_flags
}


function HashObject (hash = {}, config = Config.default) {
    assert(Hash.contains(hash))
    assert(Hash.contains(config))
    return {
        data: hash,
        config: config,
        maker: HashObject,
        __proto__: once(HashObject, Detail.Hash.Prototype)
    }
}


SetMakerConcept(HashObject)


const ImHashObject = $n(HashObject, $(x => x.config.immutable))
const MutHashObject = $n(HashObject, $(x => !x.config.immutable))


/* List Definition */


function ListObject (list = [], config = Config.default) {
    assert(is(list, List))
    assert(Hash.contains(config))
    return {
        data: list,
        config: config,
        maker: ListObject,
        __proto__: once(ListObject, Detail.List.Prototype)
    }
}


SetMakerConcept(ListObject)


const ImListObject = $n(ListObject, $(x => x.config.immutable))
const MutListObject = $n(ListObject, $(x => !x.config.immutable))


/* Compound Definition */


const CompoundObject = $u(HashObject, ListObject)
const ImCompoundObject = $u(ImHashObject, ImListObject)
const MutCompoundObject = $u(MutHashObject, MutListObject)


/* Raw Definition */


const CookedObject = $u(CompoundObject, AtomicObject)
const RawObject = $_(CookedObject)
const RawHashObject = $n(RawObject, StrictHash)
const RawListObject = $n(RawObject, List)
const RawCompoundObject = $u(RawListObject, RawHashObject)
const RawFunctionObject = $n(RawObject, Fun)
const NullObject = $1(null)
const UndefinedObject = $1(undefined)
const SymbolObject = $(x => typeof x == 'symbol')
const CompatibleObject = $u(PrimitiveObject, RawObject)
const RawableObject = $u(MutCompoundObject, FunctionalObject, CompatibleObject)


function RawCompound (compound) {
    check(RawCompound, arguments, { compound: CompoundObject })
    let err = ErrorProducer(InvalidOperation, 'RawCompound')
    let msg = 'unable to raw immutable compound object'
    err.assert(is(compound, MutCompoundObject), msg)
    function raw_element (e) {
        return transform(e, [
            { when_it_is: CompoundObject, use: c => RawCompound(c) },
            { when_it_is: RawableObject, use: r => r },
            { when_it_is: Otherwise, use: o => err.throw('rawing non-rawable') }
        ])
    }
    return transform(compound, [
        { when_it_is: HashObject, use: h => mapval(h.data, raw_element) },
        { when_it_is: ListObject, use: l => map(l.data, raw_element) }
    ])
}


function CookCompound (raw_compound) {
    check(CookCompound, arguments, { raw_compound: RawCompoundObject })
    function cook_element (e) {
        return transform(e, [
            { when_it_is: RawCompoundObject, use: c => CookCompound(c) },
            { when_it_is: Otherwise, use: o => o }
        ])
    }
    return transform(raw_compound, [
        { when_it_is: StrictHash,
          use: h => HashObject(mapval(h, cook_element)) },
        { when_it_is: List,
          use: l => ListObject(map(l, cook_element)) }
    ])
}


/* Object (Any) Definition */


const ObjectObject = $u(CompoundObject, AtomicObject, RawObject)


pour(ObjectObject, Detail.Object)


/* GetType Function Definition */


function GetType(object) {
    let non_primitive = {
        Concept: $(x => ConceptObject.contains(x)),
        Iterator: $(x => IteratorObject.contains(x)),
        Function: $(x => FunctionObject.contains(x)),
        Overload: $(x => OverloadObject.contains(x)),
        List: ListObject,
        Hash: HashObject,
        Raw: RawObject
    }
    let checker = {
        'string': () => 'String',
        'number': () => 'Number',
        'boolean': () => 'Bool',
        'undefined': () => 'Raw',
        'function': () => 'Raw',
        'symbol': () => 'Raw',
        'object': function (object) {
            let r = find(
                non_primitive,
                concept => concept.contains(object)
            )
            assert(r != NotFound)
            return r.key
        }
    }
    return checker[typeof object](object)
}


/**
 *  Mutable and Immutable Definition
 */


const MutableObject = $u(
    $n(CompoundObject, $(x => !x.config.immutable)),
    RawObject
)
const ImmutableObject = $_(MutableObject)


/**
 *  Immutable/Mutable Reference & Clone
 */


function ImRef (object) {
    return transform(object, [
        {
            when_it_is: RawObject,
            use: x => ErrorProducer(InvalidOperation, 'ImRef').throw(
                'unable to make immutable reference for raw object'
            )
        },
        {
            when_it_is: MutableObject,
            use: x => (x.maker)(
                x.data,
                Config.immutable_from(x)
            )
        },
        {
            when_it_is: Otherwise,
            use: x => x
        }
    ])
}


function Clone (object) {
    return transform(object, [
        {
            when_it_is: CompoundObject,
            use: x => (x.maker)(
                (x.mapper)(x.data, v => Clone(v)),
                Config.mutable_from(x)
            )
        },
        {
            when_it_is: $u(AtomicObject, RawObject),
            use: x => x
        }
    ])
}


/**
 *  Concept Definition
 */


function ConceptObject (name, f) {
    check(ConceptObject, arguments, {
        name: Str, f: $u(Fun, FunctionalObject)
    })
    let raw_checker = (is(f, Fun))? f: function (object) {
        return f.apply(object)
    }
    let wrapped_checker = function (object) {
        let result = raw_checker(object)
        assert(is(result, BoolObject))
        return result
    }
    return {
        name: name,
        checker: wrapped_checker,
        maker: ConceptObject,
        __proto__: once(ConceptObject, {
            toString: function () {
                return `Concept<'${this.name}'>`
            }
        })
    }
}


SetMakerConcept(ConceptObject)


pour(ConceptObject, Detail.Concept)


const ConceptListObject = $n(
    ListObject, $(x => is(x.data, ListOf(ConceptObject)))
)


/* Singleton Definition */


function SingletonObject (name) {
    let singleton = {}
    pour(singleton, ConceptObject(
        `Singleton('${name}')`,
        object => object === singleton
    ))
    return singleton
}


/**
 *  Iterator Definition
 */


const ParameterCount = (n => $( x => (
    is(x, FunctionObject)
    && x.prototype.parameters.length == n
)))


const IteratorFunctionObject = ParameterCount(0)
const MapperObject = ParameterCount(1)
const FilterObject = ParameterCount(1)
const ReducerObject = ParameterCount(2)


function IteratorObject (f) {
    check(IteratorObject, arguments, {
        f: $u(Fun, IteratorFunctionObject)
    })
    let next = (is(f, Fun))? f: function () {
        return f.apply()
    }
    let wrapped_next = (function () {
        let done = false
        return function () {
            let value = next()
            assert(!(done && value != DoneObject))
            done = (value == DoneObject)? true: false
            return value
        }
    })()
    return {
        next: wrapped_next,
        maker: IteratorObject,
        __proto__: once(IteratorObject, {
            toString: function () {
                return `<Iterator>`
            }
        })
    }
}


SetMakerConcept(IteratorObject)


const IterableObject = $u(ListObject, IteratorObject)
const ListOfIterableObject = $n(
    ListObject, $(x => is(x.data, ListOf(IterableObject)))
)
const HashOfIterableObject = $n(
    HashObject, $(x => is(x.data, HashOf(IterableObject)))
)


/**
 *  Scope Definition
 */


const NullScope = { contains: SingletonContains, _name: 'NullScope' }


function Scope (context, range, data = {}) {
    assert(is(context, Scope))
    assert(is(range, EffectRange))
    assert(Hash.contains(data))
    return {
        data: data,
        context: context,
        range: range,  // effect range of function
        maker: Scope,
        __proto__: once(Scope, Detail.Scope.Prototype)
    }
}


SetEquivalent(Scope, $u(MadeBy(Scope), NullScope) )


/**
 *  Global Object Definition
 */


const G = Scope(NullScope, 'global')
const K = G.data
K.scope = HashObject(G.data)
const scope = G
const id = (name => G.lookup(name))


const VoidObject = SingletonObject('Void')
const NaObject = SingletonObject('N/A')
const DoneObject = SingletonObject('Done')


pour(K, {
    Void: VoidObject,
    'N/A': NaObject,
    Done: DoneObject,
    true: true,
    false: false,
    null: null,
    undefined: undefined,
    CR: CR,
    LF: LF,
    TAB: TAB
})


/**
 *  Pair & Case Object Definition
 */


const PairObject = $n(HashObject, $(x => is(x.data, Struct({
    key: Any, value: Any
}))))


const PairListObject = $n(
    ListObject, $(x => forall(x.data, e => is(e, PairObject)))
)


const CaseObject = $n(HashObject, $(x => is(x.data, Struct({
    key: $u(ConceptObject, FilterObject, BoolObject), value: Any
}))))


const CaseListObject = $n(
    ListObject, $(x => forall(x.data, e => is(e, CaseObject)))
)


/**
 *  Function Definition
 */


const Parameter = Struct({
    name: Str,
    constraint: ConceptObject,
    pass_policy: PassPolicy
})


const Prototype = Struct({
    effect_range: EffectRange,
    parameters: ListOf(Parameter),
    value_constraint: ConceptObject
})


pour(Prototype, Detail.Prototype)


function FunctionObject (name, context, prototype, js_function) {
    check(FunctionObject, arguments, {
        name: Str,
        context: Scope,
        prototype: Prototype,
        js_function: Fun
    })
    fold(prototype.parameters, {}, function (parameter, appeared) {
        let err = ErrorProducer(RedundantParameter, 'Function::create()')
        let name = parameter.name
        err.if(
            appeared[name] !== undefined,
            `parameter ${name} defined more than once`
        )
        appeared[name] = true
        return appeared
    })
    return {
        name: name || '[Anonymous]',
        context: context,
        prototype: prototype,
        js_function: js_function,
        maker: FunctionObject,
        __proto__: once(FunctionObject, {
            apply: function (...args) {
                assert(is(args, ListOf(ObjectObject)))
                return this.call(fold(args, {}, (e, v, i) => (v[i] = e, v)) )
            },
            call: function (argument, caller) {
                assert(HashOf(ObjectObject).contains(argument))
                /* define error producers */
                let name = this.name
                let { err_a, err_r } = mapval({
                    err_a: InvalidArgument,
                    err_r: InvalidReturnValue,
                }, ErrorType => ErrorProducer(ErrorType, `${name}`))
                /* shortcuts */
                let Proto = Prototype
                let Range = EffectRange
                let { proto, range, f, context } = {
                    proto:   this.prototype,
                    range:   this.prototype.effect_range,
                    f:       this.js_function,
                    context: this.context
                }
                /* check if argument valid */
                err_a.if_failed(Proto.check_argument(proto, argument))
                /* create new scope */
                let normalized = Proto.normalize_argument(proto, argument)
                let final = Proto.set_mutability(proto, normalized)
                let scope = Scope(context, range, {
                    callee: this,
                    argument: HashObject(final),
                    argument_info: HashObject(
                        mapval(normalized, v => HashObject({
                            is_immutable: is(v, ImmutableObject)
                        }))
                    )
                })
                scope.data['scope'] = HashObject(scope.data)
                /**
                 *  it is not good to bind the name of function to the scope
                 *  because the function name in the scope could override
                 *  the previous overridden Overload, which is dumped in
                 *  the outer wrapper scope created by define() (runtime-tools)
                 */
                // scope.data[name] = this
                pour(scope.data, final)
                /* invoke js function */
                let value = f(scope)
                /* debug output */
                let arg_str = Detail.Argument.represent(normalized)
                let val_str = ObjectObject.represent(value)
                console.log(`${name}${arg_str} = ${val_str}`)
                return value
            },
            toString: function () {
                let proto_repr = Prototype.represent(this.prototype)
                let split = proto_repr.split(' ')
                let effect = split.shift()
                let rest = split.join(' ')
                return `${effect} ${this.name} ${rest}`
            }
        })
    }
}


SetMakerConcept(FunctionObject)


pour(FunctionObject, Detail.Function)


/**
 *  Function Overload Definition
 */


function OverloadObject (name, instances) {
    check(OverloadObject, arguments, {
        name: Str,
        instances: $n(
            ListOf(FunctionObject),
            $(array => array.length > 0)
        ),
        equivalent_concept: Optional(ConceptObject)
    })
    return {
        name: name,
        instances: instances,
        maker: OverloadObject,
        __proto__: once(OverloadObject, {
            added: function (instance) {
                assert(is(instance, FunctionObject))
                let new_list = map(this.instances, x => x)
                new_list.push(instance)
                return OverloadObject(this.name, new_list)
            },            
            apply: function (...args) {
                assert(is(args, ListOf(ObjectObject)))
                return this.call(fold(args, {}, (e, v, i) => (v[i] = e, v)) )
            },
            call: function (argument) {
                assert(HashOf(ObjectObject).contains(argument))
                let err = ErrorProducer(NoMatchingPattern, `${this.name}`)
                let check_arg = (
                    proto => Prototype.check_argument(
                        proto, argument
                    )
                )
                let match = apply_on(this.instances, chain(
                    x => rev(x),
                    x => map_lazy(x, f => ({
                        instance: f,
                        is_ok: (check_arg(f.prototype) == OK)
                    })),
                    x => find(x, attempt => attempt.is_ok)
                ))
                err.if(match == NotFound, join_lines(
                    `invalid call: match not found`,
                    `available instances are:`, `${this}`
                ))
                return match.instance.call(argument)
            },
            find_method_for: function (object) {
                let name = `<${GetType(object)}>.${this.name}`
                let found = filter(
                    this.instances,
                    I => is(I, FunctionObject.MethodFor(object))
                )
                let methods = map(found, function (method) {
                    // create wrappers
                    let p = method.prototype.parameters
                    let shifted = p.slice(1, p.length)
                    let proto = pour(pour({}, method.prototype), {
                        parameters: shifted
                    })
                    let first_parameter = p[0].name
                    return FunctionObject(name, scope, proto, function (scope) {
                        let extended_arg = {}
                        pour(extended_arg, scope.data.argument.data)
                        extended_arg[first_parameter] = object
                        return method.call(extended_arg)
                    })
                })
                if (methods.length > 0) {
                    return OverloadObject('name', methods)
                } else {
                    return NotFound
                }
            },
            toString: function () {
                return join(map(this.instances, I => I.toString()), '\n')
            }
        })
    }
}


SetMakerConcept(OverloadObject)


/**
 *  Port Native Concepts
 */


function PortConcept(concept, name) {
    check(PortConcept, arguments, {
        concept: Concept, name: Str
    })
    return ConceptObject(name, x => is(x, concept))
}


pour(K, {
    /* any */
    Any: PortConcept(Any, 'Any'),  // Any should be ObjectObject,
                                   // assert when calling FunctionObject
    /* atomic */
    Atomic: PortConcept(AtomicObject, 'Atomic'),
    /* concept */
    Concept: PortConcept(ConceptObject, 'Concept'),
    Iterator: PortConcept(IteratorObject, 'Iterator'),
    ConceptList: PortConcept(ConceptListObject, 'ConceptList'),
    /* iterable */
    Iterable: PortConcept(IterableObject, 'Iterable'),
    IterableList: PortConcept(ListOfIterableObject, 'IterableList'),
    IterableHash: PortConcept(HashOfIterableObject, 'IterableHash'),
    /* primitive */
    Bool: PortConcept(BoolObject, 'Bool'),
    Number: PortConcept(NumberObject, 'Number'),
    Int: PortConcept(Int, 'Int'),
    UnsignedInt: PortConcept(UnsignedInt, 'UnsignedInt'),
    String: PortConcept(StringObject, 'String'),
    Primitive: PortConcept(PrimitiveObject, 'Primitive'),
    /* functional */
    Function: PortConcept(FunctionObject, 'Function'),
    Overload: PortConcept(OverloadObject, 'Overload'),
    Functional: PortConcept(FunctionalObject, 'Functional'),
    IteratorFunction: PortConcept(IteratorFunctionObject, 'IteratorFunction'),
    Mapper: PortConcept(MapperObject, 'Mapper'),
    Filter: PortConcept(FilterObject, 'Filter'),
    Reducer: PortConcept(ReducerObject, 'Reducer'),
    /* compound */
    Compound: PortConcept(CompoundObject, 'Compound'),
    List: PortConcept(ListObject, 'List'),
    Hash: PortConcept(HashObject, 'Hash'),
    /* pair and case */
    Pair: PortConcept(PairObject, 'Pair'),
    PairList: PortConcept(PairListObject, 'PairList'),
    Case: PortConcept(CaseObject, 'Case'),
    CaseList: PortConcept(CaseListObject, 'CaseList'),
    /* mutability */
    ImHash: PortConcept(ImHashObject, 'ImHash'),
    MutHash: PortConcept(MutHashObject, 'MutHash'),
    ImList: PortConcept(ImListObject, 'ImList'),
    MutList: PortConcept(MutListObject, 'MutList'),
    ImCompound: PortConcept(ImCompoundObject, 'ImCompound'),
    MutCompound: PortConcept(MutCompoundObject, 'MutCompound'),
    Immutable: PortConcept(ImmutableObject, 'Immutable'),
    Mutable: PortConcept(MutableObject, 'Mutable'),
    /* raw object */
    Cooked: PortConcept(CookedObject, 'Cooked'),
    Raw: PortConcept(RawObject, 'Raw'),
    RawHash: PortConcept(RawHashObject, 'RawHash'),
    RawList: PortConcept(RawListObject, 'RawList'),
    RawCompound: PortConcept(RawCompoundObject, 'RawCompound'),
    RawFunction: PortConcept(RawFunctionObject, 'RawFunction'),
    Null: PortConcept(NullObject, 'Null'),
    Undefined: PortConcept(UndefinedObject, 'Undefined'),
    Symbol: PortConcept(SymbolObject, 'Symbol'),
    Compatible: PortConcept(CompatibleObject, 'Compatible'),
    Rawable: PortConcept(RawableObject, 'Rawable')
})


/* Concept Alias */


pour(K, {
    Object: K.Any,
    Otherwise: K.Any,
    Index: K.UnsignedInt,
    Size: K.UnsignedInt
})
