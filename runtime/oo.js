/**
 *  OOP Implementation (Class, Instance, Interface)
 */
let OO_Types = {
    Class: $(x => x instanceof Class),
    Instance: $(x => x instanceof Instance),
    Interface: $(x => x instanceof Interface),
    OO_Abstract: $(x => x instanceof Class || x instanceof Interface)
}
pour(Types, OO_Types)

let only_class = x => filter(x, y => is(y, Types.Class))
let only_interface = x => filter(x, y => is(y, Types.Interface))


/**
 *  External Helper Functions
 */
function add_exposed_internal(internal, instance) {
    // expose interface of internal object
    assert(!instance.init_finished)
    ensure(is(internal, Types.Instance), 'exposing_non_instance')
    instance.exposed.push(internal)
    foreach(internal.methods, (name, method) => {
        assert(!has(name, instance.methods))
        instance.methods[name] = method
    })
}

function get_methods_info (class_) {
    assert(is(class_, Types.Class))
    // create empty info: { name -> { method, from: class or interface } }
    let info = {}
    // add own methods
    foreach(class_.methods, (name, method) => {
        info[name] = { method: method, from: class_ }
    })
    // add exposed methods (inherited methods)
    foreach(only_class(class_.impls), super_class => {
        foreach(super_class.methods_info, (name, method_info) => {
            let ok = !has(name, info)
            ensure (
                ok, 'method_conflict',
                name, ok || info[name].from.desc, method_info.from.desc
            )
            info[name] = { method: method_info.method, from: super_class }
        })
    })
    foreach(only_interface(class_.impls), I => {
        // add interface methods (default implementations)
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
        // check if implement the interface I
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

function get_super_classes (class_) {
    // get all [ S ∈ Class | C ⊂ S ] in which C is the argument class_
    let all = list(cat([class_], flat(map(
        only_class(class_.impls),
        super_class => super_class.super_classes
    ))))
    Object.freeze(all)
    return all
}

function get_super_interfaces (class_) {
    // get all [ I ∈ Interface | C ⊂ I ] in which C is the argument class_
    let all = list(flat(map(
        class_.impls,
        S => is(S, Types.Class)? S.super_interfaces: [S]
    )))
    Object.freeze(all)
    return all
}

function apply_implemented (interface_, instance) {
    // apply implemented methods of interface
    let implemented = interface_.implemented
    if (implemented.length == 0) { return }
    // create the context scope for implemented methods
    let interface_scope = new Scope(null)
    // add blank methods to the interface scope
    foreach(interface_.proto_table, (name, _) => {
        assert(has(name, instance.methods))
        interface_scope.declare(name, instance.methods[name])
    })
    // for each implemented method
    foreach(implemented, (name, method) => {
        assert(!has(name, interface_.proto_table))
        // create a binding
        let binding = bind_context(method, interface_scope)
        // add to the context scope
        interface_scope.declare(name, binding)
        // add to the instance
        instance.methods[name] = binding
    })
}

function proto_equal (proto1, proto2) {
    // do type check before calling this function
    if (proto1.parameters.length !== proto2.parameters.length) {
        return false
    }
    if (proto1.value_type !== proto2.value_type) {
        return false
    }
    return forall (
        proto1.parameters,
        (p, i) => (p.type === proto2.parameters[i].type)
    )
}

function match_protos (method, protos) {
    assert(is(method, Types.Wrapped))
    assert(is(protos, Types.List))
    method = cancel_binding(method)
    let info = []
    if (is(method, Types.Function)) {
        info = [method[WrapperInfo]]
    } else {
        info = method[WrapperInfo].functions.map(f => f[WrapperInfo])
    }
    assert(protos.length > 0)
    return exists(
        info,
        I => exists(
            protos, proto => proto_equal(proto, I.proto)
        )
    )
}

function get_common_class (a, b, op) {
    assert(is(a, Types.Instance))
    assert(is(b, Types.Instance))
    assert(is(op, Types.String))
    let search = (A, B) => {
        if (A === B) {
            // A = B
            if (A.defined_operator(op)) {
                return A
            } else {
                for (let i = 1; i < A.super_classes.length; i += 1) {
                    if (A.super_classes[i].defined_operator(op)) {
                        return A.super_classes[i]
                    }
                }
                return NotFound
            }
        } else {
            let A1 = find(A.super_classes, (S, i) => (i > 0 && S === B))
            if (A1 !== NotFound) {
                // A ⊂ B
                return search(A1, B)
            }
            let B1 = find(B.super_classes, (S, i) => (i > 0 && S === A))
            if (B1 !== NotFound) {
                // B ⊂ A
                return search(A, B1)
            }
            return NotFound
        }
    }
    let result = search(a.class_, b.class_)
    ensure(result !== NotFound, 'no_common_class', op)
    return result
}


/**
 *  Class Object
 */
class Class {
    constructor (
        name, impls, init, pfs, methods,
        ops = {}, data = {}, def_point = null
    ) {
        assert(is(name, Types.String))
        assert(is(impls, TypedList.of(Types.OO_Abstract)))
        assert(is(init, Types.Function))
        assert(is(pfs, TypedHash.of(Types.Overload)))
        assert(is(methods, TypedHash.of(Types.Overload)))
        assert(is(ops, TypedHash.of(Types.Function)))
        assert(is(data, Types.Hash))
        this.name = name
        if (def_point !== null) {
            let { file, row, col } = def_point
            this.desc = `class ${name} at ${file} (row ${row}, column ${col})`
        } else {
            this.desc = `class ${name} at (Built-in)`
        }
        this.init = cancel_binding(init)
        this.impls = copy(impls)
        this.pfs = copy(pfs)
        this.methods = copy(methods)
        this.ops = copy(ops)
        this.data = copy(data)
        Object.freeze(this.impls)
        Object.freeze(this.pfs)
        Object.freeze(this.methods)
        Object.freeze(this.ops)
        Object.freeze(this.data)
        this.methods_info = get_methods_info(this)
        this.super_classes = get_super_classes(this)
        this.super_interfaces = get_super_interfaces(this)
        let F = init[WrapperInfo]
        this.create = wrap(
            F.context, F.proto, null, F.desc, scope => {
                let self = new Instance(this, scope)
                foreach(this.pfs, (name, pf) => {
                    scope.try_to_declare(name, bind_context(pf, scope))
                })
                scope.try_to_declare('self', self)
                let expose = fun (
                    'function expose (internal: Instance) -> Instance',
                    internal => {
                        add_exposed_internal(internal, self)
                        return internal
                    }
                )
                F.raw(scope, expose)
                for (let I of impls) {
                    if (is(I, Types.Class)) {
                        let ok = exists(
                            self.exposed,
                            internal => internal.class_ === I
                        )
                        ensure(ok, 'not_exposing', I.desc)
                    } else if (I instanceof Interface) {
                        apply_implemented(I, self)
                    }
                }
                self.init_finish()
                return self
            }
        )
        this[Checker] = (object => {
            if (!is(object, Types.Instance)) { return false }
            return exists(object.class_.super_classes, S => S === this)
        })
        Object.freeze(this)
    }
    defined_operator (name) {
        return has(name, this.ops)
    }
    get_operator (name) {
        assert(this.defined_operator(name))
        return this.ops[name]
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

let RawTable = TypedList.of(format({
    name: Types.String,
    f: Uni(Types.Function, Prototype)
}))

function check_impls (impls) {
    foreach(impls, (superset, i) => {
        ensure(is(superset, Types.OO_Abstract), 'superset_invalid', i)
    })
}

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

function create_class (
    name, impls, init, raw_pfs, raw_methods,
    options, def_point
) {
    let { ops, data } = options
    check_impls(impls)
    let pfs = build_method_table(raw_pfs)
    let methods = build_method_table(raw_methods)
    return new Class (
        name, impls, init, pfs, methods,
        ops, data, def_point
    )
}


/**
 *  Instance Object
 */
class Instance {
    constructor (class_, scope) {
        this.class_ = class_
        this.scope = scope
        this.exposed = []
        this.methods = mapval(class_.methods, f => bind_context(f, scope))
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


function call_method (
    caller_scope, object, method_name, args, file = null, row = -1, col = -1
) {
    assert(caller_scope instanceof Scope || caller_scope === null)
    assert(is(method_name, Types.String))
    if (is(object, Types.Instance)) {
        // OO: find the method on the instance object
        let method = object.methods[method_name]
        if (method) {
            // if exists, call the method
            return call(method, args, file, row, col)
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
    let method = caller_scope.find(method_name, true)
    ensure(method !== NotFound, 'method_not_found', method_name)
    assert(is(method, ES.Function))
    // call the method
    return call(method, [object, ...args], file, row, col)
}



let ProtoTable = TypedHash.of(TypedList.of(Prototype))


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
            this.desc = `interface ${name} at ${file} (row ${row}, column ${col})`
        } else {
            this.desc = `interface ${name} at (Built-in)`
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

function validate_interface (blank, implemented) {
    for (let method of Object.keys(blank)) {
        ensure(!has(method, implemented), 'interface_invalid', method)
    }
}

function create_interface (name, raw_table, def_point) {
    // TODO
    let proto_table = build_proto_table(raw_table)
    let implemented = build_method_table(raw_table)
    validate_interface(proto_table, implemented)
    return new Interface(name, proto_table, implemented, def_point)
}
