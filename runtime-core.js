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
    dirty: x => assert(x.is(MutableObject)) && x,
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
    $(x => x.is(FunctionObject)),
    $(x => x.is(OverloadObject))
)


const AtomicObject = $u(
    PrimitiveObject,
    FunctionalObject,
    $(x => x.is(ConceptObject))
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
    assert(list.is(Array))
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
        Function: $(x => FunctionObject.contains(x)),
        Concept: $(x => ConceptObject.contains(x)),
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


function ForceMutable (object) {
    return transform(object, [
        {
            when_it_is: ImmutableObject,
            use: x => (x.maker)(
                x.data,
                Config.mutable_from(x)
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
 *  Scope Definition
 *
 *  Since the constructor of singleton requires ConceptChecker,
 *  and ConceptChecker requires definition of global scope,
 *  we have to create a null pointer of NullScope here.
 */


const NullScope = {}  // SingletonObject('NullScope')


SetEquivalent(NullScope, $1(NullScope))


function Scope (context, range, data = {}) {
    assert(context.is(Scope))
    assert(range.is(EffectRange))
    assert(Hash.contains(data))
    return {
        data: data,
        context: context,
        range: range,  // effect range of function
        maker: Scope,
        __proto__: once(Scope, Detail.Scope.Prototype)
    }
}


SetEquivalent(Scope, $u(NullScope, MadeBy(Scope)) )


/**
 *  Global Object Definition
 */


const G = Scope(NullScope, 'global')
const K = G.data
K.scope = HashObject(K)


/**
 *  Concept & Function Definition
 *  
 *  We have to create null pointers of AnyConcept and BoolConcept previously.
 *  It's because a Concept is defined by its checker function,
 *  so we have to build a Function Instance before building a concept,
 *  and it is necessary to build a Function Prototype before building
 *  this Function Instance, the prototype can be described as
 *  f: (object::Any) -> Bool, which requires the references of Any and Bool.
 *  In addition, when we invoke the checker function of a concept, we
 *  should check whether the return value is a bool value. If the function
 *  were Bool::Checker, infinite recursion would happen. So it is necessary
 *  to disable return value check when the function is Bool::Checker.
 */


const AnyConcept = {}  // ConceptObject('Any', x => x)
const BoolConcept = {}  // ConceptObject('Bool', x => x.is(BoolObject))


function ConceptObject (name, f) {
    check(ConceptObject, arguments, {
        name: Str, f: $u(Function, FunctionalObject)
    })
    let checker = (f.is(Function))? ConceptChecker(`${name}`, f): f
    return {
        name: name,
        checker: checker,
        maker: ConceptObject,
        __proto__: once(ConceptObject, {
            toString: function () {
                return `Concept<'${this.name}'>`
            }
        })
    }
}


SetEquivalent(
    ConceptObject,
    $u( $f(AnyConcept, BoolConcept), MadeBy(ConceptObject) )
)


pour(ConceptObject, Detail.Concept)


/* Function Prototype Definition */


const Parameter = Struct({
    name: Str,
    constraint: ConceptObject,
    pass_policy: PassPolicy
})


const Prototype = Struct({
    effect_range: EffectRange,
    parameters: ArrayOf(Parameter),
    value_constraint: ConceptObject
})



pour(Prototype, Detail.Prototype)


/* Function Definition */


function FunctionObject (name, context, prototype, js_function) {
    check(FunctionObject, arguments, {
        name: Str,
        context: Scope,
        prototype: Prototype,
        js_function: Function
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
                assert(args.is(ArrayOf(ObjectObject)))
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
                            is_immutable: v.is(ImmutableObject)
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
                /* special process for BookConcept::Checker */
                if (this !== BoolConcept.checker) {
                    err_r.if_failed(Proto.check_return_value(proto, value))
                }
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


/* Concept Checker Definition */


const ConceptCheckerPrototype = {
    effect_range: 'local',
    parameters: [{
        name: 'object',
        constraint: AnyConcept,
        pass_policy: 'immutable'
    }],
    value_constraint: BoolConcept
}


function ConceptChecker (name, f) {
    check(ConceptChecker, arguments, { name: Str, f: Function })
    return FunctionObject(
        `${name}::Checker`, G, ConceptCheckerPrototype,
        scope => f(
            scope.data.object,
            scope.data.argument_info.data.object.data
        )
    )
}


SetEquivalent(
    ConceptChecker,
    $n(FunctionObject, $(f => f.prototype.is(
        $n(
            Struct({
                effect_range: $1('local'),
                order: $(array => array.length == 1),
                return_value: $1(BoolConcept)
            }),
            $(proto => proto.parameters[proto.order[0]].is(Struct({
                constraint: $1(AnyConcept),
                pass_policy: $1('immutable')
            })))
        )
    )))
)


/* Singleton Definition */


const SingletonOfName = {}


function SingletonObject (name) {
    let err = ErrorProducer(NameConflict, 'Singleton::Creator')
    err.if(
        Boolean(SingletonOfName[name]),
        `singleton name ${name} already in use`
    )
    let singleton = {}
    pour(singleton, ConceptObject(
        `Singleton<'${name}'>`,
        object => object === singleton
    ))
    pour(singleton, {
        contains: x => x === singleton,
        singleton_name: name,
        __proto__: once(SingletonObject, {
            toString: function () {
                return `Singleton<'${this.singleton_name}'>`
            }
        })
    })
    SingletonOfName[name] = singleton
    return singleton
}


SetEquivalent(SingletonObject, $n(
    ConceptObject,
    $(x => typeof x['singleton_name'] == 'string'),
    $(x => x === SingletonOfName[x.singleton_name])
))


/* Fix NullScope */


pour(NullScope, SingletonObject('NullScope'))


/* Default Singleton Objects */


const VoidObject = SingletonObject('Void')
const NaObject = SingletonObject('N/A')


pour(K, {
    NullScope: NullScope,
    Void: VoidObject,
    'N/A': NaObject,
})


/**
 *  Port Native Concepts
 */


function PortEquivalent(object, concept, name) {
    check(
        PortEquivalent, arguments,
        { object: Object, concept: Concept, name: Str }
    )
    pour(object, ConceptObject(name, x => x.is(concept)))
}


PortEquivalent(AnyConcept, ObjectObject, 'Any')
PortEquivalent(BoolConcept, BoolObject, 'Bool')


const ImmutableConcept = ConceptObject(
    'Immutable', (_, info) => info.is_immutable
)
const MutableConcept = ConceptObject(
    'Mutable', (_, info) => !info.is_immutable
)

function PortConcept(concept, name) {
    check(PortConcept, arguments, {
        concept: Concept, name: Str
    })
    return ConceptObject(name, x => x.is(concept))
}


pour(PortConcept, {
    Immutable: (c, n, N) => ConceptObject.Intersect(
        ImmutableConcept, PortConcept(c, N), n
    ),
    Mutable: (c, n, N) => ConceptObject.Intersect(
        MutableConcept, PortConcept(c, N), n
    )
})


pour(K, {
    /* concept */
    Concept: PortConcept(ConceptObject, 'Concept'),
    /* special */
    Any: AnyConcept,
    Bool: BoolConcept,
    /* atomic */
    Atomic: PortConcept(AtomicObject, 'Atomic'),
    /* primitive */
    Number: PortConcept(NumberObject, 'Number'),
    Int: PortConcept(Int, 'Int'),
    UnsignedInt: PortConcept(UnsignedInt, 'UnsignedInt'),
    Finite: ConceptObject('Finite', x => x.is(Num) && Number.isFinite(x)),
    NaN: ConceptObject('NaN', x => x.is(Num) && Number.isNaN(x)),
    String: PortConcept(StringObject, 'String'),
    Primitive: PortConcept(PrimitiveObject, 'Primitive'),
    /* non-primitive atomic */
    Function: PortConcept(FunctionObject, 'Function'),
    Overload: PortConcept(OverloadObject, 'Overload'),
    Singleton: PortConcept(SingletonObject, 'Singleton'),
    /* compound */
    Compound: PortConcept(CompoundObject, 'Compound'),
    List: PortConcept(ListObject, 'List'),
    Hash: PortConcept(HashObject, 'Hash'),
    /* mutability, frozen and solid */
    ImHash: PortConcept.Immutable(HashObject, 'ImHash', 'Hash'),
    MutHash: PortConcept.Mutable(HashObject, 'MutHash', 'Hash'),
    ImList: PortConcept.Immutable(ListObject, 'ImList', 'List'),
    MutList: PortConcept.Mutable(ListObject, 'MutList', 'List'),
    Immutable: ImmutableConcept,
    Mutable: MutableConcept
})


/* Concept Alias */


pour(K, {
    Object: K.Any,
    Index: K.UnsignedInt,
    Size: K.UnsignedInt
})


/**
 *  Function Chain Definition
 */


function OverloadObject (name, instances) {
    check(OverloadObject, arguments, {
        name: Str,
        instances: $n(
            ArrayOf(FunctionObject),
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
                assert(instance.is(FunctionObject))
                let new_list = map(this.instances, x => x)
                new_list.push(instance)
                return OverloadObject(this.name, new_list)
            },            
            apply: function (...args) {
                assert(args.is(ArrayOf(ObjectObject)))
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
                let match = this.instances.transform_by(chain(
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
                    I => I.is(FunctionObject.MethodFor(object))
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
 *  Fundamental Functions Definition
 */

pour(K, {
    singleton: FunctionObject.create(
        'global singleton (String name) -> Singleton',
        a => SingletonObject(a.name)
    )
})


pour(K, {
    
    is: OverloadObject('is', FunctionObject.converge([
        'local Immutable::is (Immutable object, Concept concept) -> Bool',
        'local Mutable::is (Mutable &object, Concept concept) -> Bool',    
    ], a => a.concept.checker.apply(a.object) )),
    
    union: OverloadObject('union', [ FunctionObject.create(
        'local union (Concept c1, Concept c2) -> Concept',
        a => ConceptObject.Union(a.c1, a.c2)            
    )]),
     
    intersect: OverloadObject('intersect', [ FunctionObject.create(
        'local intersect (Concept c1, Concept c2) -> Concept',
        a => ConceptObject.Intersect(a.c1, a.c2)
    )]),
     
    complement: OverloadObject('complement', [ FunctionObject.create(
        'local complement (Concept c) -> Concept',
        a => ConceptObject.Complement(a.c)
    )])
    
})


pour(K, {
    
    type_of: OverloadObject('type_of', [
        FunctionObject.create (
            'local type_of (Any object) -> String',
            a => GetType(a.object)
        )
    ]),

    Im: OverloadObject('Im', FunctionObject.converge([
        'local Im (Immutable object) -> Immutable',
        'local Im (Mutable &object) -> Immutable'
    ], a => ImRef(a.object) )),
     
    Id: OverloadObject('Id', FunctionObject.converge([
        'local Id (Immutable object) -> Immutable',
        'local Id (Mutable &object) -> Mutable'
    ], a => a.object )),

    copy: OverloadObject('copy', [
        FunctionObject.create (
            'local copy (Atomic object) -> Atomic',
            a => a.object
        ),
        FunctionObject.create (
            'local copy (ImList object) -> MutList',
            a => ListObject(map(a.object.data, e => ImRef(e)))
        ),
        FunctionObject.create (
            'local copy (MutList &object) -> MutList',
            a => ListObject(map(a.object.data, e => e))
        ),
        FunctionObject.create (
            'local copy (ImHash object) -> MutHash',
            a => HashObject(mapval(a.object.data, v => ImRef(v)))
        ),
        FunctionObject.create (
            'local copy (MutHash &object) -> MutHash',
            a => HashObject(mapval(a.object.data, v => v))
        ),
    ]),
    
    clone: OverloadObject('clone', [
        FunctionObject.create(
            'local clone (Atomic object) -> Atomic',
            a => a.object
        ),
        FunctionObject.create(
            'local clone (Compound object) -> Mutable',
            a => Clone(a.object)
        )
    ])
    
})


pour(K, {
    
    at: OverloadObject('at', [
        FunctionObject.create (
            'local ImList::at (ImList self, Index index) -> Immutable',
            a => ImRef(a.self.at(a.index))
        ),
        FunctionObject.create (
            'local MutList::at (MutList &self, Index index) -> Object',
            a => a.self.at(a.index)
        )
    ]),
     
    append: OverloadObject('append', FunctionObject.converge([
        'local MutList::append (MutList &self, Immutable element) -> Void',
        'local MutList::append (MutList &self, Mutable &element) -> Void'
    ], a => a.self.append(a.element) )),
    
    length: OverloadObject('length', [ FunctionObject.create (
        'local List::length (List self) -> Size',
        a => a.self.length()
    )])

})


pour(K, {
    
    has: OverloadObject('has', [
        FunctionObject.create (
            'local Hash::has (Hash self, String key) -> Bool',
            a => a.self.has(a.key)
        )
    ]),
    
    get: OverloadObject('get', [
        FunctionObject.create (
            'local ImHash::get (ImHash self, String key) -> Immutable',
            a => ImRef(a.self.get(a.key))
        ),
        FunctionObject.create (
            'local MutHash::get (MutHash &self, String key) -> Object',
            a => a.self.get(a.key)
        ),
        /* List also have a get function, which calls at() */
        FunctionObject.create (
            'local ImList::get (ImList self, Index index) -> Immutable',
            a => ImRef(a.self.at(a.index))
        ),
        FunctionObject.create (
            'local MutList::get (MutList &self, Index index) -> Object',
            a => a.self.at(a.index)
        )
    ]),
    
    fetch: OverloadObject('fetch', [
        FunctionObject.create (
            'local ImHash::fetch (ImHash self, String key) -> Immutable',
            a => ImRef(a.self.fetch(a.key))
        ),
        FunctionObject.create (
            'local MutHash::fetch (MutHash &self, String key) -> Object',
            a => a.self.fetch(a.key)
        )
    ]),
    
    set: OverloadObject('set', FunctionObject.converge([
        'local MutHash::set (MutHash &self, String key, Immutable value) -> Void',
        'local MutHash::set (MutHash &self, String key, Mutable &value) -> Void'
    ], a => a.self.set(a.key, a.value) ))
    
})


pour(K, {
    
    plus: OverloadObject('plus', [
        FunctionObject.create (
            'local plus (Number p, Number q) -> Number',
            a => a.p + a.q
        ),
        FunctionObject.create (
            'local plus (String s1, String s2) -> String',
            a => a.s1 + a.s2
        )
    ]),
    
    minus: OverloadObject('minus', [
        FunctionObject.create (
            'local minus (Number p, Number q) -> Number',
            a => a.p - a.q
        ),
        FunctionObject.create (
            'local minus (Number x) -> Number',
            a => -a.x
        ),
        FunctionObject.create(
            'local minus (Concept c1, Concept c2) -> Concept',
            a => ConceptObject.Intersect(
                a.c1, ConceptObject.Complement(a.c2)
            )
        )
    ]),
    
    multiply: OverloadObject('multiply', [
        FunctionObject.create (
            'local multiply (Number p, Number q) -> Number',
            a => a.p * a.q
        )
    ]),
    
    divide: OverloadObject('divide', [
        FunctionObject.create (
            'local divide (Number p, Number q) -> Number',
            a => a.p / a.q
        )
    ]),
    
    mod: OverloadObject('mod', [
        FunctionObject.create (
            'local mod (Number p, Number q) -> Number',
            a => a.p % a.q
        )
    ]),
    
    pow: OverloadObject('pow', [
        FunctionObject.create (
            'local pow (Number p, Number q) -> Number',
            a => Math.pow(a.p, a.q)
        )
    ]),
    
    floor: OverloadObject('floor', [
        FunctionObject.create (
            'local floor (Number x) -> Number',
            a => Math.floor(a.x)
        )
    ]),
    
    ceil: OverloadObject('ceil', [
        FunctionObject.create (
            'local ceil (Number x) -> Number',
            a => Math.ceil(a.x)
        )
    ]),
    
    round: OverloadObject('round', [
        FunctionObject.create (
            'local round (Number x) -> Number',
            a => Math.round(a.x)
        )
    ])
    
})
