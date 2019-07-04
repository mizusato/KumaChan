'use strict';


(function () {

    '<include> msg.js';
    '<include> assertion.js';
    '<include> toolkit.js';
    '<include> error.js';
    '<include> type.js';
    '<include> function.js';
    '<include> oo.js';
    '<include> generics.js';
    '<include> module.js';
    '<include> built-in/built-in.js';
    '<include> modules/ES.js';

    const export_name = 'KumaChan'

    let export_object = {
        CustomError, RuntimeError, AssertionFailed,
        Global, Eval, new_scope,
        call_by_js, inject_desc,
        Void,
        register_module,
        get_helpers,
    }
    Object.freeze(export_object)

    if (typeof window != 'undefined') {
        window[export_name] = export_object
    } else {
        module.exports = export_object
    }

})()
