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
    class AssignError extends RuntimeError {}
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
        if(!value) { throw new RuntimeError('Assertion Failed') }
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

    function *rev (list) {
        for (let i=list.length-1; i>=0; i--) {
            yield list[i]
        }
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

    function *mapkv (object, f) {
        for (let key of Object.keys(object)) {
            yield f(key, object[key])
        }
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

    function flkv (object, f) {
        let result = {}
        for (let key of Object.keys(object)) {
            if (f(key, object[key])) {
                result[key] = object[key]
            }
        }
        return result
    }

    function *cat (...iterables) {
        for (let iterable of iterables) {
            for (let I of iterable) {
                yield I
            }
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
    
    function no_repeat (iterable) {
        let s = new Set()
        for (let I of iterable) {
            if (!s.has(I)) {
                s.add(I)
            } else {
                return false
            }
        }
        return true
    }

    let alphabet = 'abcdefghijklmnopqrstuvwxyz'
    
    function give_arity(f, n) {
        // Tool to fix arity of wrapped function
        let para_list = join(filter(alphabet, (e,i) => i < n), ',')
        let g = new Function(para_list, 'return this.apply(null, arguments)')
        return g.bind(f)
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

    function has (key, object) {
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
    let Any = $(x => true)

    /**
     *  Category Object
     *
     *  A collection of abstraction objects can be integrated into
     *    a "category object", which is also an abstraction object.
     */
    
    class Category {
        constructor (precondition, branches) {
            this.concept = Ins(precondition, union(Object.values(branches)))
            this[Checker] = this.concept[Checker]
            pour(this, mapval(branches, A => Ins(precondition, A)) )
        }
        get [Symbol.toStringTag]() {
            return 'Category'
        }
    }
    
    let category = ((a, b) => new Category(a, b))

    /**
     *  ES6 Raw Types Defined by Abstraction Objects
     */

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

    /**
     *  Wrapped Types
     */

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
            Simple: $(f => !has(WrapperInfo, f)),
            Wrapped: category(
                $(f => has(WrapperInfo, f)),
                {
                    Sole: $(f => has('context', f[WrapperInfo])),
                    Overload: $(f => has('functions', f[WrapperInfo])),
                    Binding: $(f => has('original', f[WrapperInfo]))
                }
            )
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
        Container: category(ES.Object, {
            List: $(x => x instanceof Array),
            Hash: $(x => Object.getPrototypeOf(x) === Object.prototype)
        })
        // TODO: Instance: $(x => x instanceof Instance)
    }

    /**
     *  Singleton Object
     *
     *  The so-called "singleton object" is just a kind of abstraction
     *    in this language. If S is a singleton object, it means that
     *    S = { x | x === S }, i.e. (x ∈ S) if and only if (x === S).
     *  The singleton object mechanics is used to create special values,
     *    such as Nil, Void, Done, which are available by default.
     */

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

    /**
     *  Enumeration Object
     *
     *  An enumeration is just a set of string.
     */

    class Enum {
        constructor (str_list) {
            assert(is(str_list, StringList))
            let item_set = new Set(str_list)
            this[Checker] = (x => item_set.has(x))
            this.item_list = list(map(str_list, x => x))
        }
    }

    let list_of = (A => Ins(
        Type.Container.List,
        $(l => forall(l, e => is(e, A)))
    ))
    let StringList = list_of(Type.String)

    let one_of = ((...items) => new Enum(items))

    /**
     *  Schema Object
     *
     *  A schema is an abstraction of Hash Objects with specified structure.
     */

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
     *  Access Control of Function & Scope
     * 
     *  In some functional programming language, functions are restricted
     *    to "pure function", which does not produce side-effect.
     *  But in this language, side-effect is widly permitted, none of
     *    functions are "pure function". Instead of eliminating side-effect,
     *    we decrease side-effect by establishing access control.
     *  If a function never modify an argument, it is possible to
     *    set this argument to be immutable (read-only).
     *  Also, if a function never modify the outer scope, it is possible to
     *    set the outer scope to be immutable (read-only) to the function.
     *  The mechanics described above is like UNIX file permission,
     *    an outer scope or an argument can be set to "rwx" or "r-x".
     *  "Be conservative in what you write, be liberal in what you read."
     */

    /**
     *  Effect Range of Function
     * 
     *  The effect range of a function determines the range of scope chain
     *  that can be affected by the function, which indicates
     *  the magnitude of side-effect.
     *  
     *  |  value  | Local Scope | Upper Scope | Other Scope |
     *  |---------|-------------|-------------|-------------|
     *  |  global | full-access | full-access | full-access |
     *  |  upper  | full-access | full-access |  read-only  |
     *  |  local  | full-access |  read-only  |  read-only  |
     * 
     */

    let EffectRange = one_of('local', 'upper', 'global')
    
    /**
     *  Pass Policy of Parameter
     *  
     *  The pass policy of a parameter determines how the function
     *    process the corresponding argument.
     *  If pass policy is set to immutable, the function will not be able to
     *    modify the argument. (e.g. add element to list)
     * 
     *  |   value   | Immutable Argument |  Mutable Argument  |
     *  |-----------|--------------------|--------------------|
     *  | immutable |    direct pass     | treat as immutable |
     *  |  natural  |    direct pass     |    direct pass     |
     *  |   dirty   |     forbidden      |    direct pass     |
     *
     */

    let PassPolicy = one_of('immutable', 'natural', 'dirty')
    
    /**
     *  Parameter & Function Prototype
     */
    
    let Parameter = struct({
        name: Type.String,
        pass_policy: PassPolicy,
        constraint: Type.Abstract
    })

    let ParameterList = list_of(Parameter)

    let Prototype = Ins(struct({
        affect: EffectRange,
        value: Type.Abstract,
        parameters: ParameterList
    }), $( proto => no_repeat(map(proto.parameters, p => p.name)) ))
    
    /**
     *  Scope Object
     */

    class Scope {
        constructor (context, affect = 'local', data = {}) {
            assert(context === null || context instanceof Scope)
            assert(is(affect, EffectRange))
            assert(is(data, Type.Container.Hash))
            // <context> = upper scope
            this.context = context
            // <affect> = effect range of the corresponding function
            this.affect = affect
            // <data> = Hash { VariableName -> VariableValue }
            this.data = data
            // <assignable> = Set { Non-Constants }
            this.assignable = new Set() 
            // <ACL> = WeakMap { Object -> Immutable? 1: undefined }
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
        check_assignable (variable) {
            assert(typeof variable == 'string')
            return this.assignable.has(variable)
        }
        has (variable) {
            assert(typeof variable == 'string')
            return has(variable, this.data)
        }
        declare (variable, initial_value, is_assignable = false) {
            assert(typeof variable == 'string')
            assert(!this.has(variable))
            this.data[variable] = initial_value
            if (is_assignable) {
                this.assignable.add(variable)
            }
        }
        try_to_declare (variable, initial_value, is_assignable = false) {
            assert(typeof variable == 'string')
            if (!this.has(variable)) {
                this.declare(variable, initial_value, is_assignable)
            }
        }
        assign (variable, new_value) {
            assert(typeof variable == 'string')
            assert(this.has(variable))
            assert(this.assignable.has(variable))
            this.data[variable] = new_value
        }
        force_declare (variable, initial_value) {
            assert(typeof variable == 'string')
            if (this.has(variable)) {
                this.assign(variable, initial_value)
            } else {
                this.declare(variable, initial_value, true)
            }
        }
        unset (variable) {
            assert(typeof variable == 'string')
            assert(this.has(variable))
            delete this.data[variable]
        }
        lookup (variable) {
            assert(typeof variable == 'string')
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
            assert(typeof variable == 'string')
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
                is_assignable: scope.check_assignable(variable),
                object: scope.has(variable)? scope.data[variable]: NotFound
            })), info => info.object != NotFound)
        }
    }
    
    /**
     *  Scope Operation Functions with Error Producer
     */
    
    let name_err = new ErrorProducer(NameError)
    let assign_err = new ErrorProducer(AssignError)
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
        assign_err.assert(
            info.is_assignable,
            F(_('variable {name} is not re-assignable'), {name})
        )
        access_err.assert(
            info.is_mutable,
            F(_('variable {name} belongs to immutable scope'), {name})
        )
        info.scope.assign(name, new_value)
    }
    
    /**
     *  Function Wrapper
     */
    
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

    function check_args (args, proto, caller_scope, get_err_msg = false) {
        // IMPORTANT: return string, "OK" = valid
        let r = proto.parameters.length
        let g = args.length
        // check if argument quantity correct
        if (r != g) {
            return get_err_msg? err_msg_arg_quantity(r, g): 'NG'
        }
        // check constraints
        for (let i=0; i<proto.parameters.length; i++) {
            let parameter = proto.parameters[i]
            let arg = args[i]
            let name = parameter.name
            // check if the argument matches constraint
            if( !is(arg, parameter.constraint) ) {
                return get_err_msg? err_msg_invalid_arg(name): 'NG'
            }
            // cannot pass immutable object as dirty argument
            if (caller_scope != null) {
                let is_dirty = parameter.pass_policy == 'dirty'
                let is_immutable = caller_scope.check_immutable(arg)
                if (is_dirty && is_immutable) {
                    return get_err_msg? err_msg_immutable_dirty(name): 'NG'
                }
            }            
        }
        return 'OK'
    }

    function inject_args (args, proto, scope) {
        for (let i=0; i<proto.parameters.length; i++) {
            let parameter = proto.parameters[i]
            let arg = args[i]
            // if pass policy is immutable, register the argument
            if (parameter.pass_policy == 'immutable') {
                scope.register_immutable(arg)
            }
            // inject argument to scope
            scope.declare(parameter.name, arg)
        }
    }
    
    function wrap (context, proto, vals, desc, raw) {
        assert(context instanceof Scope)
        assert(is(proto, Prototype))
        assert(is(vals, Uni(Type.Null, Type.Container.Hash)))
        assert(is(raw, ES.Function))
        assert(is(desc, Type.String))
        let err = new ErrorProducer(CallError, desc)
        let invoke = (args, caller_scope, use_ctx = null, check = true) => {
            // check arguments
            if (check) {
                let result = check_args(args, proto, caller_scope, true)
                if (result != 'OK') {
                    err.throw(result)
                }
            }
            // generate scope
            let scope = new Scope(
                (use_ctx !== null)? use_ctx: context,
                proto.affect
            )
            inject_args(args, proto, scope)
            if (vals != null) {
                list(mapkv(vals, (k, v) => scope.declare(k, v)))
            }
            // call raw function
            let value = raw(scope, caller_scope)
            // check value
            err.assert(is(value, proto.value), _('invalid return value'))
            return value
        }
        // wrap function
        let wrapped = give_arity(
            ((...args) => invoke(args, null)),
            proto.parameters.length
        )
        // IMPORTANT: the [WrapperInfo] object should not be mutated
        wrapped[WrapperInfo] = { context, invoke, proto, vals, desc, raw }
        return wrapped
    }

    let FunctionList = list_of(Type.Function.Wrapped.Sole)

    function overload (functions) {
        assert(is(functions, FunctionList))
        let invoke = function (args, caller_scope, use_context = null) {
            for (let f of rev(functions)) {
                let info = f[WrapperInfo]
                if (check_args(args, info.proto, caller_scope) === 'OK') {
                    return info.invoke(args, caller_scope, use_context)
                }                
            }
        }
        let o = ((...args) => invoke(args, null))
        o[WrapperInfo] = { functions, invoke }
        return o
    }

    function overload_added (f, o) {
        assert(is(o, Type.Function.Wrapped.Overload))
        return overload([...o[WrapperInfo].functions, f])
    }

    function overload_concated (o2, o1) {
        assert(is(o2, Type.Function.Wrapped.Overload))
        assert(is(o1, Type.Function.Wrapped.Overload))
        return overload(list(
            cat(o1[WrapperInfo].functions, o2[WrapperInfo].functions)
        ))
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

    let NonBinding = Ins(Type.Function.Wrapped, $(
        f => !has('oringinal', f[WrapperInfo])
    ))
    
    function call (f, caller_scope, args) {
        if (is(f, Type.Function.Wrapped)) {
            // TODO: add frame to call stack (add info for debugging)
            // TODO: remove frame from call stack
            return f[WrapperInfo].invoke(args, caller_scope)
        } else {
            return Function.prototype.apply.call(f, null, args)
        }
    }

    /**
     *  Class & Instance
     */

    let exp_err = new ErrorProducer(CallError, 'expose()')

    function add_exposed_internal(internal, instance) {
        // expose interface of internal object
        exp_err.assert(
            internal instanceof Instance,
            _('unable to expose non-instance object')
        )
        instance.exposed.push(internal)
        for (let name of Object.keys(internal.methods)) {
            exp_err.assert(
                !has(name, instance.methods),
                F(_('conflict method name: {name}'), {name})
            )
            instance.methods[name] = internal.methods[name]
        }        
    }

    function is_direct_instance_of (class_, instance) {
        assert(is(class_ instanceof Class))
        assert(is(instance instanceof Instance))
        return (instance.abstraction === class_)
    }

    function is_instance_of (class_, instance) {
        assert(is(class_ instanceof Class))
        assert(is(instance instanceof Instance))
        if (is_direct_instance_of(class_, instance)) {
            return true
        } else {
            return exists (
                instance.exposed,
                E => is_instance_of(E.class_, instance)
            )
        }
    }

    class Class {
        constructor (impls, init, methods, desc) {
            pour(this, { impls, init, methods, desc })
            assert(is(init, NonBinding))
            let I = init[WrapperInfo]
            this.construct = wrap(
                null, I.proto, I.vals, I.desc, (scope, caller_scope) => {
                    let self = new Instance(this, scope, methods)
                    let expose = (I => add_exposed_interal(I, self))
                    scope.try_to_declare('self', self, true)
                    scope.try_to_declare('expose', expose, true)
                    I.raw(scope, caller_scope)
                    // TODO: check if implement all interfaces
                    return self
                }
            )
            this[Checker] = (object => {
                if (object instanceof Instance) {
                    return is_instance_of(this, object)
                } else {
                    return false
                }
            })
        }
    }

    class Instance {
        constructor (class_object, scope, methods) {
            this.abstraction = class_object
            this.scope = scope
            this.exposed = []
            this.methods = mapval(methods, f => bind_context(f, scope))
            for (let name of this.methods) {
                this.scope.declare(name, this.methods[name])
            }
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
