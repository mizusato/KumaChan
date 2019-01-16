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
    $(x => is(x, ConceptObject))
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
 *  Scope Definition
 */


const NullScope = SingletonObject('NullScope')


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


SetEquivalent(Scope, $u(MadeBy(Scope), $1(NullScope)) )


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


pour(K, {
    NullScope: NullScope,
    Void: VoidObject,
    'N/A': NaObject,
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


const AnyConcept = PortConcept(Any, 'Any')


pour(K, {
    /* concept */
    Concept: PortConcept(ConceptObject, 'Concept'),
    /* special */
    Any: AnyConcept,
    Bool: PortConcept(Bool, 'Bool'),
    /* atomic */
    Atomic: PortConcept(AtomicObject, 'Atomic'),
    /* primitive */
    Number: PortConcept(NumberObject, 'Number'),
    Int: PortConcept(Int, 'Int'),
    UnsignedInt: PortConcept(UnsignedInt, 'UnsignedInt'),
    String: PortConcept(StringObject, 'String'),
    Primitive: PortConcept(PrimitiveObject, 'Primitive'),
    /* functional */
    Function: PortConcept(FunctionObject, 'Function'),
    Overload: PortConcept(OverloadObject, 'Overload'),
    Functional: PortConcept(FunctionalObject, 'Functional'),
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


/**
 *  Fundamental Functions Definition
 */


pour(K, {
    singleton: FunctionObject.create(
        'global singleton (String name) -> Concept',
        a => SingletonObject(a.name)
    )
})


pour(K, {
    
    'call': OverloadObject('call', list(cat(
        FunctionObject.converge([
            'local call (Functional f, ImHash argument_table) -> Any',
            'local call (Functional f, MutHash &argument_table) -> Any'
        ], a => a.f.call(a.argument_table.data)),
        [
            FunctionObject.create(
                'local call (Functional f) -> Functional',
                a => OverloadObject('call_by', FunctionObject.converge([
                    'local call_by (ImHash argument_table) -> Any',
                    'local call_by (MutHash &argument_table) -> Any',
                ], b => a.f.call(b.argument_table.data))
            ))
        ]
    ))),
    
    '>>': OverloadObject('>>', FunctionObject.converge([
        'local pass_to_right (Immutable object, Functional f) -> Any',
        'local pass_to_right (Mutable &object, Functional f) -> Any',    
    ], a => a.f.apply(a.object) )),
    
    '<<': OverloadObject('<<', FunctionObject.converge([
        'local pass_to_left (Functional f, Immutable object) -> Any',
        'local pass_to_left (Functional f, Mutable &object) -> Any',    
    ], a => a.f.apply(a.object) )),
     
    'operator_by': OverloadObject('operator_by', FunctionObject.converge([
        'local pass_to_left (Functional f, Immutable object) -> Any',
        'local pass_to_left (Functional f, Mutable &object) -> Any',    
    ], a => a.f.apply(a.object) ))
    
})


pour(K, {
    
    operator_is: OverloadObject('operator_is', FunctionObject.converge([
        'local is (Immutable object, Concept concept) -> Bool',
        'local is (Mutable &object, Concept concept) -> Bool',    
    ], a => a.concept.checker(a.object) )),
    
    '|': OverloadObject('|', [ FunctionObject.create(
        'local union (Concept c1, Concept c2) -> Concept',
        a => ConceptObject.Union(a.c1, a.c2)            
    )]),
     
    '&': OverloadObject('&', [ FunctionObject.create(
        'local intersect (Concept c1, Concept c2) -> Concept',
        a => ConceptObject.Intersect(a.c1, a.c2)
    )]),
     
    '~': OverloadObject('~', [ FunctionObject.create(
        'local complement (Concept c) -> Concept',
        a => ConceptObject.Complement(a.c)
    )]),
    
    '||': OverloadObject('||', [ FunctionObject.create(
        'local or (Bool v1, Bool v2) -> Bool',
         a => a.v1 || a.v2
    )]),
    
    '&&': OverloadObject('&&', [ FunctionObject.create(
        'local and (Bool v1, Bool v2) -> Bool',
        a => a.v1 && a.v2
    )]),
     
    '!': OverloadObject('!', [ FunctionObject.create(
        'local not (Bool v) -> Concept',
        a => !a.v
    )])
     
})


pour(K, {
    
    type: OverloadObject('type', [
        FunctionObject.create (
            'local type (Any object) -> String',
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
            'local at (ImList self, Index index) -> Immutable',
            a => ImRef(a.self.at(a.index))
        ),
        FunctionObject.create (
            'local at (MutList &self, Index index) -> Object',
            a => a.self.at(a.index)
        )
    ]),
     
    append: OverloadObject('append', FunctionObject.converge([
        'local append (MutList &self, Immutable element) -> Void',
        'local append (MutList &self, Mutable &element) -> Void'
    ], a => a.self.append(a.element) )),
    
    length: OverloadObject('length', [ FunctionObject.create (
        'local length (List self) -> Size',
        a => a.self.length()
    )])

})


pour(K, {
    
    has: OverloadObject('has', [
        FunctionObject.create (
            'local has (Hash self, String key) -> Bool',
            a => a.self.has(a.key)
        )
    ]),
    
    get: OverloadObject('get', [
        FunctionObject.create (
            'local get (ImHash self, String key) -> Immutable',
            a => ImRef(a.self.get(a.key))
        ),
        FunctionObject.create (
            'local get (MutHash &self, String key) -> Object',
            a => a.self.get(a.key)
        )
    ]),
    
    fetch: OverloadObject('fetch', [
        FunctionObject.create (
            'local fetch (ImHash self, String key) -> Immutable',
            a => ImRef(a.self.fetch(a.key))
        ),
        FunctionObject.create (
            'local fetch (MutHash &self, String key) -> Object',
            a => a.self.fetch(a.key)
        )
    ]),
    
    set: OverloadObject('set', FunctionObject.converge([
        'local set (MutHash &self, String key, Immutable value) -> Void',
        'local set (MutHash &self, String key, Mutable &value) -> Void'
    ], a => a.self.set(a.key, a.value) ))
    
})


pour(K, {
    
    '+': OverloadObject('+', [
        FunctionObject.create (
            'local plus (Number p, Number q) -> Number',
            a => a.p + a.q
        ),
        FunctionObject.create (
            'local string_concat (String s1, String s2) -> String',
            a => a.s1 + a.s2
        )
    ]),
    
    '-': OverloadObject('-', [
        FunctionObject.create (
            'local minus (Number p, Number q) -> Number',
            a => a.p - a.q
        ),
        FunctionObject.create(
            'local difference (Concept c1, Concept c2) -> Concept',
            a => ConceptObject.Intersect(
                a.c1, ConceptObject.Complement(a.c2)
            )
        )
    ]),
    
    operator_negate: OverloadObject('operator_negate', [
        FunctionObject.create (
            'local negate (Number x) -> Number',
            a => -a.x
        ),
    ]),
     
    '*': OverloadObject('*', [
        FunctionObject.create (
            'local multiply (Number p, Number q) -> Number',
            a => a.p * a.q
        )
    ]),
    
    '/': OverloadObject('/', [
        FunctionObject.create (
            'local over (Number p, Number q) -> Number',
            a => a.p / a.q
        )
    ]),
    
    '%': OverloadObject('%', [
        FunctionObject.create (
            'local modulo (Number p, Number q) -> Number',
            a => a.p % a.q
        )
    ]),
    
    '^': OverloadObject('^', [
        FunctionObject.create (
            'local power (Number p, Number q) -> Number',
            a => Math.pow(a.p, a.q)
        )
    ]),
    
    '<': OverloadObject('<', [
        FunctionObject.create (
            'local less_than (Number p, Number q) -> Bool',
            a => a.p < a.q
        )
    ]),
    
    '>': OverloadObject('>', [
        FunctionObject.create (
            'local greater_than (Number p, Number q) -> Bool',
            a => a.p > a.q
        )
    ]),
    
    '>=': OverloadObject('>=', [
        FunctionObject.create (
            'local greater_than_or_equal (Number p, Number q) -> Bool',
            a => a.p >= a.q
        )
    ]),
    
    '<=': OverloadObject('<=', [
        FunctionObject.create (
            'local less_than_or_equal (Number p, Number q) -> Bool',
            a => a.p <= a.q
        )
    ]),
    
    '==': OverloadObject('==', [
        FunctionObject.create (
            'local equal (Number p, Number q) -> Bool',
            a => a.p == a.q
        ),
        FunctionObject.create (
            'local string_equal (String s1, String s2) -> Bool',
            a => a.s1 == a.s2
        )
    ]),
    
    '!=': OverloadObject('!=', [
        FunctionObject.create (
            'local not_equal (Number p, Number q) -> Bool',
            a => a.p != a.q
        ),
        FunctionObject.create (
            'local string_not_equal (String s1, String s2) -> Bool',
            a => a.s1 != a.s2
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
