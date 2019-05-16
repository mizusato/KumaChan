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
    let { conflict_if, missing_if, invalid_if } = class_error_tools(class_)
    // create empty info: { name -> { method, from: class or interface } }
    let info = {}
    // add own methods
    foreach(class_.methods, (name, method) => {
        info[name] = { method: method, from: class_ }
    })
    // add exposed methods (inherited methods)
    foreach(only_class(class_.impls), super_class => {
        foreach(super_class.methods_info, (name, method_info) => {
            ensure (
                !has(name, info), 'method_conflict',
                name, info[name].from.desc, method_info.from.desc
            )
            info[name] = { method: method_info.method, from: super_class }
        })
    })
    foreach(only_interface(class_.impls), I => {
        // add interface methods (default implementations)
        foreach(I.defaults, (name, method) => {
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
    let all = cat([class_], flat(map(
        only_classes(class_.impls),
        super_class => super_class.super_classes
    )))
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


/**
 *  Class Object
 */
class Class {
    constructor (name, impls, init, methods, data = {}, def_point = null) {
        assert(is(name, Types.String))
        assert(is(impls, TypedList.of(Types.OO_Abstract)))
        assert(is(init, Types.Function))
        assert(is(methods, TypedHash.of(Types.Overload)))
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
        this.methods = copy(methods)
        this.data = copy(data)
        Object.freeze(this.impls)
        Object.freeze(this.methods)
        Object.freeze(this.data)
        this.methods_info = get_methods_info(this)
        this.super_classes = get_super_classes(this)
        this.super_interfaces = get_super_interfaces(this)
        let F = init[WrapperInfo]
        this.create = wrap(
            F.context, F.proto, F.vals, F.desc, scope => {
                let self = new Instance(this, scope, methods)
                let expose = fun (
                    'function expose (internal: Instance) -> Instance',
                    internal => {
                        add_exposed_internal(internal, self)
                        return internal
                    }
                )
                scope.try_to_declare('self', self)
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
    get [Symbol.toStringTag]() {
        return 'Class'
    }
}

let RawMethodTable = TypedList.of(struct({
    name: Types.String,
    f: Types.Function
}))

let check_superset = fun (
    'function check_superset (impls: List) -> Void',
    impls => {
        foreach(impls, (superset, i) => {
            ensure(is(superset, OO_Abstract), 'superset_invalid', i)
        })
        return Void
    }
)

function build_method_table (raw_table) {
    assert(is(raw_table, RawMethodTable))
    let reduced = {}
    foreach(raw_table, item => {
        let { name, f } = item
        if (!has(name, methods)) {
            reduced[name] = [f]
        } else {
            reduced[name].push(f)
        }
    })
    return mapval(reduced, (f_list, name) => overload(f_list, name))
}

function create_class (name, impls, init, raw_methods, data, def_point) {
    call(check_superset, [impls], def_point.file, def_point.row, def_point.col)
    let methods = build_method_table(raw_methods)
    return new Class(name, impls, init, methods, data, def_point)
}


/**
 *  Instance Object
 */
class Instance {
    constructor (class_, scope, methods) {
        this.class_ = class_
        this.scope = scope
        this.exposed = []
        this.methods = mapval(methods, f => bind_context(f, scope))
        foreach(this.methods, (name, method) => {
            this.scope.declare(name, method)
        })
        this.init_finished = false
        this.data = {
            class: class_
        }
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
    let method = caller_scope.lookup(method_name)
    let found = (method !== NotFound && is(method, Types.ES_Function))
    ensure(found, 'method_not_found', method_name)
    // call the method
    return call(method, [object, ...args], file, row, col)
}



let ProtoTable = TypedHash.of(TypedList.of(Prototype))
let RawProtoTable = TypedList.of(struct({
    name: Types.String,
    proto: Prototype
}))

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
            let { f, row, col } = def_point
            this.desc = `interface ${name} at ${f} (row ${row}, column ${col})`
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
        this.data = {
            Impl: $(object => {
                if(!is(object, Types.Class)) { return false }
                return exists(object.get_super_interfaces, I => I === this)
            })
        }
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Interface'
    }
}

function build_proto_table (raw_table) {
    assert(is(raw_table, RawProtoTable))
    let proto_table = {}
    foreach(raw_table, item => {
        if (!has(item.name, proto_table)) {
            proto_table[name] = [item.proto]
        } else {
            proto_table[name].push(item.proto)
        }
    })
}

let validate_interface = fun (
    'function validate_interface (blank: Hash, implemented: Hash) -> Void',
    (blank, implemented) => {
        for (let method of Object.keys(blank)) {
            ensure(!has(method, implement), 'interface_invalid', method)
        }
        return Void
    }
)

function create_interface (name, raw_proto_table, raw_implemented, def_point) {
    // TODO
    let proto_table = build_proto_table(raw_proto_table)
    let implemented = build_method_table(raw_implemented)
    call (
        validate_interface, [name, proto_table, implemented],
        def_point.file, def_point.row, def_point.col
    )
    return new Interface(name, proto_table, implemented, def_point)
}
