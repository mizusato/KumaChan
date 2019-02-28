(function() {
    
    function format (string, table) {
        return string.replace(/{([^}]+)}/g, (matched, p0) => {
            return (typeof table[p0] != 'undefined')? table[p0]: p0
        })
    }
    
    let F = format
    
    let _ = (x => x)  // placeholder for l10n
    
    class RuntimeError extends Error {}
    class NameError extends RuntimeError {}
    class AccessError extends RuntimeError {}
    class CallError extends RuntimeError {}
    
    class ErrorProducer {
        constructor (error, info = '') {
            this.error = error
            this.info = info
        }
        throw (msg) {
            throw new this.error(this.info? (this.info + ': ' + msg): msg)
        }
        assert (value, msg) {
            if (!value) {
                this.throw(msg)
            }
            return value
        }
    }

    function assert (value) {
        if(!value) { throw new RuntimeError('Assertion Error') }
        return value
    }

    function pour (o1, o2) {
        return Object.assign(o1, o2)
    }
    
    function list (iterable) {
        let result = []
        for (let I of iterable) {
            result.push(I)
        }
        return result
    }
    
    function *map (iterable, f) {
        let index = 0
        for (let I of iterable) {
            yield f(I, index)
            index += 1
        }
    }

    function mapkey (object, f) {
        let mapped = {}
        for (let key of Object.keys(object)) {
            let value = object[key]
            mapped[f(key, value)] = value
        }
        return mapped
    }

    function mapval (object, f) {
        let mapped = {}
        for (let key of Object.keys(object)) {
            mapped[key] = f(object[key], key)
        }
        return mapped
    }
    
    function *filter (iterable, f) {
        let index = 0
        for (let I of iterable) {
            if (f(I, index)) {
                yield I
            }
            index += 1
        }
    }

    let NotFound = { tip: 'Object Not Found' }
    
    function find (iterable, f) {
        for (let I of iterable) {
            if (f(I)) {
                return I
            }
        }
        return NotFound
    }
    
    function *iterate (initial, next_of, terminate) {
        let value = initial
        while (!terminate(value)) {
            yield value
            value = next_of(value)
        }
    }
    
    function fold (iterable, initial, f, terminate) {
        let index = 0
        let value = initial
        for (let I of iterable) {
            value = f(I, value, index)
            if (terminate && terminate(value)) {
                break
            }
        }
        return value
    }
    
    function forall (iterable, f) {
        return fold(iterable, true, ((e,v) => v && f(e)), (v => v == false))
    }
    
    function exists (iterable, f) {
        return fold(iterable, false, ((e,v) => v || f(e)), (v => v == true))
    }

    function chain (functions) {
        return ( x => fold(functions, x, (f, v) => f(v)) )
    }

    let Checker = Symbol('Checker')
    let WrapperInfo = Symbol('WrapperInfo')
    let Symbols = { Checker, WrapperInfo }

    function is (value, abstraction) {
        return abstraction[Checker](value)
    }

    function has(key, object) {
        return Object.prototype.hasOwnProperty.call(object, key)
    }
    
    class Concept {
        constructor (checker) {
            assert(typeof checker == 'function')
            this[Checker] = checker
        }
        get [Symbol.toStringTag]() {
            return 'Concept'
        }
    }

    function union (abstracts) {
        assert(forall(abstracts, a => typeof a[Checker] == 'function'))
        return new Concept(x => exists(abstracts, a => a[Checker](x)))
    }
    
    function intersect (abstracts) {
        assert(forall(abstracts, a => typeof a[Checker] == 'function'))
        return new Concept(x => forall(abstracts, a => a[Checker](x)))
    }
    
    function complement (abstraction) {
        assert(typeof abstraction[Checker] == 'function')
        return new Concept(x => !abstraction[Checker](x))
    }

    let $ = (f => new Concept(f))
    let Uni = ((...args) => union(args))
    let Ins = ((...args) => intersect(args))
    let Not = (arg => complement(arg))

    class Category {
        constructor (abstraction, branches) {
            this[Checker] = (
                (abstraction == null)?
                    (union(Object.values(branches)))[Checker]:
                    abstraction[Checker]
            )
            pour(this, branches)
        }
        get [Symbol.toStringTag]() {
            return 'Category'
        }
    }
    
    let category = ((a, b) => new Category(a, b))
    
    let ES = {
        Undefined: $(x => typeof x == 'undefined'),
        Null: $(x => x === null),
        Boolean: $(x => typeof x == 'boolean'),
        Number: $(x => typeof x == 'number'),
        String: $(x => typeof x == 'string'),
        Symbol: $(x => typeof x == 'symbol'),
        Function: $(x => typeof x == 'function'),
        Object: $(x => typeof x == 'object' && x !== null)
    }

    let Type = {
        Undefined: ES.Undefined,
        Null: ES.Null,
        Symbol: ES.Symbol,
        Bool: ES.Boolean,
        Number: category(ES.Number, {
            Int: $(
                x => Number.isInteger(x) && assert(Number.isSafeInteger(x))
            ),
            Safe: $(x => Number.isSafeInteger(x)),
            Finite: $(x => Number.isFinite(x)),
            NaN: $(x => Number.isNaN(x))
        }),
        String: ES.String,
        Function: category(ES.Function, {
            Wrapped: Ins(ES.Function, $(x => has(WrapperInfo, x))),
            Simple: Ins(ES.Function, $(x => !has(WrapperInfo, x)))
        }),
        Abstract: category(
            $(
                x => typeof x == 'object'
                    && typeof x[Checker] == 'function'
                    && Object.getPrototypeOf(x) !== Object.prototype
            ),
            {
                Concept: $(x => x instanceof Concept),
                Category: $(x => x instanceof Category),
                Singleton: $(x => x instanceof Singleton),
                Enum: $(x => x instanceof Enum),
                Schema: $(x => x instanceof Schema)
                // TODO: Capsule
            }
        ),
        Container: category(null, {
            List: category(null, {
                Mutable: $(x => x instanceof Array),
                Immutable: $(
                    x => x instanceof Reference && x.ptr instanceof Array
                )
            }),
            Hash: category(null, {
                Mutable: Ins(ES.Object, $(
                    x => Object.getPrototypeOf(x) === Object.prototype
                )),
                Immutable: $(x => (
                    (x instanceof Reference)
                        && is(x.ptr, ES.Object)
                        && Object.getPrototypeOf(x.ptr) === Object.prototype
                ))
            })
        })
        // TODO: Instance: $(x => x instanceof Instance)
    }
    
    // TODO
    //let NonSolid = Uni(Type.Container, Type.Instance)
    let NonSolid = Type.Container
    let Solid = Not(NonSolid)
    let Immutable = $(x => x instanceof Reference || is(x, Solid))
    let Mutable = $(x => !(x instanceof Reference && is(x, NonSolid)))
    
    class Reference {
        constructor (object) {
            assert(is(object, NonSolid))
            this.ptr = object
        }
    }
    
    let ImRef = (x => is(x, Solid) && x || new Reference(x))

    class Singleton {
        constructor (description) {
            assert(typeof description == 'string')
            this.description = description
            this[Checker] = (x => x === this)
        }
        get [Symbol.toStringTag]() {
            return `Singleton<${this.description}>`
        }
    }

    let Nil = new Singleton('Nil')
    let Void = new Singleton('Void')
    let Done = new Singleton('Done')

    let list_of = (A => Ins(
        Type.Container.List,
        $(l => forall(l, e => is(e, A)))
    ))
    let StringList = list_of(Type.String)

    class Enum {
        constructor (str_list) {
            assert(is(str_list, StringList))
            let item_set = new Set(str_list)
            this[Checker] = (x => item_set.has(x))
            this.item_list = list(map(str_list, x => x))
        }
    }

    let one_of = ((...items) => new Enum(items))

    class Schema {
        constructor (table, requirement = (x => true)) {
            assert(forall(Object.values(table), v => is(v, Type.Abstract)))
            assert(is(requirement, Type.Function))
            this[Checker] = (x => (
                is(x, Type.Container.Hash)
                    && forall(Object.keys(table), k => is(x[k], table[k]))
                    && requirement(x)
            ))
        }
    }

    let struct = (table => new Schema(table))

    let PassPolicy = one_of('immutable', 'natural', 'dirty')
    let EffectRange = one_of('local', 'upper', 'global')

    let Parameter = struct({
        name: Type.String,
        pass_policy: PassPolicy,
        constraint: Type.Abstract
    })

    let ParameterList = list_of(Parameter)

    let Prototype = struct({
        affect: EffectRange,
        value: Type.Abstract,
        parameters: ParameterList
    })

    class Scope {
        constructor (context, affect = 'local', data = {}) {
            assert(context === null || context instanceof Scope)
            assert(is(affect, EffectRange))
            assert(is(data, Type.Container.Hash))
            this.context = context
            this.affect = affect
            this.data = data
        }
        has (variable) {
            return has(variable, this.data)
        }
        declare (variable, initial_value) {
            assert(!this.has(variable))
            this.data[variable] = initial_value
        }
        assign (variable, new_value) {
            assert(this.has(variable))
            this.data[variable] = new_value
        }
        lookup (variable) {
            let info = this.find(variable)
            if (info == NotFound) {
                return NotFound
            } else {
                if (info.is_mutable) {
                    return info.object
                } else {
                    return ImRef(info.object)
                }
            }
        }
        find (variable) {
            let affect = this.affect
            let mutable_depth = 0
            if (affect == 'local') {
                mutable_depth = 0
            } else if (affect == 'upper') {
                mutable_depth = 1
                let upper = this.context
                while (upper != null && upper.affect == 'upper') {
                    mutable_depth += 1
                    upper = upper.context
                }
            } else if (affect == 'global') {
                mutable_depth = Infinity
            }
            let scope_chain = iterate(
                this, scope => scope.context, scope => scope == null
            )
            return find(map(scope_chain, (scope, depth) => ({
                scope: scope,
                depth: depth,
                is_mutable: depth <= mutable_depth,
                object: scope.has(variable)? scope.data[variable]: NotFound
            })), info => info.object != NotFound)
        }
    }
    
    let Global = new Scope(null)
    
    let name_err = new ErrorProducer(NameError)
    let access_err = new ErrorProducer(AccessError)
    
    function var_lookup(scope, name) {
        assert(scope instanceof Scope)
        let value = scope.lookup(name)
        name_err.assert(
            value != NotFound, 
            F(_('variable {name} not found'), {name})
        )
        return value
    }
    
    function var_declare(scope, name, initial_value) {
        assert(scope instanceof Scope)
        name_err.assert(
            !scope.has(name),
            F(_('variable {name} already declared'), {name})
        )
        scope.declare(name, initial_value)
    }
    
    function var_assign(scope, name, new_value) {
        let info = scope.find(name)
        name_err.assert(
            info != NotFound,
            F(_('variable {name} not declared'), {name})
        )
        access_err.assert(
            info.is_mutable,
            F(_('variable {name} belongs to immutable scope'), {name})
        )
        info.scope.assign(name, new_value)
    }
    
    let err_msg_arg_quantity = (
        (r, g) => F(_('{r} arguments required but {g} given'), {r, g})
    )
    let err_msg_invalid_arg = (
        name => F(_('invalid argument {name}'), {name})
    )
    let err_msg_immutable_dirty = (
        name => F(
            _('immutable reference passed as dirty argument {name}'), {name}
        )
    )
    
    function wrap(context, proto, raw_function, info = '') {
        assert(context instanceof Scope)
        assert(is(proto, Prototype))
        assert(is(raw_function, ES.Function))
        assert(is(info, Type.String))
        let err = new ErrorProducer(CallError, info)
        let invoke = function (args, use_context = context) {
            let r = proto.parameters.length
            let g = args.length
            // check if argument quantity correct
            let ok = (r == g)
            err.assert(ok, !ok && err_msg_arg_quantity(r, g))
            // generate finally processed arguments
            let processed = {}
            for (let i=0; i<proto.parameters.length; i++) {
                let parameter = proto.parameters[i]
                let name = parameter.name
                let arg = args[i]
                // check if the argument matches constraint
                let ok = is(arg, parameter.constraint)
                err.assert(ok, !ok && err_msg_invalid_arg(name))
                // cannot pass immutable reference as dirty argument
                ok = !(parameter.pass_policy == 'dirty' && is(arg, Immutable))
                err.assert(ok, !ok && err_msg_immutable_dirty(name))
                // if pass policy is immutable, take a reference
                processed[name] = (
                    (parameter.pass_policy == 'immutable')?
                    ImRef(arg): arg
                )
            }
            // TODO: add frame to call stack (add info for debugging)
            let scope = new Scope(use_context, proto.affect, processed)
            let value = (
                Function.prototype.call.call(raw_function, null, scope)
            )
            // check the return value
            err.assert(is(value, proto.value), _('invalid return value'))
            // TODO: remove frame from call stack
            return value
        }
        let wrapped =((...args) => invoke(args))
        wrapped[WrapperInfo] = { context, proto, info, invoke, raw_function }
        return wrapped
    }
    
    pour(Global.data, {
        Nil: Nil,
        Void: Void,
        Done: Done,
        undefined: undefined,
        Undefined: Type.Undefined,
        null: null,
        Null: Type.Null,
        Symbol: Type.Symbol,
        Bool: Type.Bool,
        Number: Type.Number,
        Int: Type.Number.Int,
        Abstract: Type.Abstract,
        List: Type.Container.List,
        ImList: Type.Container.List.Immutable,
        MutList: Type.Container.List.Mutable,
        Hash: Type.Container.Hash,
        ImHash: Type.Container.Hash.Immutable,
        MutHash: Type.Container.Hash.Mutable,
        Solid: Solid,
        NonSolid: NonSolid,
        Mutable: Mutable,
        Immutable: Immutable
    })
    
    let export_object = {
        is, has, $, Uni, Ins, Not, Type, Symbols, ImRef,
        Global, var_lookup, var_declare, var_assign, wrap
    }
    let export_name = 'KumaChan'
    let global_scope = (typeof window == 'undefined')? global: window
    global_scope[export_name] = export_object
    
})()
