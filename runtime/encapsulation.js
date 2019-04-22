/**
 *  Encapsulation (Class, Instance, Signature, Interface)
 */


 /**
  *  Tool Functions
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
    function _get_super_classes (class_) {
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


/**
 *  Class Object
 */

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
                F.raw(scope, expose)
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
            if (object instanceof Instance) {
                return exists(
                    object.abstraction.super_classes,
                    super_class => super_class === this
                )
            } else {
                return false
            }
        })
        Object.freeze(this)
    }
    get [Symbol.toStringTag]() {
        return 'Class'
    }
}

function create_class (desc, impls, init, methods, static_methods = {}) {
    return new Class(impls, init, methods, static_methods, desc)
}


/**
 *  Instance Object
 */

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

let method_err = new ErrorProducer(MethodError)

function call_method (object, caller_scope, method_name, args) {
    if (object instanceof Instance) {
        // method on instance
        let method = object.methods[method_name]
        method_err.assert(method, method || MSG.method_not_found(method_name))
        let ok = !(IsRef(object) && method[WrapperInfo].proto.affect != 'local')
        method_err.assert(
            ok, ok || MSG.instance_immutable(method[WrapperInfo.desc])
        )
        return call(method, args)
    } else {
        // UFCS
        let method = caller_scope.lookup(method_name)
        let found = method != NotFound && is(method, Type.Function)
        method_err.assert(found, found || MSG.method_not_found(method_name))
        return call(method, [object, ...args])
    }
}


/**
 *  Signature Object
 */

let Input = list_of(Type.Abstract)
let Output = Type.Abstract

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
                return Signature.check_sole(f, this.input, this.output)
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
    static check_sole (f, input, output) {
        let proto = f[WrapperInfo].proto
        return (proto.value === output) && (
            proto.parameters.length == input.length
        ) && forall(
            input, (I,i) => proto.parameters[i].constraint === I
        )
    }
    get [Symbol.toStringTag]() {
        return 'Signature'
    }
}

function sig (input, output) {
    return new Signature(input, output)
}


/**
 *  Interface Object
 */

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
