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
  *  Toolkit Functions
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
        foreach(I.method_table, (name, protos) => {
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
    foreach(interface_.method_table, (name, _) => {
        assert(has(name, instance.methods))
        interface_scope.declare(name, instance.methods[name])
    })
    // for each implemented method
    foreach(implemented, (name, method) => {
        assert(!has(name, interface_.method_table))
        // create a binding
        let binding = bind_context(method, interface_scope)
        // add to the context scope
        interface_scope.declare(name, binding)
        // add to the instance
        instance.methods[name] = binding
    })
}

function match_protos (method, protos) {
    // TODO
    // don't forget to cancel binding
}


/**
 *  Class Object
 */
class Class {
    constructor (name, impls, init, methods, data = {}, def_point = null) {
        assert(is(name, Types.String))
        assert(is(impls, Types.TypedList.of(Types.OO_Abstract)))
        assert(is(init, Types.Function))
        assert(is(methods, Types.TypedHash.of(Types.Overload)))
        assert(is(data, Types.Hash))
        this.name = name
        if (def_point != null) {
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
                    'local expose (Instance *internal) -> Instance',
                    internal => {
                        add_exposed_internal(internal, self)
                        return internal
                    }
                )
                scope.try_to_declare('self', self, true)
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
        this[Solid] = true
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Class'
    }
}

function create_class (name, impls, init, methods, data, def_point) {
    // TODO
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
        ensure(method, 'method_not_found', method_name)
        // call the method
        return call(method, args, file, row, col)
    } else {
        // UFCS: find the method in the caller scope
        let method = caller_scope.lookup(method_name)
        let found = (method != NotFound && is(method, Types.ES_Function))
        ensure(found, 'method_not_found', method_name)
        // call the method
        return call(method, [object, ...args], file, row, col)
    }
}


/**
 *  Interface Object
 */

class Interface {
    constructor (name, method_table, implemented = {}, def_point = null) {
        assert(is(name, Types.String))
        assert(is(method_table, Types.TypedHash.of(Prototype)))
        assert(is(implemented, Types.TypedHash.of(Types.Overload)))
        this.name = name
        if (def_point != null) {
            let { f, row, col } = def_point
            this.desc = `interface ${name} at ${f} (row ${row}, column ${col})`
        } else {
            this.desc = `interface ${name} at (Built-in)`
        }
        this.method_table = copy(method_table)
        this.implemented = copy(implemented)
        Object.freeze(this.method_table)
        Object.freeze(this.implemented)
        this[Checker] = (object => {
            if (!is(object, Types.Instance)) { return false }
            return exists(object.class_.super_interfaces, I => I === this)
        })
        this[Solid] = true
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Interface'
    }
}

function create_interface (name, table) {
    // TODO
    /*
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
    */
}
