/**
 *  Module System
 *
 *  All registered modules are put in a global registry,
 *    therefore module names are global identifiers,
 *    two modules cannot have a same name.
 */


let modules = {}

class Module {
    constructor (name, exported) {
        assert(is(name, Types.String))
        assert(is(exported, Types.Hash))
        this.name = name
        this.exported = exported
    }
    has (name) {
        return has(name, this.exported)
    }
    get (name) {
        assert(is(name, Types.String))
        assert(this.has(name))
        return this.exported[name]
    }
    import_to (scope, name, alias) {
        assert(scope instanceof Scope)
        assert(is(name, Types.String))
        assert(is(alias, Types.String))
        ensure(!scope.has(alias), 'import_conflict', alias)
        scope.declare(alias, this.exported[name])
    }
    import_all_to (scope) {
        assert(scope instanceof Scope)
        let names = Object.keys(this.exported)
        for (let name of names) {
            ensure(!scope.has(name), 'import_conflict', name)
        }
        for (let name of names) {
            scope.declare(name, this.exported[name])
        }
    }
    get [Symbol.toStringTag] () {
        return "Module"
    }
}

Types.Module = $(x => x instanceof Module)


function register_module (module_name, export_names, init) {
    assert(is(module_name, Types.String))
    assert(is(export_names, TypedList.of(Types.String)))
    assert(is(init, ES.Function))
    ensure(!has(module_name, modules), 'module_conflict', module_name)
    assert(typeof Global != 'undefined')  // global scope should be created
    let scope = new Scope(Global)
    init(scope)
    let exported = {}
    foreach(export_names, name => {
        ensure(scope.has(name), 'missing_export', module_name, name)
        exported[name] = scope.lookup(name)
    })
    modules[module_name] = new Module(module_name, exported)
}

function register_simple_module (name, export_object) {
    assert(is(export_object, Types.Hash))
    let keys = Object.keys(export_object)
    register_module(name, keys, scope => {
        foreach(export_object, (name, value) => {
            scope.declare(name, value)
        })
    })
}


// ImportConfig: [NAME, ALIAS]
let ImportConfig = Ins(Types.List, $(
    x => x.length == 2 && forall(x, y => is(y, Types.String))
))

function import_module (scope, config) {
    // import MODULE
    assert(scope instanceof Scope)
    assert(is(config, ImportConfig))
    let [name, alias] = config
    ensure(has(name, modules), 'module_not_exist', name)
    ensure(!scope.has(alias), 'import_conflict_mod', name, alias)
    scope.declare(alias, modules[name])
    return Void
}

function import_names (scope, module_name, configs) {
    // import NAME1, NAME2, ... from MODULE
    assert(scope instanceof Scope)
    assert(is(module_name, Types.String))
    assert(is(configs, TypedList.of(ImportConfig)))
    ensure(has(module_name, modules), 'module_not_exist', module_name)
    let mod = modules[module_name]
    for (let config of configs) {
        let [_, alias] = config
        ensure(!scope.has(alias), 'import_conflict', alias)
    }
    for (let config of configs) {
        let [name, alias] = config
        mod.import_to(scope, name, alias)
    }
    return Void
}

function import_all (scope, module_name) {
    // import * from MODULE
    assert(scope instanceof Scope)
    assert(is(module_name, Types.String))
    ensure(has(module_name, modules), 'module_not_exist', module_name)
    modules[module_name].import_all_to(scope)
    return Void
}
