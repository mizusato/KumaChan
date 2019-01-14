'use strict';


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


function FunInst (context, range, parameters, value, f) {
    // create a function instance
    let normalized = map(parameters, array => ({
        name: array[0],
        constraint: array[1],
        pass_policy: array[2]
    }))
    let proto = {
        effect_range: range,
        parameters: normalized,
        value_constraint: value
    }
    return FunctionObject('[Anonymous]', context, proto, f)
}


function Lookup (scope) {
    return (string => scope.lookup(string))
}


function apply (function_object) {
    return function(...args) {
        return function_object.apply.apply(function_object, args)
    }
}


function call (function_object, argument) {
    return function_object.call(argument)
}
