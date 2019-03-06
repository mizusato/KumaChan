(function() {

    /**
     *  String Format Tool
     */
    
    function format (string, table) {
        // Avoid using ES6 template string for l10n propose.
        return string.replace(/{([^}]+)}/g, (matched, p0) => {
            return (typeof table[p0] != 'undefined')? table[p0]: p0
        })
    }
    
    let F = format
    
    let _ = (x => x)  // placeholder for l10n

    /**
     *  Error Definition & Handling
     */
    
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

    /**
     *  Toolkit Functions
     */

    function pour (o1, o2) {
        return Object.assign(o1, o2)
    }
    
    function list (iterable) {
        // convert iterable to array
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
    
    function join (iterable, separator) {
        let string = ''
        let first = true
        for (let I of iterable) {
            if (first) {
                first = false
            } else {
                string += separator
            }
            string += I
        }
        return string
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
        // apply next_of() on value until terminal condition satisfied
        let value = initial
        while (!terminate(value)) {
            yield value
            value = next_of(value)
        }
    }
    
    function fold (iterable, initial, f, terminate) {
        // reduce() with a terminal condition 
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
        // ∀ I ∈ iterable, f(I) == true
        return fold(iterable, true, ((e,v) => v && f(e)), (v => v == false))
    }
    
    function exists (iterable, f) {
        // ∃ I ∈ iterable, f(I) == true
        return fold(iterable, false, ((e,v) => v || f(e)), (v => v == true))
    }

    function chain (functions) {
        return ( x => fold(functions, x, (f, v) => f(v)) )
    }
    
    function *count (n) {
        let i = 0
        while (i < n) {
            yield i
        }
    }

    /**
     *  Symbol Definition
     */

    let Checker = Symbol('Checker')
    let WrapperInfo = Symbol('WrapperInfo')
    let Symbols = { Checker, WrapperInfo }

    /**
     *  Abstraction Mechanics
     *
     *  An object A that satisfies (typeof A[Checker] == 'function')
     *    is called an "abstraction object" or "abstract set".
     *  The function A[Checker]() should take one argument x
     *    and return a boolean value which indicates whether x ∈ A.
     *  In this programming language, data types are implemented by
     *    abstraction objects. A data type is just an abstraction object,
     *    therefore data type checking is performed at runtime,
     *    by calling the [Checker]() function of the abstraction object.
     */

    function is (value, abstraction) {
        return abstraction[Checker](value)
    }

    function has(key, object) {
        return Object.prototype.hasOwnProperty.call(object, key)
    }

    /**
     *  Concept Object
     *
     *  The so-called "concept objects" are the simplest kind of
     *    abstraction objects, which only contains a [Checker]() function.
     *  In other words, concept objects do not contain any extra information
     *    except its abstraction information.
     *  Any abstraction object can do operations such as intersect,
     *    union, complement with others, and produce a new concept object.
     */
    
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
        // (∪ A_i), for A_i in abstracts
        assert(forall(abstracts, a => typeof a[Checker] == 'function'))
        return new Concept(x => exists(abstracts, a => a[Checker](x)))
    }
    
    function intersect (abstracts) {
        // (∩ A_i), for A_i in abstracts
        assert(forall(abstracts, a => typeof a[Checker] == 'function'))
        return new Concept(x => forall(abstracts, a => a[Checker](x)))
    }
    
    function complement (abstraction) {
        // (∁ A)
        assert(typeof abstraction[Checker] == 'function')
        return new Concept(x => !abstraction[Checker](x))
    }

    let $ = (f => new Concept(f))             // create concept from f(x)
    let Uni = ((...args) => union(args))      // (A,B,...) => A ∪ B ∪ ... 
    let Ins = ((...args) => intersect(args))  // (A,B,...) => A ∩ B ∩ ...
    let Not = (arg => complement(arg))        //  A => ∁ A

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
            List: $(x => x instanceof Array),
            Hash: Ins(ES.Object, $(
                x => Object.getPrototypeOf(x) === Object.prototype
            ))
        })
        // TODO: Instance: $(x => x instanceof Instance)
    }
    
    // TODO
    //let NonSolid = Uni(Type.Container, Type.Instance)
    let NonSolid = Type.Container
    let Solid = Not(NonSolid)

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

    /**
     *  Function & Scope
     */

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
            this.ACL = new WeakMap()
        }
        register_immutable (object) {
            if (typeof object == 'object') {
                this.ACL.set(object, 1)
            }
        }
        check_immutable (object) {
            if (typeof object == 'object') {
                return (this.ACL.get(object) === 1)
            } else {
                return true
            }
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
                if (!info.is_mutable) {
                    this.register_immutable(info.object)
                }
                return info.object
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
    
    let alphabet = 'abcdefghijklmnopqrstuvwxyz'
    
    function give_arity(f, n) {
        let para_list = join(filter(alphabet, (e,i) => i < n), ',')
        let g = new Function(para_list, 'return this.apply(null, arguments)')
        return g.bind(f)
    }
    
    function wrap (context, proto, raw, desc = '') {
        assert(context instanceof Scope)
        assert(is(proto, Prototype))
        assert(is(raw, ES.Function))
        assert(is(desc, Type.String))
        let err = new ErrorProducer(CallError, desc)
        let invoke = function (args, caller_scope, use_context = context) {
            // check if argument quantity correct
            let r = proto.parameters.length
            let g = args.length
            let ok = (r == g)
            err.assert(ok, !ok && err_msg_arg_quantity(r, g))
            // generate scope
            let scope = new Scope(use_context, proto.affect)
            // check arguments
            for (let i=0; i<proto.parameters.length; i++) {
                let parameter = proto.parameters[i]
                let arg = args[i]
                let name = parameter.name
                // check if the argument matches constraint
                let ok = is(arg, parameter.constraint)
                err.assert(ok, !ok && err_msg_invalid_arg(name))
                // cannot pass immutable object as dirty argument
                if (caller_scope != null) {
                    let is_dirty = parameter.pass_policy == 'dirty'
                    let is_immutable = caller_scope.check_immutable(arg)
                    let ok = !(is_dirty && is_immutable)
                    err.assert(ok, !ok && err_msg_immutable_dirty(name))
                }
                // if pass policy is immutable, register it
                if (parameter.pass_policy == 'immutable') {
                    scope.register_immutable(arg)
                }
                // inject argument to scope
                scope.data[name] = arg
            }
            // TODO: add frame to call stack (add info for debugging)
            let value = (
                Function.prototype.call.call(raw, null, scope)
            )
            // check the return value
            err.assert(is(value, proto.value), _('invalid return value'))
            // TODO: remove frame from call stack
            return value
        }
        // wrap function
        let wrapped = give_arity(
            ((...args) => invoke(args, null)),
            proto.parameters.length
        )
        wrapped[WrapperInfo] = { context, invoke, proto, raw, desc }
        return wrapped
    }
    
    function bind_context (f, context) {
        assert(is(f, Type.Function.Wrapped))
        let info = f[WrapperInfo]
        let g = give_arity(
            ((...args) => info.invoke(args, null, context)),
            info.proto.parameters.length
        )
        let invoke = function (args, caller_scope, use_context = null) {
            assert(use_context === null)
            return info.invoke(args, caller_scope, context)
        }
        g[WrapperInfo] = { original: f, invoke: invoke }
        return g
    }
    
    function call (f, context, args) {
        if (is(f, Type.Function.Wrapped)) {
            return f[WrapperInfo].invoke(args, context)
        } else {
            return Function.prototype.apply.call(f, null, args)
        }
    }

    /**
     *  Global Scope
     */
    
    let Global = new Scope(null)
    
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
        Hash: Type.Container.Hash,
        Solid: Solid,
        NonSolid: NonSolid,
    })

    /**
     *  Export
     */
    
    let export_object = {
        is, has, $, Uni, Ins, Not, Type, Symbols,
        Global, var_lookup, var_declare, var_assign, wrap
    }
    let export_name = 'KumaChan'
    let global_scope = (typeof window == 'undefined')? global: window
    global_scope[export_name] = export_object
    
})()
