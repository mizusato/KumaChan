'use strict';


/**
 *  Fundamental Functions Definition
 */


pour(K, {
    singleton: OverloadObject('singleton', [
        FunctionObject.create(
            'global singleton (String name) -> Concept',
            a => SingletonObject(a.name)
        )
    ])
})


pour(K, {
    
    'call': OverloadObject('call', [
        FunctionObject.create(
            'local call (Functional f, ImHash argument_table) -> Any',
            a => a.f.call(mapval(a.argument_table.data, ImRef))
        ),
        FunctionObject.create(
            'local call (Functional f, MutHash &argument_table) -> Any',
            a => a.f.call(a.argument_table.data)
        ),
        FunctionObject.create(
            'local call (Functional f, ImList argument_list) -> Any',
            a => a.f.apply.apply(a.f, map(a.argument_list.data, ImRef))
        ),
        FunctionObject.create(
            'local call (Functional f, MutList &argument_list) -> Any',
            a => a.f.apply.apply(a.f, a.argument_list.data)
        ),
        FunctionObject.create(
            'local call (Functional f) -> Functional',
            a => OverloadObject('call_by', [
                FunctionObject.create(
                    'local call_by (ImHash argument_table) -> Any',
                    b => a.f.call(mapval(b.argument_table.data, ImRef))
                ),
                FunctionObject.create(
                    'local call_by (MutHash &argument_table) -> Any',
                    b => a.f.call(b.argument_table.data)
                ),
                FunctionObject.create(
                    'local call_by (ImList argument_list) -> Any',
                    b => a.f.apply.apply(a.f, map(b.argument_list.data, ImRef))
                ),
                FunctionObject.create(
                    'local call_by (MutList &argument_list) -> Any',
                    b => a.f.apply.apply(a.f, b.argument_list.data)
                )
            ])
        ),
        FunctionObject.create(
            'local call (IteratorFunction f) -> Any',
            a => a.f.apply()
        )
    ]),
    
    '>>': OverloadObject('>>', FunctionObject.converge([
        'local pass_to_right (Immutable object, Functional f) -> Any',
        'local pass_to_right (Mutable &object, Functional f) -> Any',    
    ], a => a.f.apply(a.object) )),
    
    '<<': OverloadObject('<<', FunctionObject.converge([
        'local pass_to_left (Functional f, Immutable object) -> Any',
        'local pass_to_left (Functional f, Mutable &object) -> Any',    
    ], a => a.f.apply(a.object) )),
    
    operator_to: OverloadObject('operator_to', FunctionObject.converge([
        'local pass_to_right (Immutable object, Functional f) -> Any',
        'local pass_to_right (Mutable &object, Functional f) -> Any',    
    ], a => a.f.apply(a.object) )),
     
    operator_by: OverloadObject('operator_by', FunctionObject.converge([
        'local pass_to_left (Functional f, Immutable object) -> Any',
        'local pass_to_left (Functional f, Mutable &object) -> Any'    
    ], a => a.f.apply(a.object) )),
    
    next: OverloadObject('next', [
        FunctionObject.create(
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
            FunctionObject.create(
                'local get_iterator (ImList list) -> Iterator',
                a => get(a.list.data, ImRef)
            ),
            FunctionObject.create(
                'local get_iterator (MutList &list) -> Iterator',
                a => get(a.list.data, x => x)
            )
        ])
    })(),
     
    map: OverloadObject('map', list(cat([
        FunctionObject.create(
            'local map (Iterator iterator, Mapper f) -> Iterator',
            a => IteratorObject(function () {
                let value = a.iterator.next()
                if (value != DoneObject) {
                    return a.f.apply(value)
                } else {
                    return DoneObject
                }
            })
        ),
        FunctionObject.create(
            'local map (Iterator iterator) -> Function',
            a => FunctionObject.create(
                'local map_by (Mapper f) -> Iterator',
                b => K.map.apply(a.iterator, b.f)
            )
        )], FunctionObject.converge([
            'local map_list (ImList list, Mapper f) -> Iterator',
            'local map_list (MutList &list, Mapper f) -> Iterator'
        ], a => K.map.apply(K.get_iterator.apply(a.list), a.f)),
        FunctionObject.converge([
            'local map_list (ImList list) -> Function',
            'local map_list (MutList &list) -> Function',
        ], a => FunctionObject.create(
            'local map_list_by (Mapper f) -> Iterator',
            b => K.map.apply(a.list, b.f)
        ))
    ))),
    
    '->': OverloadObject('->', list(cat([
        FunctionObject.create(
            'local map_by_right (Iterator iterator, Mapper f) -> Iterator',
            a => K.map.apply(iterator, f)
        )],
        FunctionObject.converge([
            'local map_by_right (ImList list, Mapper f) -> Iterator',
            'local map_by_right (MutList &list, Mapper f) -> Iterator'
        ], a => K.map.apply(a.list, a.f))
    ))),
     
    '<-': OverloadObject('<-', list(cat([
        FunctionObject.create(
            'local map_by_left (Mapper f, Iterator iterator) -> Iterator',
            a => K.map.apply(iterator, f)
        )],
        FunctionObject.converge([
            'local map_by_left (Mapper f, ImList list) -> Iterator',
            'local map_by_left (Mapper f, MutList &list) -> Iterator'
        ], a => K.map.apply(a.list, a.f))
    ))),
     
    list: OverloadObject('list', [
        FunctionObject.create(
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
     
    count: OverloadObject('count', [
        FunctionObject.create(
            'local count (UnsignedInt n) -> Iterator',
            a => IteratorObject((function () {
                let i = 0
                return (() => (i < a.n)? i++: DoneObject)
            })())
        ),
        FunctionObject.create(
            'local count () -> Iterator',
            a => IteratorObject((function () {
                let limit = Number.MAX_SAFE_INTEGER
                let i = 0 
                return (() => assert(i < limit) && i++)
            })())
        )
    ]),
     
    filter: OverloadObject('filter', list(cat([
        FunctionObject.create(
            'local filter (Iterator iterator, Filter f) -> Iterator',
            a => IteratorObject(function () {
                let err = ErrorProducer(InvalidReturnValue, 'filter')
                let msg = 'filter function should return boolean value'
                let value = a.iterator.next()
                while (value != DoneObject) {
                    let ok = a.f.apply(value)
                    err.assert(is(ok, BoolObject), msg)
                    if (ok) {
                        return value
                    }
                    value = a.iterator.next()
                }
                return DoneObject
            })
        ),
        FunctionObject.create(
            'local filter (Iterator iterator) -> Function',
            a => FunctionObject.create(
                'local filter_by (Filter f) -> Iterator',
                b => K.filter.apply(a.iterator, b.f)
            )
        )], FunctionObject.converge([
            'local filter_list (ImList list, Filter f) -> Iterator',
            'local filter_list (MutList &list, Filter f) -> Iterator'
        ], a => K.filter.apply(K.get_iterator.apply(a.list), a.f)),
        FunctionObject.converge([
            'local filter_list (ImList list) -> Function',
            'local filter_list (MutList &list) -> Function',
        ], a => FunctionObject.create(
            'local filter_list_by (Filter f) -> Iterator',
            b => K.filter.apply(a.list, b.f)
        ))
    ))),
     
    create_iterator: OverloadObject('create_iterator', [
        FunctionObject.create(
            'local create_iterator (IteratorFunction f) -> Iterator',
            a => IteratorObject(a.f)
        )
    ]),
     
    fold: OverloadObject('fold', list(cat([
        FunctionObject.create(
            'local fold (Iterator iterator, Any initial, Reducer f) ->  Any',
            a => fold((function* () {
                let value = a.iterator.next()
                while (value != DoneObject) {
                    yield value
                    value = a.iterator.next()
                }
            })(), a.initial, (e, v) => a.f.apply(e,v))
        ),
        FunctionObject.create(
            'local fold_from (Any initial) -> Functional',
            a => OverloadObject('fold', list(cat([
                FunctionObject.create(
                    'local fold (Iterator iterator) -> Function',
                    b => FunctionObject.create(
                        'local fold_by (Reducer f) -> Any',
                        c => K.fold.apply(b.iterator, a.initial, c.f)
                    )
                )], FunctionObject.converge([
                    'local fold (ImList list) -> Function',
                    'local fold (MutList &list) -> Function'
                ], b => FunctionObject.create(
                    'local fold_by (Reducer f) -> Any',
                    c => K.fold.apply(b.list, a.initial, c.f)
                ))
            )))
        )], FunctionObject.converge([
            'local fold_list (ImList list, Any initial, Reducer f) -> Any',
            'local fold_list (MutList &list, Any initial, Reducer f) -> Any'
        ], a => K.fold.apply(K.get_iterator.apply(a.list), a.initial, a.f) )
    )))
    
})


pour(K, {
    
    operator_is: OverloadObject('operator_is', FunctionObject.converge([
        'local is (Immutable object, Concept concept) -> Bool',
        'local is (Mutable &object, Concept concept) -> Bool',    
    ], a => a.concept.checker(a.object) )),
     
    Is: OverloadObject('Is', [
        FunctionObject.create(
            'local Is (Concept concept) -> Functional',
            a => OverloadObject(`Is('${a.concept.name}')`,
                FunctionObject.converge([
                'local checker (Immutable object) -> Bool',
                'local checker (Mutable &object) -> Bool'
            ], b => a.concept.checker(b.object) ))
        )
    ]),
     
    IsNot: OverloadObject('IsNot', [
        FunctionObject.create(
            'local IsNot (Concept concept) -> Functional',
            a => K.Is.apply(ConceptObject.Complement(a.concept))
        )
    ]),
    
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
