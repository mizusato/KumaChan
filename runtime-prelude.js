'use strict';


/**
 *  Fundamental Functions Definition
 */


pour(K, {
    singleton: OverloadObject('singleton', [
        FunctionObject.create (
            'global singleton (String name) -> Concept',
            a => SingletonObject(a.name)
        )
    ])
})


pour(K, {
    
    'call': OverloadObject('call', [
        FunctionObject.create (
            'local call (Functional f, ImHash argument_table) -> Any',
            a => a.f.call(mapval(a.argument_table.data, ImRef))
        ),
        FunctionObject.create (
            'local call (Functional f, MutHash &argument_table) -> Any',
            a => a.f.call(a.argument_table.data)
        ),
        FunctionObject.create (
            'local call (Functional f, ImList argument_list) -> Any',
            a => a.f.apply.apply(a.f, map(a.argument_list.data, ImRef))
        ),
        FunctionObject.create (
            'local call (Functional f, MutList &argument_list) -> Any',
            a => a.f.apply.apply(a.f, a.argument_list.data)
        ),
        FunctionObject.create (
            'local call (Functional f) -> Functional',
            a => OverloadObject('call_by', [
                FunctionObject.create (
                    'local call_by (ImHash argument_table) -> Any',
                    b => a.f.call(mapval(b.argument_table.data, ImRef))
                ),
                FunctionObject.create (
                    'local call_by (MutHash &argument_table) -> Any',
                    b => a.f.call(b.argument_table.data)
                ),
                FunctionObject.create (
                    'local call_by (ImList argument_list) -> Any',
                    b => a.f.apply.apply(a.f, map(b.argument_list.data, ImRef))
                ),
                FunctionObject.create (
                    'local call_by (MutList &argument_list) -> Any',
                    b => a.f.apply.apply(a.f, b.argument_list.data)
                )
            ])
        ),
        FunctionObject.create (
            'local call (IteratorFunction f) -> Any',
            a => a.f.apply()
        )
    ]),
    
    '>>': OverloadObject('>>', [
        FunctionObject.create (
            'local pass_to_right (Any *object, Functional f) -> Any',
            a => a.f.apply(a.object)
        )
    ]),
    
    '<<': OverloadObject('<<', [
        FunctionObject.create (
            'local pass_to_left (Functional f, Any *object) -> Any',
            a => a.f.apply(a.object)
        )
    ]),
    
    operator_to: OverloadObject('operator_to', [
        FunctionObject.create (
            'local pass_to_right (Any *object, Functional f) -> Any',
            a => a.f.apply(a.object)
        )
    ]),
     
    operator_by: OverloadObject('operator_by', [
        FunctionObject.create (
            'local pass_to_left (Functional f, Any *object) -> Any',
            a => a.f.apply(a.object)
        )
    ]),
    
    next: OverloadObject('next', [
        FunctionObject.create (
            'local next (Iterator iterator) -> Any',
            a => a.iterator.next()
        )
    ]),
     
    get_iterator: (function () {
        function get (list, f) {
            return IteratorObject((function () {
                let index = 0
                return function () {
                    if (index < list.length) {
                        return f(list[index++])
                    } else {
                        return DoneObject
                    }
                }
            })())
        }
        return OverloadObject('get_iterator', [
            FunctionObject.create (
                'local get_iterator (ImList list) -> Iterator',
                a => get(a.list.data, ImRef)
            ),
            FunctionObject.create (
                'local get_iterator (MutList &list) -> Iterator',
                a => get(a.list.data, x => x)
            ),
            FunctionObject.create (
                'local get_iterator (Iterator iterator) -> Iterator',
                a => a.iterator
            )
        ])
    })(),
     
    map: OverloadObject('map', [
        FunctionObject.create (
            'local map (Iterable *iterable, Mapper f) -> Iterator',
            a => (function () {
                let iterator = K.get_iterator.apply(a.iterable)
                let f = (x => a.f.apply(x))
                return IteratorObject(function () {
                    let value = iterator.next()
                    if (value != DoneObject) {
                        return f(value)
                    } else {
                        return DoneObject
                    }
                })
            })()
        ),
        FunctionObject.create (
            'local map (Iterable *iterable) -> Function',
            a => FunctionObject.create (
                'local map_by (Mapper f) -> Iterator',
                b => K.map.apply(a.iterable, b.f)
            )
        )
    ]),
    
    '->': OverloadObject('->', [
        FunctionObject.create (
            'local map_by_right (Iterable *iterable, Mapper f) -> Iterator',
            a => K.map.apply(a.iterable, a.f)
        )
    ]),
     
    '<-': OverloadObject('<-', [
        FunctionObject.create (
            'local map_by_left (Mapper f, Iterable *iterable) -> Iterator',
            a => K.map.apply(a.iterable, a.f)
        )
    ]),
     
    list: OverloadObject('list', [
        FunctionObject.create (
            'local list (Hash *hash) -> List',
            a => ListObject(
                map(a.hash.data, (k, v) => HashObject({ key: k, value: v }))
            )
        ),
        FunctionObject.create (
            'local list (Iterator iterator) -> List',
            a => ListObject(list((function* () {
                let value = a.iterator.next()
                while (value != DoneObject) {
                    yield value
                    value = a.iterator.next()
                }
            })()))
        )
    ]),
    
    cat: OverloadObject('cat', [
        FunctionObject.create (
            'local cat (IterableList *list) -> Iterator',
            a => (function () {
                let iterators = map(a.list.data, x => K.get_iterator.apply(x))
                let i = 0
                return IteratorObject(function () {
                    while (i < iterators.length) {
                        let value = iterators[i].next()
                        if (value != DoneObject) {
                            return value
                        } else {
                            i += 1
                        }
                    }
                    return DoneObject
                })
            })()
        )
    ]),
     
    count: OverloadObject('count', [
        FunctionObject.create (
            'local count (UnsignedInt n) -> Iterator',
            a => IteratorObject((function () {
                let i = 0
                return (() => (i < a.n)? i++: DoneObject)
            })())
        ),
        FunctionObject.create (
            'local count () -> Iterator',
            a => IteratorObject((function () {
                let limit = Number.MAX_SAFE_INTEGER
                let i = 0 
                return (() => assert(i < limit) && i++)
            })())
        )
    ]),
     
    filter: OverloadObject('filter', [
        FunctionObject.create (
            'local filter (Iterable *iterable, Filter f) -> Iterator',
            a => (function () {
                let err = ErrorProducer(InvalidReturnValue, 'filter')
                let msg = 'filter function should return boolean value'
                let iterator = K.get_iterator.apply(a.iterable)
                let f = (x => a.f.apply(x))
                return IteratorObject(function () {
                    let value = iterator.next()
                    while (value != DoneObject) {
                        let ok = f(value)
                        err.assert(is(ok, BoolObject), msg)
                        if (ok) {
                            return value
                        }
                        value = iterator.next()
                    }
                    return DoneObject
                })
            })()
        ),
        FunctionObject.create (
            'local filter (Iterable *iterable) -> Function',
            a => FunctionObject.create (
                'local filter_by (Filter f) -> Iterator',
                b => K.filter.apply(a.iterable, b.f)
            )
        )
    ]),
     
    create_iterator: OverloadObject('create_iterator', [
        FunctionObject.create (
            'local create_iterator (IteratorFunction f) -> Iterator',
            a => IteratorObject(a.f)
        )
    ]),
     
    fold: OverloadObject('fold', [
        FunctionObject.create (
            'local fold (Iterable *iterable, Any *initial, Reducer f) ->  Any',
            a => fold((function* () {
                let iterator = K.get_iterator.apply(a.iterable)
                let value = iterator.next()
                while (value != DoneObject) {
                    yield value
                    value = iterator.next()
                }
            })(), a.initial, (e, v) => a.f.apply(e,v))
        ),
        FunctionObject.create (
            'local fold_from (Any *initial) -> Function',
            a => FunctionObject.create (
                'local fold (Iterable *iterable) -> Function',
                b => FunctionObject.create (
                    'local fold_by (Reducer f) -> Any',
                    c => K.fold.apply(b.iterable, a.initial, c.f)
                )
            )
        )
    ]),
     
    '**': OverloadObject('**', [
        FunctionObject.create (
            'local cart (Iterable *x, Iterable *y) -> Iterator',
            function (a) {
                let [x, y] = map([a.x, a.y], t => K.get_iterator.apply(t))
                return IteratorObject(function () {
                    let u = x.next()
                    let v = y.next()
                    if (u != DoneObject && v != DoneObject) {
                        return ListObject([u, v])
                    } else {
                        return DoneObject
                    }
                })
            }
        )
    ]),
     
    zip: OverloadObject('zip', [
        FunctionObject.create (
            'local zip (IterableHash *hash) -> Iterator',
            a => (function () {
                let iterator_hash = mapval(
                    a.hash.data,
                    x => K.get_iterator.apply(x)
                )
                return IteratorObject(function () {
                    let value_hash = mapval(iterator_hash, it => it.next())
                    let values = map_lazy(value_hash, (k, v) => v)
                    if ( forall(values, v => v != DoneObject) ) {
                        return HashObject(value_hash)
                    } else {
                        return DoneObject
                    }
                })
            })()
        )
    ])
    
})


pour(K, {
    
    operator_is: OverloadObject('operator_is', [
        FunctionObject.create (
            'local is (Any *object, Concept concept) -> Bool',
            a => a.concept.checker(a.object)
        )
    ]),
     
    Is: OverloadObject('Is', [
        FunctionObject.create (
            'local Is (Concept concept) -> Function',
            a => FunctionObject(
                'local checker (Any *object) -> Bool',
                b => a.concept.checker(b.object)
            )
        )
    ]),
     
    IsNot: OverloadObject('IsNot', [
        FunctionObject.create (
            'local IsNot (Concept concept) -> Functional',
            a => K.Is.apply(ConceptObject.Complement(a.concept))
        )
    ]),
    
    '|': OverloadObject('|', [ FunctionObject.create (
        'local union (Concept c1, Concept c2) -> Concept',
        a => ConceptObject.Union(a.c1, a.c2)            
    )]),
     
    '&': OverloadObject('&', [ FunctionObject.create (
        'local intersect (Concept c1, Concept c2) -> Concept',
        a => ConceptObject.Intersect(a.c1, a.c2)
    )]),
     
    '~': OverloadObject('~', [ FunctionObject.create (
        'local complement (Concept c) -> Concept',
        a => ConceptObject.Complement(a.c)
    )]),
    
    '||': OverloadObject('||', [ FunctionObject.create (
        'local or (Bool v1, Bool v2) -> Bool',
         a => a.v1 || a.v2
    )]),
    
    '&&': OverloadObject('&&', [ FunctionObject.create (
        'local and (Bool v1, Bool v2) -> Bool',
        a => a.v1 && a.v2
    )]),
     
    '!': OverloadObject('!', [ FunctionObject.create (
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

    Im: OverloadObject('Im', [
        FunctionObject.create (
            'local Im (Any *object) -> Immutable',
            a => ImRef(a.object)
        )
    ]),
     
    Id: OverloadObject('Id', [
        FunctionObject.create (
            'local Id (Any *object) -> Any',
            a => a.object
        )
    ]),

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
        FunctionObject.create (
            'local clone (Atomic object) -> Atomic',
            a => a.object
        ),
        FunctionObject.create (
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
     
    append: OverloadObject('append', [
        FunctionObject.create (
            'local append (MutList &self, Any *element) -> Void',
            a => a.self.append(a.element)
        )
    ]),
    
    length: OverloadObject('length', [
        FunctionObject.create (
            'local length (List self) -> Size',
            a => a.self.length()
        )
    ])

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
    
    set: OverloadObject('set', [
        FunctionObject.create (
            'local set (MutHash &self, String key, Any *value) -> Void',
            a => a.self.set(a.key, a.value)
        )
    ])
    
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
        FunctionObject.create (
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
            'local equal (Any left, Any right) -> Bool',
            a => false
        ),
        FunctionObject.create (
            'local equal (Compound left, Compound right) -> Bool',
            a => a.left.data === a.right.data
        ),
        FunctionObject.create (
            'local equal (Atomic left, Atomic right) -> Bool',
            a => a.left === a.right
        )
    ]),
    
    '!=': OverloadObject('!=', [
        FunctionObject.create (
            'local not_equal (Any left, Any right) -> Bool',
            a => !(K['=='].apply(a.left, a.right))
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
