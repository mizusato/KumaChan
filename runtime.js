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
 *  and outside variables that can be affected by the function,
 *  which indicates the magnitude of side-effect.
 *  
 *  |  value  | Global Scope | Upper Scope | Local Scope | Dirty Argument |
 *     global       rw             rw            rw             rw           -
 *     nearby       ro             rw            rw             rw 
 *     local        ro             ro            rw             rw
 *     dirty       dump           dump           rw             rw        
 *     cached      dump           dump           rw             --
 *      pure       dump           dump           rw             --
 *               
 *  Note: The functions with effect range in {global, nearby, pure}
 *        are called affected functions, that means, these functions
 *        may be affected by the outside state. 
 *        The functions with effect range in {dirty, cache, pure}
 *        are called independent functions, that means, these functions
 *        will never be affected by the outside state. This restriction
 *        is implemented by dumping a part of outside *solid* variables
 *        when we *define* the function, instead of lazy evaluation.
 *        That means when we execute the function, we won't be able to
 *        refer the variables outside its pure scope, but we can refer
 *        the variables which are dumped at the function definition.
 *        However, if the effect range value were "dirty", the function
 *        could modify its argument to affect the outside scope.
 *        The functions with effect range in {cached, pure} are called
 *        pure functions. Pure function cannot be affected by the outside
 *        state, and won't modify the outside scopes or its arguments.
 *        If the effect range value were set to "cached", that means
 *        the value of function will be cached, which indicates the
 *        function will modify (affect) the caches when we invoke it.
 */


const EffectRange = Enum('global', 'nearby', 'pure', 'dirty', 'cached', 'pure')
pour(EffectRange, {
    affected: one_of('global', 'nearby', 'pure'),
    independent: one_of('dirty', 'cached', 'pure'),
    pure: one_of('cached', 'pure')
})


/**
 *  Object Type Definition
 * 
 *  Object (Any) ┬ Compound ┬ Hash
 *               │          ┴ List
 *               ┴ Atomic ┬ Function
 *                        ┼ Concept
 *                        ┼ Chain
 *                        ┴ Primitive ┬ String
 *                                    ┼ Number
 *                                    ┴ Bool
 *
 *  Note: atomic object must be totally immutable.
 */


/* Primitive Definition */


const StringObject = $(x => typeof x == 'string')
const NumberObject = $(x => typeof x == 'number')
const BoolObject = $(x => typeof x == 'boolean')
const PrimitiveObject = $u(StringObject, NumberObject, BoolObject)


/* Atomic Definition */


const AtomicObject = $u(
    PrimitiveObject,
    $(x => x.is(FunctionObject)),
    $(x => x.is(ChainObject)),
    $(x => x.is(ConceptObject))
)


/* Hash Definition */


const Config = {
    default: {
        immutable: false,
        frozen: false
    },
    from: object => pour({}, object.config),
    get_flags: Detail.Config.get_flags
}


function HashObject (hash = {}, config = Config.default) {
    assert(Hash.contains(hash))
    assert(Hash.contains(config))
    return {
        data: hash,
        config: config,
        maker: HashObject,
        __proto__: once(HashObject, Detail.Hash.get_prototype())
    }
}


SetMakerConcept(HashObject)


const ImHashObject = $n(HashObject, $(x => x.config.immutable))
const MutHashObject = $n(HashObject, $(x => !x.config.immutable))
const FzHashObject = $n(ImHashObject, $(x => x.config.frozen))


/* List Definition */


function ListObject (list = [], config = Config.default) {
    assert(list.is(Array))
    assert(Hash.contains(config))
    return {
        data: list,
        config: config,
        maker: ListObject,
        __proto__: once(ListObject, Detail.List.get_prototype())
    }
}


SetMakerConcept(ListObject)


const ImListObject = $n(ListObject, $(x => x.config.immutable))
const MutListObject = $n(ListObject, $(x => !x.config.immutable))
const FzListObject = $n(ImListObject, $(x => x.config.frozen))


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
        Chain: $(x => ChainObject.contains(x)),
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
            assert(r != NA)
            return r
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
 *  Frozen and Solid Definition
 */

const FrozenObject = $u(FzHashObject, FzListObject)
const SolidObject = $u(AtomicObject, FrozenObject)


/**
 *  Immutably Refer, Clone and Freeze
 */


function ImRef (object) {
    if ( object.is(MutableObject) ) {
        return object.maker(
            object.data,
            pour(Config.from(object), {
                immutable: true
            })
        )
    } else {
        return object
    }
}


function ForceMutable (object) {
    if ( object.is(ImmutableObject) ) {
        return object.maker(
            object.data,
            pour(Config.from(object), {
                immutable: false
            })
        )
    } else {
        return object
    }
}


function Clone (object) {
    if ( object.is(CompoundObject) ) {
        if ( object.is(HashObject) ) {
            return HashObject(
                mapval(object.data, v => Clone(v)),
                pour(Config.from(object), {
                    immutable: false,
                    frozen: false
                })
            )
        } else if ( object.is(ListObject) ) {
            return HashObject(
                map(object.data, v => Clone(v)),
                pour(Config.from(object), {
                    immutable: false,
                    frozen: false
                })
            )
        }
    } else {
        return object
    }
}


function Freeze (object) {
    if ( object.is(CompoundObject) ) {
        if ( object.is(HashObject) ) {
            return HashObject(
                mapval(object.data, v => Freeze(v)),
                pour(Config.from(object), {
                    immutable: true,
                    frozen: true
                })
            )
        } else if ( object.is(ListObject) ) {
            return ListObject(
                map(object.data, v => Freeze(v)),
                pour(Config.from(object), {
                    immutable: true,
                    frozen: true
                })
            )
        }
    } else {
        return object
    }
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


function Scope (context, data = {}) {
    assert(context.is(Scope))
    assert(Hash.contains(data))
    return {
        data: data,
        context: context,
        maker: Scope
    }
}


SetEquivalent(Scope, $u(NullScope, MadeBy(Scope)) )


/**
 *  Global Object Definition
 */


const G = Scope(NullScope)
const K = G.data
K.scope = G


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
    check(ConceptObject, arguments, { name: Str, f: Function })
    return {
        name: name,
        checker: ConceptChecker(`${name}`, f),
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
    constraint: ConceptObject,
    pass_policy: PassPolicy
})


const Prototype = $n(
    Struct({
        effect_range: EffectRange,
        parameters: HashOf(Parameter),
        order: ArrayOf(Str),
        return_value: ConceptObject
    }),
    $( proto => forall(proto.order, key => proto.parameters[key]) )
)


pour(Prototype, Detail.Prototype)


/* Function Definition */


function FunctionObject (name, context, prototype, js_function) {
    check(FunctionObject, arguments, {
        name: Str,
        context: Scope,
        prototype: Prototype,
        js_function: Function
    })
    context = {
        definition: context,
        execution: (
            prototype.effect_range.is(EffectRange.affected)?
            context: NullScope
        )
    }
    if ( prototype.effect_range.is(EffectRange.pure) ) {
        let err = ErrorProducer(InvalidDefinition, 'Function::Creator')
        err.assert(
            forall(prototype.parameters, p => p.pass_policy != 'dirty'),
            'pure function cannot have dirty parameter'
        )
    }
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
            call: function (argument) {
                assert(HashOf(ObjectObject).contains(argument))
                let err_a = ErrorProducer(InvalidArgument, `${name}`)
                let err_r = ErrorProducer(InvalidReturnValue, `${name}`)
                let Proto = Prototype
                let Range = EffectRange
                let proto = this.prototype
                let range = proto.effect_range
                let f = this.js_function
                err_a.if_failed(Proto.check_argument(proto, argument))
                let normalized = Proto.normalize_argument(proto, argument)
                let final = Proto.set_mutability(proto, normalized)
                let scope = Scope(context.execution, {
                    callee: this,
                    argument: HashObject(final),
                    argument_info: HashObject(
                        mapval(normalized, v => HashObject({
                            is_immutable: v.is(ImmutableObject)
                        }))
                    )
                })
                scope.data.scope = HashObject(scope.data)
                pour(scope.data, final)
                let value = f(scope)
                let arg_str = Detail.Argument.represent(normalized)
                let val_str = ObjectObject.represent(value)
                console.log(`${this.name}${arg_str} = ${val_str}`)
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
    effect_range: 'pure',
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
                effect_range: $1('pure'),
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
    Chain: PortConcept(ChainObject, 'Chain'),
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
    FzHash: PortConcept(FzHashObject, 'FzHash'),
    FzList: PortConcept(FzListObject, 'FzList'),
    Immutable: ImmutableConcept,
    Mutable: MutableConcept,
    Frozen: PortConcept(FrozenObject, 'Frozen'),
    Solid: PortConcept(SolidObject, 'Solid')
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


function ChainObject (name, instances) {
    check(ChainObject, arguments, {
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
        maker: ChainObject,
        __proto__: once(ChainObject, {
            added: function (instance) {
                assert(instance.is(FunctionObject))
                let new_list = map(instances, x=>x)
                new_list.push(instance)
                return ChainObject(name, new_list)
            },            
            apply: function (...args) {
                assert(args.is(ArrayOf(ObjectObject)))
                return this.call(fold(args, {}, (e, v, i) => (v[i] = e, v)) )
            },
            call: function (argument) {
                assert(HashOf(ObjectObject).contains(argument))
                for(let instance of rev(this.instances)) {
                    let p = Prototype
                    let check = p.check_argument(instance.prototype, argument)
                    if ( check === OK ) {
                        return instance.call(argument)
                    }
                }
                let err = ErrorProducer(NoMatchingPattern, `${this.name}`)
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


SetMakerConcept(ChainObject)


const HasMethod = (...names) => $(
    x => assert(x.is(ObjectObject))
        && forall(names, name => G.has(name) && K[name].is(ChainObject)
                  && K[name].has_method_of(x))
)


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
    
    is: ChainObject('is', FunctionObject.converge([
        'pure Immutable::is (Immutable object, Concept concept) -> Bool',
        'dirty Mutable::is (Mutable &object, Concept concept) -> Bool',    
    ], a => a.concept.checker.apply(a.object) )),
    
    union: ChainObject('union', [ FunctionObject.create(
        'pure union (Concept c1, Concept c2) -> Concept',
        a => ConceptObject.Union(a.c1, a.c2)            
    )]),
     
    intersect: ChainObject('intersect', [ FunctionObject.create(
        'pure intersect (Concept c1, Concept c2) -> Concept',
        a => ConceptObject.Intersect(a.c1, a.c2)
    )]),
     
    complement: ChainObject('complement', [ FunctionObject.create(
        'pure complement (Concept c) -> Concept',
        a => ConceptObject.Complement(a.c)
    )])
    
})


pour(K, {
    
    type_of: ChainObject('type_of', [
        FunctionObject.create (
            'pure type_of (Any object) -> String',
            a => GetType(a.object)
        )
    ]),

    Im: ChainObject('Im', FunctionObject.converge([
        'pure Im (Immutable object) -> Immutable',
        'dirty Im (Mutable &object) -> Immutable'
    ], a => ImRef(a.object) )),
     
    Id: ChainObject('Id', FunctionObject.converge([
        'pure Id (Immutable object) -> Immutable',
        'dirty Id (Mutable &object) -> Mutable'
    ], a => a.object )),

    copy: ChainObject('copy', [
        FunctionObject.create (
            'pure copy (Atomic object) -> Atomic',
            a => a.object
        ),
        FunctionObject.create (
            'pure copy (ImList object) -> MutList',
            a => ListObject(map(a.object.data, e => ImRef(e)))
        ),
        FunctionObject.create (
            'dirty copy (MutList &object) -> MutList',
            a => ListObject(map(a.object.data, e => e))
        ),
        FunctionObject.create (
            'pure copy (ImHash object) -> MutHash',
            a => HashObject(mapval(a.object.data, v => ImRef(v)))
        ),
        FunctionObject.create (
            'dirty copy (MutHash &object) -> MutHash',
            a => HashObject(mapval(a.object.data, v => v))
        ),
    ]),
    
    clone: ChainObject('clone', [
        FunctionObject.create(
            'pure clone (Atomic object) -> Atomic',
            a => a.object
        ),
        FunctionObject.create(
            'pure clone (Compound object) -> Mutable',
            a => Clone(a.object)
        )
    ]),
    
    freeze: ChainObject('freeze', [ FunctionObject.create(
        'pure freeze (Any object) -> Solid',
        a => Freeze(a.object)
    )])
    
})


pour(K, {
    
    at: ChainObject('at', [
        FunctionObject.create (
            'pure ImList::at (ImList self, Index index) -> Immutable',
            a => ImRef(a.self.at(a.index))
        ),
        FunctionObject.create (
            'dirty MutList::at (MutList &self, Index index) -> Object',
            a => a.self.at(a.index)
        )
    ]),
     
    append: ChainObject('append', FunctionObject.converge([
        'dirty MutList::append (MutList &self, Immutable element) -> Void',
        'dirty MutList::append (MutList &self, Mutable &element) -> Void'
    ], a => a.self.append(a.element) )),
    
    length: ChainObject('length', [ FunctionObject.create (
        'pure List::length (List self) -> Size',
        a => a.self.length()
    )])

})


pour(K, {
    has: ChainObject('has', [
        FunctionObject.create (
            'pure Hash::has (Hash self, String key) -> Bool',
            a => a.self.has(a.key)
        )
    ]),
    get: ChainObject('get', [
        FunctionObject.create (
            'pure ImHash::get (ImHash self, String key) -> Immutable',
            a => ImRef(a.self.get(a.key))
        ),
        FunctionObject.create (
            'dirty MutHash::get (MutHash &self, String key) -> Object',
            a => a.self.get(a.key)
        ),
        /* List also have a get function, which calls at() */
        FunctionObject.create (
            'pure ImList::get (ImList self, Index index) -> Immutable',
            a => ImRef(a.self.at(a.index))
        ),
        FunctionObject.create (
            'dirty MutList::get (MutList &self, Index index) -> Object',
            a => a.self.at(a.index)
        )
    ]),
    fetch: ChainObject('fetch', [
        FunctionObject.create (
            'pure ImHash::fetch (ImHash self, String key) -> Immutable',
            a => ImRef(a.self.fetch(a.key))
        ),
        FunctionObject.create (
            'dirty MutHash::fetch (MutHash &self, String key) -> Object',
            a => a.self.fetch(a.key)
        )
    ]),
    set: ChainObject('set', FunctionObject.converge([
        'dirty MutHash::set (MutHash &self, String key, Immutable value) -> Void',
        'dirty MutHash::set (MutHash &self, String key, Mutable &value) -> Void'
    ], a => a.self.set(a.key, a.value) ))
})


pour(K, {
    plus: ChainObject('plus', [
        FunctionObject.create (
            'pure plus (Number p, Number q) -> Number',
            a => a.p + a.q
        ),
        FunctionObject.create (
            'pure plus (String s1, String s2) -> String',
            a => a.s1 + a.s2
        )
    ]),
    minus: ChainObject('minus', [
        FunctionObject.create (
            'pure minus (Number p, Number q) -> Number',
            a => a.p - a.q
        ),
        FunctionObject.create (
            'pure minus (Number x) -> Number',
            a => -a.x
        ),
        FunctionObject.create(
            'pure minus (Concept c1, Concept c2) -> Concept',
            a => ConceptObject.Intersect(
                a.c1, ConceptObject.Complement(a.c2)
            )
        )
    ]),
    multiply: ChainObject('multiply', [
        FunctionObject.create (
            'pure multiply (Number p, Number q) -> Number',
            a => a.p * a.q
        )
    ]),
    divide: ChainObject('divide', [
        FunctionObject.create (
            'pure divide (Number p, Number q) -> Number',
            a => a.p / a.q
        )
    ]),
    mod: ChainObject('mod', [
        FunctionObject.create (
            'pure mod (Number p, Number q) -> Number',
            a => a.p % a.q
        )
    ]),
    pow: ChainObject('pow', [
        FunctionObject.create (
            'pure pow (Number p, Number q) -> Number',
            a => Math.pow(a.p, a.q)
        )
    ]),
    floor: ChainObject('floor', [
        FunctionObject.create (
            'pure floor (Number x) -> Number',
            a => Math.floor(a.x)
        )
    ]),
    ceil: ChainObject('ceil', [
        FunctionObject.create (
            'pure ceil (Number x) -> Number',
            a => Math.ceil(a.x)
        )
    ]),
    round: ChainObject('round', [
        FunctionObject.create (
            'pure round (Number x) -> Number',
            a => Math.round(a.x)
        )
    ])
})
