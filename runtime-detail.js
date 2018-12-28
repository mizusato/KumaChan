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
class ForbiddenCall extends RuntimeError {}


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
                throw new err_class(`${f_name}(): ${err_type}: ${err_msg}`)
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
 *  Detail Functions
 */


const Detail = {
    Config: {},
    Hash: {},
    List: {},
    Concept: {},
    Prototype: {},
    Argument: {},
    Function: {},
    Object: {}
}


Detail.Config.get_flags = function (object) {
    check(Detail.Config.get_flags, arguments, {
        object: ObjectObject
    })
    let flag = { frozen: '[Fz]', immutable: '[Im]' }
    let order = ['frozen', 'immutable']
    let has = filter(order, k => object.config[k])
    return (has.length > 0)? flag[has[0]] : ''
}


Detail.Hash.get_prototype = () => ({
    mapper: mapval,
    has: function (key) {
        return Object.prototype.has.call(this.data, key)
    },
    get: function (key) {
        let err = ErrorProducer(KeyError, 'Hash::get')
        err.assert(this.has(key), `'${key}' does not exist`)
        return this.data[key]
    },
    fetch: function (key) {
        return this.has(key) && this.data[key] || NaObject
    },
    set: function (key, value) {
        this.data[key] = value
        return VoidObject
    },
    emplace: function (key, value) {
        let err = ErrorProducer(KeyError, 'Hash::emplace')
        err.if(this.has(key), `'${key}' already exist`)
        this.data[key] = value
        return VoidObject
    },
    replace: function (key, value) {
        let err = ErrorProducer(KeyError, 'Hash::replace')
        err.assert(this.has(key), `'${key}' does not exist`)
        this.data[key] = value
        return VoidObject
    },
    take: function (key) {
        let err = ErrorProducer(KeyError, 'Hash::take')
        err.assert(this.has(key), `'${key}' does not exist`)
        let value = this.data[key]
        delete this.data[key]
        return value
    },
    toString: function () {
        let flags = Config.get_flags(this)
        let list = map(
            this.data,
            (k,v) => `${k}: ${ObjectObject.represent(v)}`
        )
        return `${flags}{${join(list, ', ')}}`
    }
})


Detail.List.get_prototype = () => ({
    mapper: map,
    length: function () { return this.data.length },
    at: function (index) {
        let err = ErrorProducer(IndexError, 'List::at')
        err.assert(index < this.data.length, `${index}`)
        assert(typeof this.data[index] != 'undefined')
        return this.data[index]
    },
    append: function (element) {
        this.data.push(element)
        return VoidObject
    },
    toString: function () {
        let flag = ''
        if ( this.config.frozen ) {
            flag = '[Fz]'
        } else if ( this.config.immutable ) {
            flag = '[Im]'
        }
        let list = map(this.data, x => ObjectObject.represent(x))
        return `${flag}[${join(list, ', ')}]`
    }
})


Detail.Concept.Union = function (concept1, concept2, new_name) {
    check(Detail.Concept.Union, arguments, {
        concept1: ConceptObject,
        concept2: ConceptObject,
        new_name: Optional(Str)
    })
    let name = new_name || `(${concept1.name} | ${concept2.name})`
    let f = (x, info) => exists(
        map_lazy([concept1, concept2], c => c.checker),
        f => f.apply(info.is_immutable? x: ForceMutable(x)) === true
    )
    return ConceptObject(name, f)
}


Detail.Concept.Intersect = function (concept1, concept2, new_name) {
    check(Detail.Concept.Intersect, arguments, {
        concept1: ConceptObject,
        concept2: ConceptObject,
        new_name: Optional(Str)
    })
    let name = new_name || `(${concept1.name} & ${concept2.name})`
    let f = (x, info) => forall(
        map_lazy([concept1, concept2], c => c.checker),
        f => f.apply(info.is_immutable? x: ForceMutable(x)) === true
    )
    return ConceptObject(name, f)
}


Detail.Concept.Complement = function (concept, new_name) {
    check(Detail.Concept.Complement, arguments, {
        concept: ConceptObject,
        new_name: Optional(Str)
    })
    let name = new_name || `!${concept.name}`
    let f = (x, info) => (
        false === concept.checker.apply(
            info.is_immutable? x: ForceMutable(x)
        ) 
    )
    return ConceptObject(name, f)
}


Detail.Prototype.check_argument = function (prototype, argument) {
    check( Detail.Prototype.check_argument, arguments, {
        prototype: Prototype, argument: Hash
    })
    let proto = prototype
    let parameters = proto.parameters
    let order = proto.order
    // argument = { is: 123 } => argument.is() fails
    let has = key => Object.prototype.has.call(argument, key)
    map(argument, arg => assert(ObjectObject.contains(arg)))
    return need (
        cat(
            map_lazy(Object.keys(argument), key => suppose(
                !(key.is(NumStr) && order.has_no(key)),
                `redundant argument ${key}`
            )),
            map_lazy(Object.keys(argument), key => suppose(
                !(key.is(NumStr) && has(order[key])),
                `conflict argument ${key}`
            )),
            map_lazy(order, (key, index) => suppose(
                has(index) || has(key),
                `missing argument ${key}`
            )),
            lazy(function () {
                let arg = mapkey(
                    argument,
                    key => key.is(NumStr)? order[key]: key
                )
                return map_lazy(parameters, (key, p) => suppose(
                    !Object.prototype.has.call(arg, key)
                        || (p.constraint === AnyConcept
                            && ObjectObject.contains(arg[key]))
                        || p.constraint.checker.apply(arg[key]),
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
    return mapkey(
        argument,
        key => key.is(NumStr)? prototype.order[key]: key
    )
}


Detail.Prototype.set_mutability = function (prototype, normalized) {
    return mapval(
        normalized, (val, key) => PassAction[prototype.parameters[key].pass_policy](val)
    )
}


Detail.Prototype.check_return_value = function (prototype, value) {
    check( Detail.Prototype.check_return_value, arguments, {
        prototype: Prototype, value: Any
    })
    return suppose(
        prototype.return_value.checker.apply(value),
        `invalid return value ${value}`
    )
}


Detail.Prototype.represent = function (prototype) {
    check(Detail.Prototype.represent, arguments, { prototype: Prototype })
    function repr_parameter (key, parameter) {
        check(repr_parameter, arguments, { key: Str, parameter: Parameter })
        let type = parameter.constraint.name
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
    ], x => x), ', ') + ') -> ' + `${retval_constraint.name}`
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
        if ( K.has_(string) && K[string].is(ConceptObject) ) {
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


Detail.Argument.represent = function (argument) {
    let list = map(
        argument,
        (k,v) => `${k}=${ObjectObject.represent(v)}`
    )
    return `(${join(list, ', ')})`
}


Detail.Function.create = function (name_and_proto, js_function) {
    check(Detail.Function.create, arguments, {
        name_and_proto: Str, js_function: Function
    })
    let name = name_and_proto.split(' ')[1]
    let prototype = join(
        filter(name_and_proto.split(' '), (s,index) => index != 1), ' '
    )
    return FunctionObject(
        name, G, Prototype.parse(prototype), function (scope) {
            return js_function (scope.data.argument.data)
        }
    )
}


Detail.Function.converge = function (proto_list, f) {
    check(Detail.Function.converge, arguments, {
        proto_list: ArrayOf(Str),
        f: Function
    })
    return map(proto_list, p => Detail.Function.create(p, f))
}


Detail.Object.represent = function (object) {
    check(Detail.Object.represent, arguments, {
        object: ObjectObject
    })
    if ( object.is(StringObject) ) {
        return `"${object}"`
    } else {
        return `${object}`
    }
}
