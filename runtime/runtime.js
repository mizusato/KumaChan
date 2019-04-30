'use strict';


(function () {

    '<include> toolkit.js';
    '<include> msg.js';
    '<include> error.js';
    '<include> type.js';
    '<include> function.js';
    '<include> generics.js';
    '<include> oo.js';
    '<include> built-in.js';

    let export_name = 'KumaChan'
    let export_object = {
        is, Uni, Ins, Not, Types,
        fun, f, operators,
        wrap, get_vals, new_scope,
        overload, overload_added, overload_concated,
        create_interface, create_class,
        call, call_method,
        RuntimeError,
        helpers: get_helpers,
        scope: default_scopes
    }
    Object.freeze(export_object)
    if (typeof window != 'undefined') {
        window[export_name] = export_object
    } else {
        module.exports = export_object
    }

})()
