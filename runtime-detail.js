'use strict';


/**
 *  Exceptions Definition
 */


class RuntimeError extends Error {}
class RedundantParameter extends RuntimeError {}
class InvalidOperation extends RuntimeError {}
class InvalidArgument extends RuntimeError {}
class InvalidReturnValue extends RuntimeError {}
class InvalidDefinition extends RuntimeError {}
class NoMatchingPattern extends RuntimeError {}
class KeyError extends RuntimeError {}
class IndexError extends RuntimeError {}
class NameConflict extends RuntimeError {}
class ObjectNotFound extends RuntimeError {}
class VariableConflict extends RuntimeError {}
class InvalidAssignment extends RuntimeError {}


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
            if ( is(result, Failed) ) {
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
    Scope: {},
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
    let flag = { immutable: '[Im]' }
    let order = ['immutable']
    let has = filter(order, k => object.config[k])
    return (has.length > 0)? flag[has[0]] : ''
}


Detail.Hash.Prototype = {
    mapper: mapval,
    has: function (key) {
        return has(this.data, key)
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
}


Detail.List.Prototype = {
    mapper: map,
    length: function () { return this.data.length },
    at: function (index) {
        let err = ErrorProducer(IndexError, 'List::at')
        err.assert(index < this.data.length, `${index}`)
        assert(typeof this.data[index] != 'undefined')
        return this.data[index]
    },
    change: function (index, value) {
        let err = ErrorProducer(IndexError, 'List::change')
        err.assert(index < this.data.length, `${index}`)
        this.data[index] = value
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
}


Detail.Scope.Prototype = {
    has: function (key) {
        return has(this.data, key)
    },
    get: function (key) {
        assert(this.has(key))
        return this.data[key]
    },
    set: function (key, value) {
        assert(is(value, ObjectObject))
        this.data[key] = value
    },
    emplace: function (key, value) {
        let err = ErrorProducer(VariableConflict, 'Scope::Emplace')
        err.if(this.has(key), `variable ${key} already declared`)
        this.set(key, value)
    },
    replace: function (key, value) {
        let err = ErrorProducer(InvalidAssignment, 'Scope::Replace')
        err.if(!this.has(key), `variable ${key} not declared`)
        this.set(key, value)
    },
    upward_iterator: function () {
        return iterate(this, x => x.context, NullScope)
    },
    find_name: function (name) {
        let get_scope_chain = (() => this.upward_iterator())
        let range = this.range
        /**
         *  upper_max indicates the most out scope
         *  that can be modified by function whose EffectRange = 'upper'
         *  for example:
         *     0      1      2
         *  [upper, local, global] => upper_max = 1
         *  [upper, upper, local, global] => upper_max = 2
         */
        let upper_max = fold(get_scope_chain(), 0, (scope, index) => (
            scope.range == 'upper' && index + 1 || Break
        ))
        let is_immutable = function (layer) {
            let outside_local = (range == 'local' && layer > 0)
            let outside_upper = (range == 'upper' && layer > upper_max)
            return (outside_local || outside_upper)
        }
        return apply_on(get_scope_chain(), chain(
            x => map_lazy(x, (scope, index) => ({
                is_immutable: is_immutable(index),
                layer: index,
                scope: scope,
                object: scope.has(name)? scope.get(name): null
            })),
            x => find(x, item => item.object != null)
        ))
    },
    lookup: function (name) {
        check(this.__proto__.lookup, arguments, { name: Str })
        let result = this.find_name(name)
        if (result != NotFound) {
            let object = result.object
            let immutable = result.is_immutable
            return (immutable)? ImRef(object): object
        } else {
            ErrorProducer(ObjectNotFound, 'Scope::Lookup').throw(
                `there is no object named '${name}'`
            )
        }
    },
    try_to_lookup: function (name) {
        try {
            return this.lookup(name)
        } catch (ObjectNotFound) {
            return NotFound
        }
    }
}


Detail.Concept.Union = function (concept1, concept2, new_name) {
    check(Detail.Concept.Union, arguments, {
        concept1: ConceptObject,
        concept2: ConceptObject,
        new_name: Optional(Str)
    })
    let name = new_name || `(${concept1.name} | ${concept2.name})`
    let f = ( x => exists([concept1, concept2], c => c.checker(x)) )
    return ConceptObject(name, f)
}


Detail.Concept.Intersect = function (concept1, concept2, new_name) {
    check(Detail.Concept.Intersect, arguments, {
        concept1: ConceptObject,
        concept2: ConceptObject,
        new_name: Optional(Str)
    })
    let name = new_name || `(${concept1.name} & ${concept2.name})`
    let f = ( x => forall([concept1, concept2], c => c.checker(x)) )
    return ConceptObject(name, f)
}


Detail.Concept.Complement = function (concept, new_name) {
    check(Detail.Concept.Complement, arguments, {
        concept: ConceptObject,
        new_name: Optional(Str)
    })
    let name = new_name || `~${concept.name}`
    let f = ( x => !concept.checker(x) )
    return ConceptObject(name, f)
}


Detail.Prototype.get_param_hash = function (prototype) {
    return fold(prototype.parameters, {}, function (parameter, hash) {
        hash[parameter.name] = parameter
        return hash
    })
}


Detail.Prototype.check_argument = function (prototype, argument) {
    check( Detail.Prototype.check_argument, arguments, {
        prototype: Prototype, argument: Hash
    })
    let proto = prototype
    let parameters = proto.parameters
    let hash = Prototype.get_param_hash(proto)
    let arg_has = (key => has(argument, key))
    map(argument, arg => assert(ObjectObject.contains(arg)))
    return need (
        cat(    
            map_lazy(Object.keys(argument), key => suppose(
                !(is(key, NumStr) && !parameters[key])
                && !(is_not(key, NumStr) && !hash[key]),
                `redundant argument ${key}`
            )),
            map_lazy(Object.keys(argument), key => suppose(
                !(is(key, NumStr) && arg_has(parameters[key].name)),
                `conflict argument ${key}`
            )),
            map_lazy(parameters, (parameter, index) => suppose(
                arg_has(index) || arg_has(parameter.name),
                `missing argument ${parameter.name}`
            )),
            lazy(function () {
                let arg = mapkey(
                    argument,
                    key => is(key, NumStr)? parameters[key].name: key
                )
                return map_lazy(hash, (key, param) => suppose(
                        (param.constraint === AnyConcept
                            && ObjectObject.contains(arg[key]))
                        || param.constraint.checker(arg[key]),
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
        key => is(key, NumStr)? prototype.parameters[key].name: key
    )
}


Detail.Prototype.set_mutability = function (prototype, normalized) {
    let h = Prototype.get_param_hash(prototype)
    return mapval(
        normalized, (val, key) => PassAction[h[key].pass_policy](val)
    )
}


Detail.Prototype.check_return_value = function (prototype, value) {
    check( Detail.Prototype.check_return_value, arguments, {
        prototype: Prototype, value: Any
    })
    return suppose(
        prototype.value_constraint.checker(value),
        `invalid return value ${value}`
    )
}


Detail.Prototype.represent = function (prototype) {
    check(Detail.Prototype.represent, arguments, { prototype: Prototype })
    function repr_parameter (parameter) {
        check(repr_parameter, arguments, { parameter: Parameter })
        let type = parameter.constraint.name
        let flags = PassFlag[parameter.pass_policy]
        return `${type} ${flags}${parameter.name}`
    }
    let effect = prototype.effect_range
    let parameters = prototype.parameters
    let retval_constraint = prototype.value_constraint
    return effect + ' ' + '(' + join(filter([
        join(map(parameters,p => repr_parameter(p)), ', ')
    ], x => x), ', ') + ') -> ' + `${retval_constraint.name}`
}


Detail.Prototype.parse = function (string) {
    const pattern = /\((.*)\) -> (.*)/
    check(Detail.Prototype.parse, arguments, { string: Regex(pattern) })
    let str = {
        parameters: string.match(pattern)[1].trim(),
        return_value: string.match(pattern)[2].trim()
    }
    function assert_concept (string) {
        if ( ConceptObject.contains(K[string]) ) {
            return K[string]
        } else {
            throw Error('prototype parsing error: invalid constraint')
        }
    }
    function parse_parameter (string) {
        const pattern = /([^ ]*) *([\*\&]?)(.+)/
        check(parse_parameter, arguments, { string: Regex(pattern) })
        let str = {
            constraint: string.match(pattern)[1].trim(),
            pass_policy: string.match(pattern)[2].trim()
        }
        let name = string.match(pattern)[3]
        return pour({ name: name }, mapval(str, function (s, item) {
            return ({
                constraint: () => assert_concept(s),
                pass_policy: () => PassFlagValue[s]
            })[item]()
        }))
    }
    function safe_split (s) {
        let array = s.split(',')
        return (array.length == 1 && array[0].trim() == '')? []: array
    }
    let trim_all = x => map_lazy(x, y => y.trim())
    return {
        effect_range: string.split(' ')[0],
        parameters: map(
            trim_all(safe_split(str.parameters)),
            s => parse_parameter(s)
        ),
        value_constraint: assert_concept(str.return_value)
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
        name_and_proto: Str, js_function: Fun
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
        proto_list: ListOf(Str),
        f: Fun
    })
    return map(proto_list, p => Detail.Function.create(p, f))
}


Detail.Object.represent = function (object) {
    check(Detail.Object.represent, arguments, {
        object: ObjectObject
    })
    if ( is(object, PrimitiveObject) ) {
        if ( is(object, StringObject) ) {
            return `"${object}"`
        } else {
            return `${object}`
        }
    } else {
        return `<${GetType(object)}>`
    }
}


Detail.Function.MethodFor = function (object) {
    return $n(FunctionObject, $(function (f) {
        let p = f.prototype.parameters
        return (p.length > 0) && (p[0].constraint.checker(object))
    }))
}
