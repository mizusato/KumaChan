'use strict';


/**
 *  Pass Policy Definition
 * 
 *  - dirty: pass argument as a mutable reference
 *           (the argument must be mutable)
 *  - immutable: pass argument as an immutable reference
 *               (the argument can be mutable or immutable)
 */


const PassPolicy = Enum('dirty', 'immutable')
const PassAction = {
    dirty: x => assert(is(x, MutableObject)) && x,
    immutable: x => ImRef(x)
}
const PassFlag = {
    dirty: '&',
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
 *               ┴ Atomic ┬ Concept
 *                        ┼ Iterator
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


/* Object (Any) Definition */


const ObjectObject = $u(CompoundObject, AtomicObject)


pour(ObjectObject, Detail.Object)


/* GetType Function Definition */


function GetType(object) {
    let non_primitive = {
        Concept: $(x => ConceptObject.contains(x)),
        Iterator: $(x => IteratorObject.contains(x)),
        Function: $(x => FunctionObject.contains(x)),
        Overload: $(x => OverloadObject.contains(x)),
        List: ListObject,
        Hash: HashObject
    }
    let checker = {
        'string': () => 'String',
        'number': () => 'Number',
        'boolean': () => 'Bool',
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


const MutableObject = $n(CompoundObject, $(x => !x.config.immutable))
const ImmutableObject = $_(MutableObject)


/**
 *  Immutable/Mutable Reference & Clone
 */


function ImRef (object) {
    return transform(object, [
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
            when_it_is: AtomicObject,
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
const FolderObject = ParameterCount(2)


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
    true: true,
    false: false,
    NullScope: NullScope,
    Void: VoidObject,
    'N/A': NaObject,
    Done: DoneObject
})


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
                return find(
                    this.instances,
                    I => is(I, FunctionObject.MethodFor(object))
                )
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


/* Any should be ObjectObject, assert when calling FunctionObject */
const AnyConcept = PortConcept(Any, 'Any')


pour(K, {
    /* any */
    Any: AnyConcept,
    /* atomic */
    Atomic: PortConcept(AtomicObject, 'Atomic'),
    /* concept */
    Concept: PortConcept(ConceptObject, 'Concept'),
    Iterator: PortConcept(IteratorObject, 'Iterator'),
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
    /* compound */
    Compound: PortConcept(CompoundObject, 'Compound'),
    List: PortConcept(ListObject, 'List'),
    Hash: PortConcept(HashObject, 'Hash'),
    /* mutability, frozen and solid */
    ImHash: PortConcept(ImHashObject, 'ImHash'),
    MutHash: PortConcept(MutHashObject, 'MutHash'),
    ImList: PortConcept(ImListObject, 'ImList'),
    MutList: PortConcept(MutListObject, 'MutList'),
    Immutable: PortConcept(ImmutableObject, 'Immutable'),
    Mutable: PortConcept(MutableObject, 'Mutable')
})


/* Concept Alias */


pour(K, {
    Object: K.Any,
    Index: K.UnsignedInt,
    Size: K.UnsignedInt
})
