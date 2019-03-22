'use strict';


(function() {

    /**
     *  Global Scope (Uninitialized)
     */

    let Global = null

    /**
     *  Error Messages
     */

    let MSG = {
        schema_invalid_default: f => `invalid default value for field ${f}`,
        variable_not_found: name => `variable ${name} not found`,
        variable_declared: name => `variable ${name} already declared`,
        variable_not_declared: name => `variable ${name} not declared`,
        variable_const: name => `variable ${name} is not re-assignable`,
        variable_immutable: name => `outer variable ${name} is immutable`,
        arg_wrong_quantity: (r, g) => `${r} arguments required but ${g} given`,
        arg_invalid: name => `invalid argument ${name}`,
        arg_immutable: name => `immutable value for dirty argument ${name}`,
        retval_invalid: 'invalid return value',
        no_matching_function: 'invalid arguments: no matching function',
        method_conflict: (A1, name, A2) => (
            `exposed method conflict: ${A1} and ${A2} both have method ${name}`
        ),
        method_missing: (name, C, I) => (
            `class ${C} does not implement ${I} (missing method ${name})`
        ),
        method_invalid: (name, C, I) => (
            `class ${C} does not implement ${I} (invalid method ${name})`
        ),
        exposing_non_instance: 'unable to expose non-instance object',
        not_exposing: C => `created instance does not expose instance of ${C}`,
    }

    /**
     *  Error Definition & Handling
     */

    class RuntimeError extends Error {}
    class SchemaError extends RuntimeError {}
    class NameError extends RuntimeError {}
    class AssignError extends RuntimeError {}
    class AccessError extends RuntimeError {}
    class CallError extends RuntimeError {}
    class ClassError extends RuntimeError {}
    class InitError extends RuntimeError {}
    

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

    function run (iterable) {
        for (let I of iterable) {}
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

    function foreach (something, f) {
        if (typeof something[Symbol.iterator] == 'function') {
            return run(map(something, f))
        } else {
            return run(mapkv(something, f))
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

    function *flat (iterable_of_iterable) {
        for (let iterable of iterable_of_iterable) {
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
            index += 1
            if (terminate && terminate(value)) {
                break
            }
        }
        return value
    }

    function forall (iterable, f) {
        // ∀ I ∈ iterable, f(I) == true
        return fold(
            iterable, true, ((e,v,i) => v && f(e,i)), (v => v == false)
        )
    }

    function exists (iterable, f) {
        // ∃ I ∈ iterable, f(I) == true
        return fold(
            iterable, false, ((e,v,i) => v || f(e,i)), (v => v == true)
        )
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
    let BranchInfo = Symbol('BranchInfo')
    let Symbols = { Checker, WrapperInfo, BranchInfo }

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
        assert(typeof key == 'string' || typeof key == 'symbol')
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
            Object.freeze(this)
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
            let concept = Ins(precondition, union(Object.values(branches)))
            this[Checker] = concept[Checker]
            pour(this, mapval(branches, A => {
                if (A instanceof Category) {
                    return new Category(
                        Ins(precondition, A[BranchInfo].precondition),
                        A[BranchInfo].branches
                    )
                } else {
                    return Ins(precondition, A)
                }
            }))
            this[BranchInfo] = Object.freeze({ precondition, branches })
            Object.freeze(this)
        }
        get [Symbol.toStringTag]() {
            return 'Category'
        }
        static get_branch (object, category) {
            assert(is(object, category[BranchInfo].precondition))
            let branches = category[BranchInfo].branches
            for (let b of Object.keys(branches)) {
                if (is(object, branches[b])) {
                    if (branches[b] instanceof Category) {
                        return Array.concat(
                            [b],
                            Category.get_branch(object, branches[b])
                        )
                    } else {
                        return [b]
                    }
                }
            }
            assert(false)
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
        Number: ES.Number,
        String: ES.String,
        Function: category(ES.Function, {
            Bare: $(f => !has(WrapperInfo, f)),
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
                Schema: $(x => x instanceof Schema),
                Class: $(x => x instanceof Class),
                Signature: $(x => x instanceof Signature),
                Interface: $(x => x instanceof Interface)
            }
        ),
        Container: category(ES.Object, {
            List: $(x => x instanceof Array),
            Hash: $(x => Object.getPrototypeOf(x) === Object.prototype)
        }),
        Instance: $(x => x instanceof Instance)
    }

    let UserlandTypeRename = {
        Sole: 'Function',
        Binding: 'Function'
    }

    function get_type(object) {
        let type = 'Raw'
        for (let T of Object.keys(Type)) {
            if (is(object, Type[T])) {
                if (Type[T] instanceof Category) {
                    let l = Category.get_branch(object, Type[T])
                    type = l[l.length-1]
                    break
                } else {
                    type = T
                    break
                }
            }
        }
        return UserlandTypeRename[type] || type
    }

    /**
     *  Singleton Object
     *
     *  The so-called "singleton object" is just a kind of abstraction
     *    in this language. If S is a singleton object, it means that
     *    S = { x | x === S }, i.e. (x ∈ S) if and only if (x === S).
     *  The singleton object mechanics is used to create special values,
     *    such as Nil and Void, which are available by default.
     */

    class Singleton {
        constructor (description) {
            assert(typeof description == 'string')
            this.description = description
            this[Checker] = (x => x === this)
            Object.freeze(this)
        }
        get [Symbol.toStringTag]() {
            return `Singleton<${this.description}>`
        }
    }

    let Nil = new Singleton('Nil')
    let Void = new Singleton('Void')

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
            this.item_list = Object.freeze(list(map(str_list, x => x)))
            Object.freeze(this)
        }
        get [Symbol.toStringTag]() {
            return 'Enum'
        }
    }

    let list_of = (A => Ins(
        Type.Container.List,
        $(l => forall(l, e => is(e, A)))
    ))
    let StringList = list_of(Type.String)

    let hash_of = (A => Ins(
        Type.Container.Hash,
        $(h => forall(Object.keys(h), k => is(h[k], A)))
    ))

    let one_of = ((...items) => new Enum(items))

    /**
     *  Schema Object
     *
     *  A schema is an abstraction of Hash Objects with specified structure.
     */

    class Schema {
        constructor (table, defaults = null, requirement = (x => true)) {
            assert(forall(Object.values(table), v => is(v, Type.Abstract)))
            assert(is(defaults, Uni(Type.Null, Type.Container.Hash)))
            assert(is(requirement, Type.Function))
            this.table = table
            this.requirement = requirement
            if (defaults != null) {
                this.defaults = Object.freeze(defaults)
                let err = new ErrorProducer(SchemaError)
                for (let key of Object.keys(this.defaults)) {
                    assert(has(key, this.table))
                    err.assert(
                        is(this.defaults[key], this.table[key]),
                        MSG.schema_invalid_default(key)
                    )
                }
            }
            this[Checker] = (x => {
                if(!is(x, Type.Container.Hash)) { return false }
                if (this.defaults == null) {
                    return (forall(
                        Object.keys(this.table),
                        k => is(x[k], this.table[k])
                    ) && this.requirement(x))
                } else {
                    for (let key of Object.keys(this.table)) {
                        if (has(key, x)) {
                            if(!is(x[key], this.table[key])) {
                                return false
                            }
                        } else if(has(key, this.defaults)) {
                            x[key] = this.defaults[key]
                        } else {
                            return false
                        }
                    }
                    return this.requirement(x)
                }
            })
            Object.freeze(this)
        }
        get [Symbol.toStringTag]() {
            return 'Schema'
        }
    }

    let struct = ((table, def, req) => new Schema(table, def, req))

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

    let PassFlag = { natural: '*', dirty: '&', immutable: '' }
    let FlagValue = { '*': 'natural', '&': 'dirty', '': 'immutable' }

    function parse_decl (string) {
        let match = string.match(/([^ ]+) ([^\( ]+) *\(([^\)]*)\) -> (.+)/)
        let [_, affect, name, params_str, value_str] = match
        let has_p = params_str.trim().length > 0
        let parameters = has_p? (list(map(params_str.split(','), para_str => {
            para_str = para_str.trim()
            let match = para_str.match(/([^ ]+) (\*|\&)?(.+)/)
            let [_, type_str, policy_str, name] = match
            let constraint = Global.lookup(type_str)
            policy_str = policy_str || ''
            let pass_policy = FlagValue[policy_str]
            return { name, constraint, pass_policy }
        }))): []
        let value = Global.lookup(value_str)
        let proto = { affect, parameters, value }
        assert(is(proto, Prototype))
        return { name, proto }
    }

    /*
    function represent_decl (proto_desc, desc) {
        let params = join(map(
            proto_desc.parameters,
            p => `${p.constraint} ${PassFlag[p.pass_policy]}${p.name}`
        ), ', ')
        return (
            `${proto_desc.affect} (${params}) -> ${proto_desc.value} (${desc})`
        )
    }
    */

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
            Object.freeze(this)
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
        try_to_lookup (variable) {
            assert(typeof variable == 'string')
            return (this.has(variable))? variable: null
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
            return find(map(scope_chain, (scope, depth) => {
                let object = (
                    scope.has(variable)? scope.data[variable]: NotFound
                )
                let is_mutable = (
                    depth <= mutable_depth && !scope.check_immutable(object)
                )
                let is_assignable = scope.check_assignable(variable)
                return { scope, depth, object, is_mutable, is_assignable }
            }), info => info.object != NotFound)
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
        name_err.assert(value != NotFound, MSG.variable_not_found(name))
        return value
    }

    function var_declare(scope, name, initial_value) {
        assert(scope instanceof Scope)
        name_err.assert(!scope.has(name), MSG.variable_declared(name))
        scope.declare(name, initial_value)
    }

    function var_assign(scope, name, new_value) {
        let info = scope.find(name)
        name_err.assert(info != NotFound, MSG.variable_not_declared(name))
        assign_err.assert(info.is_assignable, MSG.variable_const(name))
        access_err.assert(info.is_mutable, MSG.variable_immutable(name))
        info.scope.assign(name, new_value)
    }

    /**
     *  Function Wrapper
     */

    function check_args (args, proto, caller_scope, get_err_msg = false) {
        // IMPORTANT: return string, "OK" = valid
        let r = proto.parameters.length
        let g = args.length
        // check if argument quantity correct
        if (r != g) {
            return get_err_msg? MSG.arg_wrong_quantity(r, g): 'NG'
        }
        // check constraints
        for (let i=0; i<proto.parameters.length; i++) {
            let parameter = proto.parameters[i]
            let arg = args[i]
            let name = parameter.name
            // check if the argument matches constraint
            if( !is(arg, parameter.constraint) ) {
                return get_err_msg? MSG.arg_invalid(name): 'NG'
            }
            // cannot pass immutable object as dirty argument
            if (caller_scope != null) {
                let is_dirty = parameter.pass_policy == 'dirty'
                let is_immutable = caller_scope.check_immutable(arg)
                if (is_dirty && is_immutable) {
                    return get_err_msg? MSG.arg_immutable(name): 'NG'
                }
            }
        }
        return 'OK'
    }

    function inject_args (args, proto, scope, caller_scope) {
        for (let i=0; i<proto.parameters.length; i++) {
            let parameter = proto.parameters[i]
            let arg = args[i]
            // if pass policy is immutable, register the argument
            if (parameter.pass_policy == 'immutable') {
                scope.register_immutable(arg)
            } else if (parameter.pass_policy == 'natural') {
                if (caller_scope != null) {
                    let arg_is_immutable = caller_scope.check_immutable(arg)
                    if (arg_is_immutable) {
                        scope.register_immutable(arg)
                    }
                }
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
            inject_args(args, proto, scope, caller_scope)
            if (vals != null) {
                list(mapkv(vals, (k, v) => scope.declare(k, v)))
            }
            // call raw function
            let value = raw(scope)
            // check value
            err.assert(is(value, proto.value), MSG.retval_invalid)
            return value
        }
        // wrap function
        let wrapped = give_arity(
            ((...args) => invoke(args, null)),
            proto.parameters.length
        )
        wrapped[WrapperInfo] = Object.freeze({
            context, invoke, proto, vals, desc, raw
        })
        return wrapped
    }

    let SoleList = list_of(Type.Function.Wrapped.Sole)

    function overload (functions, desc = '') {
        assert(is(functions, SoleList))
        assert(is(desc, Type.String))
        let only1 = (functions.length == 1)
        let invoke = !only1? ((args, caller_scope, use_context = null) => {
            for (let f of rev(functions)) {
                let info = f[WrapperInfo]
                if (check_args(args, info.proto, caller_scope) === 'OK') {
                    return info.invoke(args, caller_scope, use_context)
                }
            }
            let err = new ErrorProducer(CallError, desc)
            err.throw(MSG.no_matching_function)
        }): functions[0][WrapperInfo].invoke
        let o = ((...args) => invoke(args, null))
        functions = Object.freeze(functions)
        o[WrapperInfo] = Object.freeze({ functions, invoke, desc })
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
        f = cancel_binding(f)
        let info = f[WrapperInfo]
        let g = give_arity(
            ((...args) => info.invoke(args, null, context)),
            info.proto? info.proto.parameters.length: 0
        )
        let invoke = function (args, caller_scope, use_context = null) {
            assert(use_context === null)
            return info.invoke(args, caller_scope, context)
        }
        g[WrapperInfo] = { original: f, invoke: invoke }
        return g
    }

    function cancel_binding (f) {
        assert(is(f, Type.Function.Wrapped))
        return f[WrapperInfo].original || f
    }

    function call (f, caller_scope, args) {
        if (is(f, Type.Function.Wrapped)) {
            // TODO: add frame to call stack (add info for debugging)
            // TODO: remove frame from call stack
            return f[WrapperInfo].invoke(args, caller_scope)
        } else {
            return Function.prototype.apply.call(f, null, args)
        }
    }

    function fun (decl_string, body) {
        let parsed = parse_decl(decl_string)
        return wrap(Global, parsed.proto, null, parsed.name, scope => {
            return body.apply(
                null,
                list(cat([scope], map(
                    parsed.proto.parameters,
                    p => scope.lookup(p.name)
                )))
            )
        })
    }

    /**
     *  Class & Instance
     */

    let exp_err = new ErrorProducer(CallError, '::expose()')

    function add_exposed_internal(internal, instance) {
        // expose interface of internal object
        assert(!instance.init_finished)
        exp_err.assert(
            internal instanceof Instance,
            MSG.exposing_non_instance
        )
        instance.exposed.push(internal)
        foreach(internal.methods, (name, method) => {
            assert(!has(name, instance.methods))
            instance.methods[name] = method
        })
    }

    function class_error_tools (class_) {
        let err = new ErrorProducer(ClassError, '::create_class()')
        let msg_conflict = (info1, name, info2) => (
            MSG.method_conflict(info1.from.desc, name, info2.from.desc)
        )
        let conflict_if = ((bool, i1, name, i2) => err.assert(
            !bool, bool && msg_conflict(i1, name, i2)
        ))
        let msg_missing = (name, I) => (
            MSG.method_missing(name, class_.desc, I.desc)
        )
        let missing_if = (bool, name, I) => err.assert(
            !bool, bool && msg_missing(name, I)
        )
        let msg_invalid = (name, I) => (
            MSG.method_invalid(name, class_.desc, I.desc)
        )
        let invalid_if = (bool, name, I) => err.assert(
            !bool, bool && msg_invalid(name, I)
        )
        return { conflict_if, missing_if, invalid_if }
    }

    let only_classes = (x => filter(x, y => y instanceof Class))
    let only_interfaces = (x => filter(x, y => y instanceof Interface))

    function get_methods_info (class_) {
        assert(class_ instanceof Class)
        let { conflict_if, missing_if, invalid_if } = class_error_tools(class_)
        // create empty info: { name -> { method, from: class or interface } }
        let info = {}
        // add own methods
        foreach(class_.methods, (name, method) => {
            info[name] = { method: method, from: class_ }
        })
        // add exposed methods (inherited methods)
        foreach(only_classes(class_.impls), super_class => {
            foreach(super_class.methods_info, (name, method_info) => {
                conflict_if(has(name, info), info[name], name, method_info)
                info[name] = { method: method_info.method, from: super_class }
            })
        })
        foreach(only_interfaces(class_.impls), I => {
            // add interface methods (default implementations)
            foreach(I.defaults, (name, method) => {
                if (!has(name, info)) {
                    info[name] = { method: method, from: I }
                } else {
                    let is_default = (info[name].from instanceof Interface)
                    conflict_if(is_default, info[name], name, { from: I })
                }
            })
            // check if implement the interface I
            foreach(I.sign_table, (name, signature) => {
                missing_if(!has(name, info), name, I)
                invalid_if(!is(info[name].method, signature), name, I)
            })
        })
        // output the final info
        return info
    }

    function get_super_classes (class_) {
        // get all [ S ∈ Class | C ⊂ S ] in which C is the argument class_
        function *_get_super_classes (class_) {
            return cat(
                [class_], flat(map(
                    only_classes(class_.impls), super_class => (
                        _get_super_classes(super_class)
                    )
                ))
            )
        }
        return list(_get_super_classes(class_))
    }

    function get_super_interfaces (class_) {
        // get all [ I ∈ Interface | C ⊂ I ] in which C is the argument class_
        return list(flat(map(
            class_.impls,
            I => (I instanceof Class)? I.super_interfaces: [I]
        )))
    }

    function apply_defaults (interface_, instance) {
        let defaults = interface_.defaults
        if (defaults.length == 0) { return }
        // create the context scope for default implementations
        let interface_scope = new Scope(null)
        // filter methods
        let names = list(mapkv(interface_.sign_table, name => name))
        let f_implemented = (name => has(name, instance.methods))
        let f_not_implemented = (name => !has(name, instance.methods))
        let implemented = list(filter(names, f_implemented))
        let not_implemented = list(filter(names, f_not_implemented))
        // add implemented methods to the context scope
        foreach(implemented, name => {
            interface_scope.declare(name, instance.methods[name])
        })
        // for each default implementation
        foreach(not_implemented, name => {
            assert(has(name, defaults))
            // create a method
            let method = bind_context(defaults[name], interface_scope)
            // add to the context scope
            interface_scope.declare(name, method)
            // add to the instance
            instance.methods[name] = method
        })
    }

    let MethodTable = hash_of(Type.Function.Wrapped)
    let GeneralInterface = Uni(Type.Abstract.Class, Type.Abstract.Interface)
    let GeneralList = list_of(GeneralInterface)

    class Class {
        constructor (impls, init, methods, static_methods, desc) {
            assert(is(impls, GeneralList))
            assert(is(init, Type.Function.Wrapped))
            assert(is(methods, MethodTable))
            assert(is(static_methods, MethodTable))
            assert(is(desc, Type.String))
            this.impls = Object.freeze(impls)
            this.init = cancel_binding(init)
            this.methods = Object.freeze(methods)
            this.static_methods = Object.freeze(static_methods)
            this.desc = desc
            this.methods_info = Object.freeze(get_methods_info(this))
            this.super_classes = Object.freeze(get_super_classes(this))
            this.super_interfaces = Object.freeze(get_super_interfaces(this))
            let F = init[WrapperInfo]
            let err = new ErrorProducer(InitError, desc)
            this.create = wrap(
                F.context, F.proto, F.vals, F.desc, scope => {
                    let self = new Instance(this, scope, methods)
                    let expose = (I => (add_exposed_internal(I, self), I))
                    scope.try_to_declare('self', self, true)
                    scope.try_to_declare('expose', expose, true)
                    F.raw(scope)
                    if (scope.try_to_lookup('self') === self) {
                        scope.unset('self')
                    }
                    if (scope.try_to_lookup('expose') === expose) {
                        scope.unset('expose')
                    }
                    for (let I of impls) {
                        if (I instanceof Class) {
                            err.assert(
                                exists(
                                    self.exposed,
                                    instance => (instance.abstraction === I)
                                ),
                                MSG.not_exposing(I.desc)
                            )
                        } else if (I instanceof Interface) {
                            apply_defaults(I, self)
                        }
                    }
                    self.init_finish()
                    return self
                }
            )
            this[Checker] = (object => {
                return exists(
                    object.abstraction.super_classes,
                    super_class => super_class === this
                )
            })
            Object.freeze(this)
        }
        get [Symbol.toStringTag]() {
            return 'Class'
        }
    }

    function create_class(desc, impls, init, methods, static_methods = {}) {
        return new Class(impls, init, methods, static_methods, desc)
    }

    class Instance {
        constructor (class_object, scope, methods) {
            this.abstraction = class_object
            this.scope = scope
            this.exposed = []
            this.methods = mapval(methods, f => bind_context(f, scope))
            foreach(this.methods, (name, method) => {
                this.scope.declare(name, method)
            })
            this.init_finished = false
        }
        init_finish () {
            this.init_finished = true
            Object.freeze(this.exposed)
            Object.freeze(this.methods)
            Object.freeze(this)
        }
        get [Symbol.toStringTag]() {
            return 'Instance'
        }
    }

    let Input = list_of(Type.Abstract)
    let Output = Type.Abstract

    function check_sole(f, input, output) {
        let proto = f[WrapperInfo].proto
        return (proto.value === output) && (
            proto.parameters.length == input.length
        ) && forall(
            input, (I,i) => proto.parameters[i].constraint === I
        )
    }

    class Signature {
        constructor (input, output) {
            assert(is(input, Input))
            assert(is(output, Output))
            this.input = Object.freeze(input)
            this.output = output
            this[Checker] = (f => {
                if (!is(f, Type.Function.Wrapped)) { return false }
                f = cancel_binding(f)
                if (is(f, Type.Function.Wrapped.Sole)) {
                    return check_sole(f, this.input, this.output)
                } else if (is(f, Type.Function.Wrapped.Overload)) {
                    let functions = f[WrapperInfo].functions
                    return exists(
                        functions,
                        f => check_sole(f, this.input, this.output)
                    )
                }
                assert(false)
            })
            Object.freeze(this)
        }
        get [Symbol.toStringTag]() {
            return 'Signature'
        }
    }

    function sig (input, output) {
        return new Signature(input, output)
    }

    let SignTable = hash_of(Type.Abstract.Signature)

    class Interface {
        constructor (sign_table, defaults = {}, desc = '') {
            assert(is(sign_table, SignTable))
            assert(is(defaults, MethodTable))
            assert(is(desc, Type.String))
            assert(forall(Object.keys(defaults), name => (
                has(name, sign_table) && is(defaults[name], sign_table[name])
            )))
            this.sign_table = Object.freeze(sign_table)
            this.defaults = Object.freeze(defaults)
            this.desc = desc
            this[Checker] = (instance => {
                if (instance instanceof Instance) {
                    return exists(
                        instance.abstraction.super_interfaces,
                        I => (I === this)
                    )
                } else {
                    return false
                }
            })
            Object.freeze(this)
        }
        get [Symbol.toStringTag]() {
            return 'Interface'
        }
    }

    function create_interface (desc, table) {
        let sign_table = mapval(table, v => {
            if (is(v, Type.Function.Wrapped)) {
                let proto = v[WrapperInfo].proto
                return new Signature(
                    list(map(proto.parameters, p => p.constraint)),
                    proto.value
                )
            } else {
                return v
            }
        })
        let defaults = flkv(table, (k,v) => is(v, Type.Function.Wrapped))
        return new Interface(sign_table, defaults, desc)
    }

    /**
     *  Initialize Global Scope
     */

    Global = new Scope(null)
    let G = Global.data

    pour(Global.data, {
        Any: Any,
        Nil: Nil,
        Void: Void,
        undefined: undefined,
        Undefined: Type.Undefined,
        null: null,
        Null: Type.Null,
        Symbol: Type.Symbol,
        Bool: Type.Bool,
        Number: category(Type.Number, {
            Safe: $(x => Number.isSafeInteger(x)),
            Finite: $(x => Number.isFinite(x)),
            NaN: $(x => Number.isNaN(x))
        }),
        Int: Ins(Type.Number, $(
            x => Number.isInteger(x) && assert(Number.isSafeInteger(x))
        )),
        String: Type.String,
        Function: Uni(
            Type.Function.Wrapped.Sole,
            Type.Function.Wrapped.Binding
        ),
        Overload: Type.Function.Overload,
        Abstract: Type.Abstract,
        List: Type.Container.List,
        Hash: Type.Container.Hash,
    })

    /**
     *  Export
     */

    let export_object = {
        is, has, $, Uni, Ins, Not, Type, Symbols, get_type,
        Global, G, var_lookup, var_declare, var_assign,
        wrap, parse_decl, fun, overload, overload_added, overload_concated,
        sig, create_interface, create_class
    }
    let export_name = 'KumaChan'
    let global_scope = (typeof window == 'undefined')? global: window
    global_scope[export_name] = export_object

})()
