/**
 *  OOP Implementation: Class & Interface
 *
 *  In this language, the term 'object' does not have to mean
 *    an instance of a class, and anything manipulatable is called an 'object'.
 *  An instance of a class is called an 'Instance Object' or 'Instance',
 *    which is just one kind of 'object'.
 *  Note that instance objects in this language are fully encapsulated,
 *    calling public methods is the only way to manipulate internal data of
 *    instance objects.
 */
pour(Types, {
    Class: $(x => x instanceof Class),
    Instance: $(x => x instanceof Instance),
    Interface: $(x => x instanceof Interface),
})

let OO_Abstract = Uni(Types.Class, Types.Interface)
let ProtoTable = TypedHash.of(TypedList.of(Prototype))
let RawTable = TypedList.of(format({
    name: Types.String,
    f: Uni(Types.Function, Prototype)
}))
let only_class = x => filter(x, y => is(y, Types.Class))
let only_interface = x => filter(x, y => is(y, Types.Interface))


function build_method_table (raw_table) {
    assert(is(raw_table, RawTable))
    let reduced = {}
    foreach(raw_table, item => {
        let { name, f } = item
        if (!is(f, Types.Function)) {
            return
        }
        if (!has(name, reduced)) {
            reduced[name] = [f]
        } else {
            reduced[name].push(f)
        }
    })
    return mapval(reduced, (f_list, name) => overload(f_list, name))
}


function build_proto_table (raw_table) {
    assert(is(raw_table, RawTable))
    let proto_table = {}
    foreach(raw_table, item => {
        if (is(item.f, Types.Function)) {
            return
        }
        if (!has(item.name, proto_table)) {
            proto_table[item.name] = [item.f]
        } else {
            proto_table[item.name].push(item.f)
        }
    })
    return proto_table
}


/**
 *  Class Object
 */
class Class {
    constructor (
        name, impls, init, creators, pfs, methods,
        ops = {}, data = {}, def_point = null
    ) {
        // class NAME
        assert(is(name, Types.String))
        // is IMPLS {
        assert(is(impls, TypedList.of(OO_Abstract)))
        // init (...) { ... }
        assert(is(init, Types.Function))
        // create (...) { ... } create (...) { ... } ...
        assert(is(creators, TypedList.of(Types.Function)))
        // private PF1 (...) { ... } private PF2 (...) { ... } ...
        assert(is(pfs, TypedHash.of(Types.Overload)))
        // METHOD1 (...) { ... } METHOD2 (...) { ... } ...
        assert(is(methods, TypedHash.of(Types.Overload)))
        // operator OPERATOR1 (..) { ... } operator OPERATOR2 (..) { ... } ...
        assert(is(ops, TypedHash.of(Types.Function)))
        // data { ... } }
        assert(is(data, Types.Hash))
        this.name = name
        if (def_point !== null) {
            let { file, row, col } = def_point
            this.desc = `class ${name} at ${file} (row ${row}, column ${col})`
        } else {
            this.desc = `class ${name} (built-in)`
        }
        this.init = init
        this.creators = Object.freeze(copy(creators))
        this.impls = Object.freeze(copy(impls))
        this.pfs = Object.freeze(copy(pfs))
        this.methods = Object.freeze(copy(methods))
        this.ops = Object.freeze(copy(ops))
        this.data = Object.freeze(copy(data))
        this.methods_info = get_methods_info(this)
        this.operators_info = get_operators_info(this)
        this.super_classes = get_super_classes(this)
        this.super_interfaces = get_super_interfaces(this)
        this.create = get_integrated_constructor(this)
        this[Checker] = (object => {
            if (!is(object, Types.Instance)) { return false }
            return exists(object.class_.super_classes, S => S === this)
        })
        Object.freeze(this)
    }
    defined_operator (name) {
        return has(name, this.operators_info)
    }
    get_operator (name) {
        assert(this.defined_operator(name))
        return this.operators_info[name].f
    }
    has (key) {
        return has(key, this.data)
    }
    get (key) {
        assert(this.has(key))
        return this.data[key]
    }
    get [Symbol.toStringTag]() {
        return 'Class'
    }
}


function create_class (
    name, impls, init, raw_pfs, raw_methods,
    options, def_point
) {
    let [ init_main, creators ] = init
    let { ops, data } = options
    check_impls(impls)
    let pfs = build_method_table(raw_pfs)
    let methods = build_method_table(raw_methods)
    return new Class (
        name, impls, init_main, creators, pfs, methods,
        ops, data, def_point
    )
}


function check_impls (impls) {
    assert(is(impls, Types.List))
    foreach(impls, (superset, i) => {
        ensure(is(superset, OO_Abstract), 'superset_invalid', i)
    })
}


/**
 *  Instance Object
 */
class Instance {
    constructor (class_, scope) {
        assert(is(class_, Types.Class))
        assert(scope instanceof Scope)
        let mounted_classes = new Set()
        let methods = mapval(class_.methods, f => bind_context(f, scope))
        let init_finished = false
        this.iterate_methods = f => {
            foreach(methods, (name, method) => f(name, method))
        }
        this.has_method = name => {
            return has(name, methods)
        }
        this.get_method = name => {
            assert(has(name, methods))
            return methods[name]
        }
        this.mount = another => {
            assert(!init_finished)
            ensure(is(another, Types.Instance), 'mounting_non_instance')
            let declared = exists(class_.impls, I => I === another.class_)
            ensure(declared, 'mounting_undeclared', another.class_.desc)
            let not_mounted = !mounted_classes.has(another.class_)
            ensure(not_mounted, 'mounting_mounted')
            another.iterate_methods((name, method) => {
                assert(!has(name, methods))
                methods[name] = method
            })
            mounted_classes.add(another.class_)
        }
        this.finish_init = another => {
            assert(!init_finished)
            foreach(only_class(class_.impls), C => {
                ensure(mounted_classes.has(C), 'not_mounting', C)
            })
            foreach(only_interface(class_.impls), I => {
                apply_implemented(I, this, (name, method) => {
                    assert(is(name, Types.String))
                    assert(is(method, Types.Binding))
                    assert(!has(name, methods))
                    methods[name] = method
                })
            })
            init_finished = true
        }
        this.class_ = class_
        this.scope = scope
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Instance'
    }
}


/**
 *  Creates a wrapped initializer that returns an instance of the given class
 *
 *  @param class_ Class
 *  @return Function
 */
function wrap_initializer (class_) {
    let init = class_.init[WrapperInfo]
    return wrap(init.context, init.proto, init.desc, scope => {
        // create an instance object using the scope of initializer
        let self = new Instance(class_, scope)
        // inject private functions
        foreach(class_.pfs, (name, pf) => {
            scope.try_to_declare(name, bind_context(pf, scope))
        })
        // create mount() for this instance
        let mount = another => {
            self.mount(another)
            return another
        }
        inject_desc(mount, 'mount')
        // invoke the initializer
        init.raw(scope, mount)
        // do some necessary work
        self.finish_init()
        // inject self reference
        scope.try_to_declare('self', self)
        // return the initialized instance
        return self
    })
}


/**
 *  Integrate the main initializer and alternative creators of a class
 *
 *  @param class_ Class
 *  @return Overload
 */
function get_integrated_constructor (class_) {
    let init = wrap_initializer(class_)
    let creators = class_.creators.map(creator => {
        creator = creator[WrapperInfo]
        return wrap(creator.context, creator.proto, creator.desc, scope => {
            let created = creator.raw(scope)
            ensure(is(created, class_), 'creator_returned_invalid')
            return created
        })
    })
    return overload([...creators, init], class_.name)
}


/**
 *  Collects all methods of the given class and performs conflict check
 *
 *  @param class_ Class
 *  @return object
 */
function get_methods_info (class_) {
    assert(is(class_, Types.Class))
    // info format: { NAME -> { method: Binding, from: OO_Abstract } }
    let info = {}
    // add own methods
    foreach(class_.methods, (name, method) => {
        info[name] = { method: method, from: class_ }
    })
    // add mounted methods
    foreach(only_class(class_.impls), super_class => {
        foreach(super_class.methods_info, (name, method_info) => {
            let ok = !has(name, info)
            ensure (
                ok, 'method_conflict',
                name, ok || info[name].from.desc, super_class.desc
            )
            info[name] = copy(method_info)
        })
    })
    foreach(only_interface(class_.impls), I => {
        // add interface (pre-implemented) methods
        foreach(I.implemented, (name, method) => {
            if (!has(name, info)) {
                // if there is no existing method with this name, apply default
                info[name] = { method: method, from: I }
            } else {
                // if such a method exists, it cannot be from another interface
                let from_class = is(info[name].from, Types.Class)
                ensure (
                    from_class, 'method_conflict',
                    name, info[name].from.desc, I.desc
                )
            }
        })
        // check if implements the interface I
        foreach(I.proto_table, (name, protos) => {
            ensure (
                has(name, info), 'method_missing',
                name, class_.desc, I.desc
            )
            ensure (
                match_protos(info[name].method, protos), 'method_invalid',
                name, class_.desc, I.desc
            )
        })
    })
    // output the final info
    Object.freeze(info)
    return info
}


/**
 *  Collects all operators defined on the class and performs conflict check
 *
 *  @param class_ Class
 *  @return object
 */
function get_operators_info (class_) {
    // info format: { NAME -> { f: Function, from: Class } }
    let info = {}
    foreach(class_.ops, (op, f) => {
        info[op] = { f, from: class_ }
    })
    foreach(only_class(class_.impls), super_class => {
        foreach(super_class.operators_info, (op, super_info) => {
            let ok = !has(op, info)
            ensure (
                ok, 'operator_conflict',
                op, ok || info[op].from.desc, super_class.desc
            )
            info[op] = copy(super_info)
        })
    })
    Object.freeze(info)
    return info
}


/**
 *  Tries to get a common operator `op` defined on both `a` and `b`
 *
 *  @param a Instance
 *  @param b Instance
 *  @param op string
 *  @return Function
 */
function get_common_operator (a, b, op) {
    assert(is(a, Types.Instance))
    assert(is(b, Types.Instance))
    assert(is(op, Types.String))
    let A = a.class_
    let B = b.class_
    let X = A.operators_info[op].from
    let Y = B.operators_info[op].from
    ensure(X === Y, 'no_common_class', op)
    return A.operators_info[op].f
}


/**
 *  Collects all [ S ∈ Class | C ⊂ S ] in which C is the argument class_
 *
 *  @param class_ Class
 *  @return array of Class
 */
function get_super_classes (class_) {
    let all = list(cat([class_], flat(map(
        only_class(class_.impls),
        super_class => super_class.super_classes
    ))))
    Object.freeze(all)
    return all
}


/**
 *  Collects all [ I ∈ Interface | C ⊂ I ] in which C is the argument class_
 *
 *  @param class_ Class
 *  @return array of Interface
 */
function get_super_interfaces (class_) {
    let all = list(flat(map(
        class_.impls,
        S => is(S, Types.Class)? S.super_interfaces: [S]
    )))
    Object.freeze(all)
    return all
}


/**
 *  Interface Object
 */
class Interface {
    constructor (name, proto_table, implemented = {}, def_point = null) {
        assert(is(name, Types.String))
        assert(is(proto_table, ProtoTable))
        assert(is(implemented, TypedHash.of(Types.Overload)))
        this.name = name
        if (def_point !== null) {
            let { file, row, col } = def_point
            let pos = `${file} (row ${row}, column ${col})`
            this.desc = `interface ${name} at ${pos}`
        } else {
            this.desc = `interface ${name} (built-in)`
        }
        this.proto_table = mapval(proto_table, l => Object.freeze(copy(l)))
        this.implemented = copy(implemented)
        Object.freeze(this.proto_table)
        Object.freeze(this.implemented)
        this[Checker] = (object => {
            if (!is(object, Types.Instance)) { return false }
            return exists(object.class_.super_interfaces, I => I === this)
        })
        this.Impl = $(object => {
            if(!is(object, Types.Class)) { return false }
            return exists(object.super_interfaces, I => I === this)
        })
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Interface'
    }
}


function create_interface (name, raw_table, def_point) {
    let proto_table = build_proto_table(raw_table)
    let implemented = build_method_table(raw_table)
    validate_interface(proto_table, implemented)
    return new Interface(name, proto_table, implemented, def_point)
}


function validate_interface (blank, implemented) {
    for (let method of Object.keys(blank)) {
        ensure(!has(method, implemented), 'interface_invalid', method)
    }
}


/**
 *  Adds implemented methods of an interface to the given instance
 *
 *  @param interface_ Interface
 *  @param instance Instance
 *  @param add_method function
 */
function apply_implemented (interface_, instance, add_method) {
    let implemented = interface_.implemented
    let keys = Object.keys(implemented)
    if (keys.length == 0) { return }
    let sample = implemented[keys[0]]
    let context = sample[WrapperInfo].functions[0][WrapperInfo].context
    // create the context scope for implemented methods
    let interface_scope = new Scope(context)
    // add blank methods to the scope
    foreach(interface_.proto_table, (name, _) => {
        assert(instance.has_method(name))
        interface_scope.declare(name, instance.get_method(name))
    })
    // for each implemented method
    foreach(implemented, (name, method) => {
        assert(!has(name, interface_.proto_table))
        // create a binding
        let binding = bind_context(method, interface_scope)
        // add to the context scope
        interface_scope.declare(name, binding)
        // add to the instance
        add_method([name], binding)
    })
}


/**
 *  Checks if two function prototypes are equivalent
 *
 *  @param proto1 Prototype
 *  @param proto2 Prototype
 *  @return boolean
 */
function proto_equal (proto1, proto2) {
    assert(is(proto1, Prototype))
    assert(is(proto2, Prototype))
    if (proto1.parameters.length != proto2.parameters.length) {
        return false
    }
    if (!type_equivalent(proto1.value_type, proto2.value_type)) {
        return false
    }
    return equal(proto1.parameters, proto2.parameters, (p1, p2) => {
        return type_equivalent(p1.type, p2.type)
    })
}


/**
 *  Checks if a method matches the given function prototypes
 *
 *  @param method Wrapped
 *  @param protos array of Prototype
 */
function match_protos (method, protos) {
    assert(is(method, Types.Wrapped))
    assert(is(protos, TypedList.of(Prototype)))
    assert(protos.length > 0)
    method = cancel_binding(method)
    let method_protos = []
    if (is(method, Types.Function)) {
        method_protos = [method[WrapperInfo].proto]
    } else {
        assert(is(method, Types.Overload))
        method_protos = method[WrapperInfo].functions.map(
            f => f[WrapperInfo].proto
        )
    }
    // ∃ p ∈ method_protos, ∃ q ∈ protos, such that p is equivalent to q
    return exists(method_protos, p => exists(protos, q => proto_equal(p, q)))
}


/**
 *  Calls a method through OO or UFCS
 *
 *  @param caller_scope Scope | null
 *  @param object any
 *  @param method_name string
 *  @param args array
 *  @param file string
 *  @param row integer
 *  @param col integer
 *  @return any
 */
function call_method (
    caller_scope, object, method_name, args, file = null, row = -1, col = -1
) {
    assert(caller_scope instanceof Scope || caller_scope === null)
    assert(is(method_name, Types.String))
    assert(is(args, Types.List))
    // OO: find the method on the instance object
    if (is(object, Types.Instance)) {
        if (object.has_method(method_name)) {
            return call(object.get_method(method_name), args, file, row, col)
        }
    }
    if (is(object, Types.ES_Object)) {
        let method = object[method_name]
        if (is(method, ES.Function)) {
            return call(method.bind(object), args, file, row, col)
        }
    }
    // UFCS: find the method in the caller scope
    ensure(caller_scope !== null, 'method_not_found', method_name)
    let result = caller_scope.find_function(method_name)
    if (result.value === NotFound || !result.ok) {
        let point_desc = `${file} (row ${row}, column ${col})`
        if (result.value === NotFound) {
            ensure(false, 'method_not_found', method_name, point_desc)
        }
        if (!result.ok) {
            ensure(false, 'variable_inconsistent', method_name, point_desc)
        }
    }
    let method = result.value
    assert(is(method, ES.Function))
    return call(method, [object, ...args], file, row, col)
}
