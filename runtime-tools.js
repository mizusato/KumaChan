'use strict';


function Lookup (scope) {
    return (string => scope.lookup(string))
}


function FormatString (string, id_ref) {
    check(FormatString, arguments, { string: Str, id_ref: Function })
    return string.replace(/\${([^}]+)}/g, (_, arg) => id_ref(arg))
}


function Lambda (context, parameter_names, f) {
    check(Lambda, arguments, {
        context: Scope, parameter_names: ArrayOf(Str), f: Function
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
    if ( existing.is(OverloadObject) ) {
        // the new function should have access to the overridden old function
        let wrapper_scope = Scope(scope, effect)
        wrapper_scope.set(name, existing)
        scope.set(name, existing.added(create_at(wrapper_scope)))
    } else {
        scope.emplace(name, OverloadObject(name, [create_at(scope)]))
    }
}


function apply (functional) {
    assert(functional.is(FunctionalObject))
    return function(...args) {
        return functional.apply.apply(functional, args)
    }
}


function call (functional, argument) {
    assert(FunctionalObject.contains(functional))
    assert(Hash.contains(argument))
    return functional.call(argument)
}
