'use strict';


function Lookup (scope) {
    return (string => scope.lookup(string))
}


function FormatString (string, id_ref) {
    check(FormatString, arguments, { string: Str, id_ref: Fun })
    return string.replace(/\${([^}]+)}/g, (_, arg) => id_ref(arg))
}


function Abstract (checker, name) {
    check(Abstract, arguments, {
        checker: FunctionalObject, name: Optional(Str)
    })
    return ConceptObject(name || '{Temp}', checker)
}


function Structure (hash_object) {
    check(Structure, arguments, { hash_object: HashObject })
    let err = ErrorProducer(InvalidDefinition, 'runtime::structure')
    err.if_failed(need(map_lazy(
        hash_object.data,
        (key, value) => suppose(
            is(value, ConceptObject), `${key} is not Concept`
        )
    )))
    let converted = mapval(
        hash_object.data,
        concept_object => $(object => concept_object.checker(object))
    )
    let struct = Struct(converted)
    let struct_object = $n( HashObject, $(x => is(x.data, struct)) )
    let key_list = join(map(converted, key => key), ', ')
    return PortConcept(struct_object, `struct {${key_list}}`)
}


function Lambda (context, parameter_names, f) {
    check(Lambda, arguments, {
        context: Scope, parameter_names: ListOf(Str), f: Fun
    })
    // if no parameter provided by user, assume a parameter called "__"
    let normalized = (parameter_names.length > 0)? parameter_names: ['__']
    let parameters = map(normalized, name => ({
        name: name,
        pass_policy: 'immutable',
        constraint: AnyConcept
    }))
    let proto = {
        effect_range: 'local',
        parameters: parameters,
        value_constraint: AnyConcept
    }
    return FunctionObject('<Lambda>', context, proto, f)
}


function FunInst (context, effect, parameters, target, f, name = '[Anonymous]') {
    // create a function instance
    let normalized = map(parameters, array => ({
        name: array[0],
        constraint: array[1],
        pass_policy: array[2]
    }))
    let proto = {
        effect_range: effect,
        parameters: normalized,
        value_constraint: target
    }
    return FunctionObject(name, context, proto, f)
}


function define (scope, name, effect, parameters, target, f) {
    let create_at = (
        scope => FunInst(scope, effect, parameters, target, f, name)
    )
    let existing = scope.try_to_lookup(name)
    if ( is(existing, OverloadObject) ) {
        // the new function should have access to the overridden old function
        let wrapper_scope = Scope(scope, effect)
        wrapper_scope.set(name, existing)
        scope.set(name, existing.added(create_at(wrapper_scope)))
    } else {
        scope.emplace(name, OverloadObject(name, [create_at(scope)]))
    }
}


function apply (functional) {
    assert(is(functional, FunctionalObject))
    return function(...args) {
        return functional.apply.apply(functional, args)
    }
}


function call (functional, argument) {
    let e = ErrorProducer(InvalidOperation, 'runtime::call')
    e.assert(is(functional, FunctionalObject), 'calling non-functional')
    assert(is(argument, Hash))
    return functional.call(argument)
}


function get (object, name) {
    let e = ErrorProducer(InvalidOperation, 'runtime::get')
    let f = (is(object, ImmutableObject))? ImRef: (x => x)
    return transform(object, [
        { when_it_is: HashObject,
          use: h => assert(is(name, Str)) && f(h.get(name)) },
        { when_it_is: ListObject,
          use: l => assert(is(name, Num)) && f(l.at(name)) },
        { when_it_is: Otherwise,
          use: x => e.throw(`except Hash or List: ${GetType(object)} given`) }
    ])
}


function set (object, name, value) {
    let e = ErrorProducer(InvalidOperation, 'runtime::set')
    let msg = 'changing element value of immutable compound object'
    e.if(is(object, ImmutableObject), msg)
    transform(object, [
        { when_it_is: HashObject,
          use: h => assert(is(name, Str)) && h.set(name, value) },
        { when_it_is: ListObject,
          use: l => assert(is(name, Num)) && l.change(name, value) },
        { when_it_is: Otherwise,
          use: x => e.throw(`except Hash or List: ${GetType(object)} given`) }
    ])
}


function access (object, name, scope) {
    function wrap (f) {
        check(wrap, arguments, { f: FunctionObject })
        return OverloadObject(f.name, [f])
    }
    function find_on (overload) {
        check(find_on, arguments, { overload: OverloadObject })
        return overload.find_method_for(object)
    }
    let maybe_method = scope.try_to_lookup(name)
    let method = transform(maybe_method, [
        { when_it_is: FunctionObject, use: f => find_on(wrap(f)) },
        { when_it_is: OverloadObject, use: o => find_on(o) },
        { when_it_is: Otherwise, use: x => NotFound }
    ])
    if (method != NotFound) {
        return method
    } else if (is(object, HashObject)) {
        return get(object, name)
    } else {
        let err = ErrorProducer(ObjectNotFound, 'runtime::access')
        let repr = ObjectObject.represent(object)
        err.throw(`unable to find a method called ${name} for ${repr}`)
    }
}


function assign_outer (scope, key, value) {
    let err = ErrorProducer(InvalidAssignment, 'runtime::assign_outer')
    let result = scope.find_name(key)
    err.if(result.scope === scope, `${key} is not an outer variable`)
    err.if(result.scope === K, `global scope is read-only`)
    err.if(is(result, NotFound), `variable ${key} not declared`)
    err.if(result.is_immutable, `the scope containing ${key} is immutable`)
    result.scope.replace(key, value)
}


function assert_bool (value) {
    assert(typeof value == 'boolean')
    return value
}
