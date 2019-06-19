'use strict';


(function () {

    '<include> toolkit.js';
    '<include> msg.js';
    '<include> error.js';
    '<include> type.js';
    '<include> function.js';
    '<include> oo.js';
    '<include> generics.js';
    '<include> module.js';
    '<include> built-in/built-in.js';

    const export_name = 'KumaChan'

    let export_object = {
        Void,
        RuntimeError, CustomError,
        Global, Eval, new_scope,
        register_module,
        get_helpers
    }
    Object.freeze(export_object)

    if (typeof window != 'undefined') {
        window[export_name] = export_object
    } else {
        module.exports = export_object
    }

})()
