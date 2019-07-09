'use strict';


(function () {

    '<constants>';
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

    let export_object = {
        CustomError, RuntimeError, AssertionFailed,
        Global, call_by_js, inject_desc,
        [R_VOID]: Void,
        [R_EVAL_SCOPE]: Eval,
        [R_NEW_SCOPE]: new_scope,
        [R_REG_MODULE]: register_module,
        [R_GET_HELPERS]: get_helpers,
    }
    Object.freeze(export_object)

    if (typeof window != 'undefined') {
        window[RUNTIME] = export_object
    } else {
        module.exports = export_object
    }

})()
