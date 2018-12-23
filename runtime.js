'use strict';


/**
 *  Exceptions Definition
 */


class RuntimeError extends Error {}
class InvalidOperation extends RuntimeError {}
class InvalidArgument extends RuntimeError {}
class InvalidReturnValue extends RuntimeError {}
class InvalidDefinition extends RuntimeError {}
class NoMatchingPattern extends RuntimeError {}
class KeyError extends RuntimeError {}
class IndexError extends RuntimeError {}
class NameConflict extends RuntimeError {}
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
                    /([a-z])([A-Z])/g, '$1 $2'
                )
                throw new err_class(`${f_name}: ${err_type}: ${err_msg}`)
            }
        },
        assert: function (bool, err_msg) {
            check(this.assert, arguments, { bool: Bool, err_msg: Str })
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
 *     pure        ro             ro            rw             rw
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
    from: object => pour({}, object.config)
}


function HashObject (hash = {}, config = Config.default) {
    assert(hash.is(Hash))
    assert(config.is(Hash))
    return {
        data: hash,
        config: config
        maker: HashObject,
        __proto__: once({
            has: function (key) { return this.data.has(key) },
            get: function (key) {
                let err = ErrorProducer(KeyError, 'Hash::get')
                err.assert(this.data.has(key), `'${a.key}' does not exist`)
                return this.data[key]
            },
            fetch: function (key) {
                return this.data.has(key) && this.data[key] || NaObject
            },
            set: function (key, value) {
                this.data[key] = value
                return VoidObject
            },
            emplace: function (key, value) {
                let err = ErrorProducer(KeyError, 'Hash::emplace')
                err.if(this.data.has(key), `'${a.key}' already exist`)
                this.data[key] = value
                return VoidObject
            },
            replace: function (key, value) {
                let err = ErrorProducer(KeyError, 'Hash::replace')
                err.assert(this.data.has(key), `'${a.key}' does not exist`)
                this.data[key] = value
                return VoidObject
            },
            take: function (key) {
                let err = ErrorProducer(KeyError, 'Hash::take')
                err.assert(this.data.has(key), `'${a.key}' does not exist`)
                let value = this.data[key]
                delete this.data[key]
                return value
            }
        })
    }
}


SetMakerConcept(HashObject)


const ImHashObject = $n(HashObject, $(x => x.config.immutable))
const MutHashObject = $n(HashObject, $(x => !x.config.immutable))
const FzHashObject = $n(ImHashObject, $(x => x.config.frozen))


/* List Definition */


function ListObject (list = [], config = Config.default) {
    assert(list.is(Array))
    assert(config.is(Hash))
    return {
        data: list,
        config: config,
        maker: ListObject,
        __proto__: once({
            length: function () { return this.data.length },
            at: function (index) {
                let err = ErrorProducer(IndexError, 'List::at')
                err.assert(index < this.data.length, `${a.index}`)
                assert(typeof this.data[index] != 'undefined')
                return this.data[index]
            },
            append: function (element) {
                this.data.push(element)
                return VoidObject
            }
        })
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
        return HashObject(
            object.data,
            pour(Config.from(object), {
                immutable: true
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
        } else if ( object.is(ListObject) )
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


const NotNullScope = $n(HashObject, Struct({
    context: $u( NullScope, $(x => x.is(NotNullScope)) ),
    range: EffectRange
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
        maker: ConceptObject
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
    $( proto => forall(proto.order, key => proto.parameters.has(key)) )
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
    let Proto = Prototype
    let Range = EffectRange
    let proto = prototype
    let range = proto.effect_range
    let f = js_function
    let context = {
        definition: context,
        execution: range.is(Range.affected)? context: NullScope
    }
    if ( range.is(Range.pure) ) {
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
                assert(argument.is(HashOf(ObjectObject)))
                let err_a = ErrorProducer(InvalidArgument, `${name}()`)
                let err_r = ErrorProducer(InvalidReturnValue, `${name}()`)
                err_a.if_failed(p.check_argument(proto, argument))
                let normalized = p.normalize_argument(proto, argument)
                let scope = Scope(context.execution, {
                    callee: this,
                    argument: HashObject(normalized),
                    argument_info: HashObject(
                        mapval(normalized, v => HashObject({
                            is_immutable: v.is(ImmutableObject)
                        }))
                    )
                })
                scope.data.scope = HashObject(scope.data)
                pour(scope.data, normalized)
                let value = f(scope)
                if (this !== BoolConcept.checker) {
                    err_r.if_failed(p.check_return_value(proto, value))
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
        scope => f(scope.data.object, scope.data.argument_info.data.object)
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
    let err = ErrorProducer(NameConflict, 'Singleton::Creator()')
    err.if(SingletonOfName.has(name), 'singleton name ${name} is in use')
    let singleton = {}
    pour(singleton, ConceptObject(
        `Singleton<'${name}'>`,
        object => object === singleton
    )
    pour(singleton, {
        contains: x => x === singleton,
        singleton_name: name
    })
    SingletonOfName[name] = singleton
    return singleton
}


SetEquivalent(SingletonObject, $n(
    ConceptObject,
    $(x => x.has(singleton_name)),
    $(x => x === SingletonOfName[x.singleton_name])
))


/* Fix NullScope */


pour(NullScope, SingletonObject('NullScope'))


/* Default Singleton Objects */


const VoidObject = Singleton('Void')
const NaObject = Singleton('N/A')


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


const ImmutableConcept = ConceptObject('Immutable', (_, info) => info.is_immutable)
const MutableConcept = ConceptObject.Complement(ImmutableConcept)


function PortConcept(concept, name) {
    check(PortConcept, arguments, {
        concept: Concept, name: Str
    })
    return ConceptObject(name, x => x.is(concept))
}


pour(PortConcept, {
    Immutable: (c, n) => ConceptObject.Intersect(
        ImmutableConcept, PortConcept(c, n)
    ),
    Mutable: (c, n) => ConceptObject.Intersect(
        MutableConcept, PortConcept(c, n)
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
    ImHash: PortConcept.Immutable(HashObject, 'ImHash'),
    MutHash: PortConcept.Mutable(HashObject, 'MutHash'),
    ImList: PortConcept.Immutable(ListObject, 'ImList'),
    MutList: PortConcept.Mutable(ListObject, 'MutList'),
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
                assert(argument.is(HashOf(ObjectObject)))
                for(let instance of rev(this.instances)) {
                    let p = Prototype
                    let check = p.check_argument(instance.prototype, argument)
                    if ( check === OK ) {
                        return instance.call(argument)
                    }
                }
                let err = ErrorProducer(NoMatchingPattern, `${this.name}()`)
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
        && forall(names, name => K.has(name) && K[name].is(ChainObject)
                  && K[name].has_method_of(x))
)


/**
 *  Fundamental Functions Definition
 */

Function.create(
    'global singleton (String name) -> Singleton',
    a => SingletonObject(a.name)
)


pour(K, {
    
    is: ChainObject('is', Function.converge([
        'pure Immutable::is (Immutable object, Concept concept) -> Bool',
        'dirty Mutable::is (Mutable &object, Concept concept) -> Bool',    
    ], a => a.concept.data.checker.apply(a.object) )),
    
    union: ChainObject('union', [ Function.create(
        'pure union (Concept c1, Concept c2) -> Concept',
        a => ConceptObject.Union(a.c1, a.c2)            
    )]),
     
    intersect: ChainObject('intersect', [ Function.create(
        'pure intersect (Concept c1, Concept c2) -> Concept',
        a => ConceptObject.Intersect(a.c1, a.c2)
    )]),
     
    complement: ChainObject('complement', [ Function.create(
        'pure complement (Concept c) -> Concept',
        a => ConceptObject.Complement(a.c)
    )])
    
})


pour(K, {

    ImRef: ChainObject('ImRef', Function.converge([
        'pure ImRef (Immutable object) -> Immutable',
        'dirty ImRef (Mutable &object) -> Immutable'
    ], a => ImRef(a.object) )),
     
    Id: ChainObject('Id', Function.converge([
        'pure Id (Immutable object) -> Immutable',
        'dirty Id (Mutable &object) -> Mutable'
    ], a => a.object ))

    copy: ChainObject('copy', [
        Function.create (
            'pure copy (Atomic object) -> Atomic',
            a => a.object
        ),
        Function.create (
            'pure copy (ImList object) -> MutList',
            a => ListObject(map(a.object.data, e => ImRef(e)))
        ),
        Function.create (
            'dirty copy (MutList &object) -> MutList',
            a => ListObject(map(a.object.data, e => e))
        ),
        Function.create (
            'pure copy (ImHash object) -> MutHash',
            a => HashObject(mapval(a.object.data, v => ImRef(v)))
        ),
        Function.create (
            'dirty copy (MutHash &object) -> MutHash',
            a => HashObject(mapval(a.object.data, v => v))
        ),
    ]),
    
    clone: ChainObject('clone', [
        Function.create(
            'pure clone (Atomic object) -> Atomic',
            a => a.object
        ),
        Function.create(
            'pure clone (Compound object) -> Mutable',
            a => Clone(a.object)
        )
    ]),
    
    freeze: ChainObject('freeze', [ Function.create(
        'pure freeze (Any object) -> Solid',
        a => Freeze(a.object)
    )])
    
})


pour(K, {
    
    at: ChainObject('at', [
        Function.create (
            'pure ImList::at (ImList self, Index index) -> Immutable',
            a => ImRef(a.self.at(a.index))
        ),
        Function.create (
            'dirty MutList::at (MutList &self, Index index) -> Object',
            a => a.self.at(a.index)
        )
    ]),
     
    append: ChainObject('append', Function.converge([
        'dirty MutList::append (MutList &self, Immutable element) -> Void',
        'dirty MutList::append (MutList &self, Mutable &element) -> Void'
    ], a => a.self.append(a.element) )),
    
    length: ChainObject('length', [ Function.create (
        'pure List::length (List self) -> Size',
        a => a.self.length()
    )])

})


pour(K, {
    has: ChainObject('has', [
        Function.create (
            'pure Hash::has (Hash self, String key) -> Bool',
            a => a.self.has(a.key)
        )
    ]),
    get: ChainObject('get', [
        Function.create (
            'pure ImHash::get (ImHash self, String key) -> Immutable',
            a => ImRef(a.self.get(a.key))
        ),
        Function.create (
            'dirty MutHash::get (MutHash &self, String key) -> Object',
            a => a.self.get(a.key)
        ),
        /* List also have a get function, which calls at() */
        Function.create (
            'pure ImList::get (ImList self, Index index) -> Immutable',
            a => ImRef(a.self.at(a.index))
        ),
        Function.create (
            'dirty MutList::get (MutList &self, Index index) -> Object',
            a => a.self.at(a.index)
        )
    ]),
    fetch: ChainObject('fetch', [
        Function.create (
            'pure ImHash::fetch (ImHash self, String key) -> Immutable',
            a => ImRef(a.self.fetch(a.key))
        ),
        Function.create (
            'dirty MutHash::fetch (MutHash &self, String key) -> Object',
            a => a.self.fetch(a.key)
        )
    ]),
    set: ChainObject('set', Function.converge([
        'dirty MutHash::set (MutHash &self, String key, Immutable value) -> Void',
        'dirty MutHash::set (MutHash &self, String key, Mutable &value) -> Void'
    ], a => a.self.set(a.key, a.value) ))
})


pour(K, {
    plus: ChainObject('plus', [
        Function.create (
            'pure plus (Number p, Number q) -> Number',
            a => a.p + a.q
        ),
        Function.create (
            'pure plus (String s1, String s2) -> String',
            a => a.s1 + a.s2
        )
    ]),
    minus: ChainObject('minus', [
        Function.create (
            'pure minus (Number p, Number q) -> Number',
            a => a.p - a.q
        ),
        Function.create (
            'pure minus (Number x) -> Number',
            a => -a.x
        ),
        Function.create(
            'pure minus (Concept c1, Concept c2) -> Concept',
            a => ConceptObject.Intersect(
                a.c1, ConceptObject.Complement(a.c2)
            )
        )
    ]),
    multiply: ChainObject('multiply', [
        Function.create (
            'pure multiply (Number p, Number q) -> Number',
            a => a.p * a.q
        )
    ]),
    divide: ChainObject('divide', [
        Function.create (
            'pure divide (Number p, Number q) -> Number',
            a => a.p / a.q
        )
    ]),
    mod: ChainObject('mod', [
        Function.create (
            'pure mod (Number p, Number q) -> Number',
            a => a.p % a.q
        )
    ]),
    pow: ChainObject('pow', [
        Function.create (
            'pure pow (Number p, Number q) -> Number',
            a => Math.pow(a.p, a.q)
        )
    ]),
    is_finite: ChainObject('is_finite', [
        Function.create (
            'pure Number::is_finite (Number self) -> Bool',
            a => Number.isFinite(a.self)
        )
    ]),
    is_NaN: ChainObject('is_NaN', [
        Function.create (
            'pure Number::is_NaN (Number self) -> Bool',
            a => Number.isNaN(a.self)
        )
    ]),
    floor: ChainObject('floor', [
        Function.create (
            'pure floor (Number x) -> Number',
            a => Math.floor(a.x)
        )
    ]),
    ceil: ChainObject('ceil', [
        Function.create (
            'pure ceil (Number x) -> Number',
            a => Math.ceil(a.x)
        )
    ]),
    round: ChainObject('round', [
        Function.create (
            'pure round (Number x) -> Number',
            a => Math.round(a.x)
        )
    ])
})


