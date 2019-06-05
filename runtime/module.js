let modules = {}

function register_module (module_name, export_names, init) {
    assert(is(name, Types.String))
    assert(is(export_names, TypedList.of(Types.String)))
    assert(is(init, ES.Function))
    ensure(!has(module_name, modules), 'module_conflict', module_name)
    let scope = new Scope(Global)
    init(scope)
    let exported = {}
    foreach(export_names, name => {
        ensure(scope.has(name), 'missing_export', module_name, name)
        exported[name] = scope.lookup(name)
    })
    modules[module_name] = { scope, exported }
}
