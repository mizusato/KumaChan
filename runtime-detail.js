/* Tookit for Concept and Function */


const Detail = {
    Concept: {},
    Prototype: {},
    Function: {}
}


Detail.Concept.Union = function (concept1, concept2) {
    check(Detail.Concept.Union, arguments, {
        concept1: ConceptObject,
        concept2: ConceptObject
    })
    let name = `(${concept1.name} | ${concept2.name})`
    let f = x => exists(
        map_lazy([concept1, concept2], c => c.checker),
        f => f.apply(x) === true
    )
    return ConceptObject(name, f)
}


Detail.Concept.Intersect = function (concept1, concept2) {
    check(Detail.Concept.Intersect, arguments, {
        concept1: ConceptObject,
        concept2: ConceptObject
    })
    let name = `(${concept1.name} & ${concept2.name})`
    let f = x => forall(
        map_lazy([concept1, concept2], c => c.checker),
        f => f.apply(x) === true
    )
    return ConceptObject(name, f)
}


Detail.Concept.Complement = function (concept) {
    check(Detail.Concept.Complement, arguments, {
        concept: ConceptObject,
    })
    let name = `!${concept.name}`
    let f = x => concept.checker.apply(x) === false
    return ConceptObject(name, f)
}


Detail.Prototype.check_argument = function (prototype, argument) {
    check( Detail.Prototype.check_argument, arguments, {
        prototype: Prototype, argument: Hash
    })
    let proto = prototype
    let parameters = proto.parameters
    let order = proto.order
    return need (
        cat(
            map_lazy(Object.keys(argument), key => suppose(
                !(key.is(NumStr) && order.has_no(key)),
                `redundant argument ${key}`
            )),
            map_lazy(Object.keys(argument), key => suppose(
                !(key.is(NumStr) && argument.has(order[key])),
                `conflict argument ${key}`
            )),
            map_lazy(order, (key, index) => suppose(
                argument.has(index) || argument.has(key),
                `missing argument ${key}`
            )),
            lazy(function () {
                let arg = mapkey(
                    argument,
                    key => key.is(NumStr)? order[key]: key
                )
                return map_lazy(parameters, (key, p) => suppose(
                    !arg.has(key)
                        || (p.constraint === AnyConcept
                            && ObjectObject.contains(arg[key]))
                        || p.constraint.data.checker.apply(arg[key]),
                    `illegal argument '${key}'`
                ))
            })
        )
    )
}


Detail.Prototype.normalize_argument = function (prototype, argument) {
    check( Detail.Prototype.normalize_argument, arguments, {
        prototype: Prototype, argument: Hash
    })
    return mapval(
        mapkey(argument, key => key.is(NumStr)? prototype.order[key]: key),
        (val, key) => PassAction[prototype.parameters[key].pass_policy](val)
    )
}


Detail.Prototype.check_return_value = function (prototype, value) {
    check( Detail.Prototype.check_return_value, arguments, {
        prototype: Prototype, value: Any
    })
    return suppose(
        prototype.return_value.data.checker.apply(value),
        `invalid return value ${value}`
    )
}


Detail.Prototype.represent = function (prototype) {
    check(Detail.Prototype.represent, arguments, { prototype: Prototype })
    function repr_parameter (key, parameter) {
        check(repr_parameter, arguments, { key: Str, parameter: Parameter })
        let type = parameter.constraint.data.name
        let flags = PassFlag[parameter.pass_policy]
        return `${type} ${flags}${key}`
    }
    function opt (string, non_empty_callback) {
        check(opt, arguments, { string: Str, non_empty_callback: Function })
        return string && non_empty_callback(string) || ''
    }
    let effect = prototype.effect_range
    let order = prototype.order
    let parameters = prototype.parameters
    let retval_constraint = prototype.return_value
    let necessary = Enum.apply({}, order)
    return effect + ' ' + '(' + join(filter([
        join(map(
            order,
            key => repr_parameter(key, parameters[key])
        ), ', '),
        opt(join(map(
            filter(parameters, key => key.is_not(necessary)),
            (key, val) => repr_parameter(key, val)
        ), ', '), s => `[${s}]`),
    ], x => x), ', ') + ') -> ' + `${retval_constraint.data.name}`
}


Detail.Prototype.parse = function (string) {
    const pattern = /\((.*)\) -> (.*)/
    const sub_pattern = /(.*), *[(.*)]/
    check(Detail.Prototype.parse, arguments, { string: Regex(pattern) })
    let str = {
        parameters: string.match(pattern)[1].trim(),
        return_value: string.match(pattern)[2].trim()
    }
    let has_optional = str.parameters.match(sub_pattern)
    str.parameters = {
        necessary: has_optional? has_optional[1].trim(): str.parameters,
        all: str.parameters
    }
    function check_concept (string) {
        if ( K.has(string) && K[string].is(ConceptObject) ) {
            return K[string]
        } else {
            throw Error('prototype parsing error: invalid constraint')
        }
    }
    function parse_parameter (string) {
        const pattern = /([^ ]*) *([\*\&\~]?)(.+)/
        check(parse_parameter, arguments, { string: Regex(pattern) })
        let str = {
            constraint: string.match(pattern)[1].trim(),
            pass_policy: string.match(pattern)[2].trim()
        }
        let name = string.match(pattern)[3]
        return { key: name, value: mapval(str, function (s, item) {
            return ({
                constraint: () => check_concept(s),
                pass_policy: () => PassFlagValue[s]
            })[item]()
        }) }
    }
    let trim_all = x => map_lazy(x, y => y.trim())
    return {
        effect_range: string.split(' ')[0],
        order: map(map(
            trim_all(str.parameters.necessary.split(',')),
            s => parse_parameter(s)
        ), p => p.key),
        parameters: fold(map(
            trim_all(str.parameters.all.split(',')),
            s => parse_parameter(s)
        ), {}, (e,v) => (v[e.key]=e.value, v)),
        return_value: check_concept(str.return_value)
    }
}


Detail.Function.create = function (name_and_proto, js_function) {
    check(Detail.Function.create, {
        name_and_proto: Str, js_function: Function
    })
    let name = name_and_proto.split(' ')[1]
    let prototype = join(
        filter(name_and_proto.split(' '), (s,index) => index != 1), ' '
    )
    return FunctionInstanceObject(
        name, G, Prototype.parse(prototype), function (scope) {
            return js_function (scope.data.argument.data)
        }
    )
}


Detail.Function.converge = function (proto_list, f) {
    return map(proto_list, p => Detail.Function.create(p, f))
}
