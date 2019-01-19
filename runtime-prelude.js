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
     
    create_concept: OverloadObject('create_concept', [
        FunctionObject.create (
            'local create_concept (Filter f) -> Concept',
            a => ConceptObject('{Temp}', function (object) {
                let err = ErrorProducer(InvalidReturnValue, 'create_concept')
                let msg = 'filter function should return boolean value'
                let ok = a.f.apply(object)
                err.assert(is(ok, BoolObject), msg)
                return ok
            })
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
    ]),
     
    keys: OverloadObject('keys', [
        FunctionObject.create (
            'local keys (PairList pairs) -> List',
            a => ListObject(map(a.pairs.data, x => x.get('key')))
        ),
        FunctionObject.create (
            'local keys (Hash hash) -> List',
            a => ListObject(map(a.hash.data, (k, v) => k))
        )
    ]),
     
    values: OverloadObject('values', [
        FunctionObject.create (
            'local values (PairList *pairs) -> List',
            a => ListObject(map(a.pairs.data, x => x.get('value')))
        ),
        FunctionObject.create (
            'local values (Hash *hash) -> List',
            a => ListObject(map(a.hash.data, (k, v) => v))
        )
    ]),
     
    mapkey: OverloadObject('mapkey', [
        FunctionObject.create (
            'local mapkey (Hash *hash, Mapper f) -> Hash',
            (function () {
                let err = ErrorProducer(InvalidReturnValue, 'mapkey')
                let msg = 'mapper function of mapkey() should return string'
                return a => HashObject(mapkey(a.hash.data, function (key) {
                    let new_key = a.f.apply(key)
                    err.assert(is(new_key, StringObject), msg)
                    return new_key
                }))
            })()
        ),
        FunctionObject.create (
            'local mapkey (Hash *hash) -> Function',
            a => FunctionObject.create (
                'local mapkey_by (Mapper f) -> Hash',
                b => K.mapkey.apply(a.hash, b.f)
            )
        )
    ]),
     
    mapval: OverloadObject('mapval', [
        FunctionObject.create (
            'local mapval (Hash *hash, Mapper f) -> Hash',
            a => HashObject(mapval(a.hash.data, v => a.f.apply(v)))
        ),
        FunctionObject.create (
            'local mapval (Hash *hash) -> Function',
            a => FunctionObject.create (
                'local mapval_by (Mapper f) -> Hash',
                b => K.mapval.apply(a.hash, b.f)
            )
        )
    ]),
     
    '=>': OverloadObject('=>', [
        FunctionObject.create (
            'local switch (Any *object, CaseList *case_list) -> Any',
            function (a) {
                let err = ErrorProducer(InvalidReturnValue, 'switch')
                let msg = 'filter function should return boolean value'
                let require_bool = (
                    x => (err.assert(is(x, BoolObject), msg), x)
                )
                let object = a.object
                let case_list = a.case_list
                let match = find(map_lazy(case_list.data, function (pair) {
                    let key = pair.get('key')
                    let value = pair.get('value')
                    let ok = transform(key, [
                        { when_it_is: FilterObject,
                          use: f => require_bool(f.apply(object)) },
                        { when_it_is: ConceptObject,
                          use: c => c.checker(object) },
                        { when_it_is: BoolObject,
                          use: b => b }
                    ])
                    return {
                        ok: ok,
                        value: value
                    }
                }), x => x.ok)
                if (match != NotFound) {
                    return match.value
                } else {
                    let err = ErrorProducer(NoMatchingCase, 'switch')
                    err.throw('cannot find a matching case')
                }
            }
        ),
        FunctionObject.create (
            'local switch_string (String key, Hash *hash) -> Any',
            function (a) {
                let err = ErrorProducer(NoMatchingCase, 'switch')
                let hash = a.hash.data
                let key = a.key
                err.assert(has(hash, key), `cannot find key '${key}' in hash`)
                return hash[key]
            }
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
            a => FunctionObject.create (
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
     
    union: OverloadObject('union', [
        FunctionObject.create (
            'local union (ConceptList concepts) -> Concept',
            a => ConceptObject.UnionAll(a.concepts.data)
        )
    ]),

    intersect: OverloadObject('intersect', [
        FunctionObject.create (
            'local intersect (ConceptList concepts) -> Concept',
            a => ConceptObject.IntersectAll(a.concepts.data)
        )
    ]),
          
    complement: OverloadObject('complement', [
        FunctionObject.create (
            'local complement (Concept concept) -> Concept',
            a => ConceptObject.Complement(a.concept)
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
    
    raw_type: OverloadObject('raw_type', [
        FunctionObject.create (
            'local raw_type (Any *object) -> String',
            a => transform(a.object, [
                { when_it_is: List, use: x => 'array' },
                { when_it_is: NullObject, use: x => 'null' },
                { when_it_is: Otherwise, use: x => (typeof x) }
            ])
        )
    ]),
    
    type: OverloadObject('type', [
        FunctionObject.create (
            'local type (Any *object) -> String',
            a => GetType(a.object)
        )
    ]),

    Im: OverloadObject('Im', [
        FunctionObject.create (
            'local Im (Cooked *object) -> Immutable',
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
            'local clone (Raw &object) -> Raw',
            a => a.object
        ),
        FunctionObject.create (
            'local clone (Atomic object) -> Atomic',
            a => a.object
        ),
        FunctionObject.create (
            'local clone (Compound object) -> Mutable',
            a => Clone(a.object)
        )
    ]),
     
    raw: OverloadObject('raw', [
        FunctionObject.create (
            'local raw (Functional f) -> RawFunction',
            a => function () {
                return a.f.apply.apply(a.f, map(arguments, x => x))
            }
        ),
        FunctionObject.create (
            'local raw (Compatible *value) -> Compatible',
            a => a.value
        ),
        FunctionObject.create (
            'local raw (MutCompound &compound) -> RawCompound',
            a => RawCompound(a.compound)
        )
    ]),
    
    cook: OverloadObject('cook', [
        FunctionObject.create (
            'local cook (RawCompound &raw_compound) -> MutCompound',
            a => CookCompound(a.raw_compound)
        )
    ])
    
})


pour(K, {
    
    at: OverloadObject('at', [
        FunctionObject.create (
            'local at (RawList &self, Index index) -> Any',
            function (a) {
                let err = ErrorProducer(IndexError, 'RawList::at')
                err.assert(a.index < a.self.length, `${a.index}`)
                return a.self[a.index]
            }
        ),
        FunctionObject.create (
            'local at (ImList self, Index index) -> Immutable',
            a => ImRef(a.self.at(a.index))
        ),
        FunctionObject.create (
            'local at (MutList &self, Index index) -> Any',
            a => a.self.at(a.index)
        )
    ]),
     
    change: OverloadObject('change', [
        FunctionObject.create (
            'local change (RawList &self, Index index, Any *new_value) -> Void',
            function (a) {
                let err = ErrorProducer(IndexError, 'RawList::change')
                err.assert(a.index < a.self.length, `${a.index}`)
                a.self[a.index] = a.new_value
                return VoidObject
            }
        ),
        FunctionObject.create (
            'local change (MutList &self, Index index, Any *new_value) -> Void',
            a => a.self.change(a.index, a.new_value)
        )
    ]),
     
    append: OverloadObject('append', [
        FunctionObject.create (
            'local append (RawList &self, Any *element) -> Void',
            a => (a.self.push(a.element), VoidObject)
        ),
        FunctionObject.create (
            'local append (MutList &self, Any *element) -> Void',
            a => a.self.append(a.element)
        )
    ]),
    
    length: OverloadObject('length', [
        FunctionObject.create (
            'local length (RawList &self) -> Size',
            a => a.self.length
        ),
        FunctionObject.create (
            'local length (List self) -> Size',
            a => a.self.length()
        )
    ])

})


pour(K, {
    
    has: OverloadObject('has', [
        FunctionObject.create (
            'local hash (RawHash &self, String key) -> Bool',
            a => has(a.self, a.key)
        ),
        FunctionObject.create (
            'local has (Hash self, String key) -> Bool',
            a => a.self.has(a.key)
        )
    ]),
    
    get: OverloadObject('get', [
        FunctionObject.create (
            'local get (RawHash &self, String key) -> Any',
            function (a) {
                let err = ErrorProducer(KeyError, 'RawHash::get')
                err.assert(has(a.self, a.key), `${a.key}`)
                return a.self[a.key]
            }
        ),
        FunctionObject.create (
            'local get (ImHash self, String key) -> Immutable',
            a => ImRef(a.self.get(a.key))
        ),
        FunctionObject.create (
            'local get (MutHash &self, String key) -> Any',
            a => a.self.get(a.key)
        )
    ]),
    
    fetch: OverloadObject('fetch', [
        FunctionObject.create (
            'local fetch (RawHash &self, String key) -> Any',
            a => has(a.self, a.key)? a.self[a.key]: NaObject
        ),
        FunctionObject.create (
            'local fetch (ImHash self, String key) -> Immutable',
            a => ImRef(a.self.fetch(a.key))
        ),
        FunctionObject.create (
            'local fetch (MutHash &self, String key) -> Any',
            a => a.self.fetch(a.key)
        )
    ]),
    
    set: OverloadObject('set', [
        FunctionObject.create (
            'local set (RawHash &self, String key, Any *value) -> Void',
            a => (a.self[a.key] = a.value, VoidObject)
        ),
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
