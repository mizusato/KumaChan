'use strict';


(function () {

    '<include> toolkit.js';
    '<include> msg.js';
    '<include> error.js';
    '<include> type.js';
    '<include> function.js';
    '<include> oo.js';
    '<include> built-in.js';

    let export_name = 'KumaChan'
    let export_object = {
        scope: { Global, Eval },
        fun, f, operators,
        wrap, overload, overload_added, overload_concated,
        create_interface, create_class,
        call, call_method,
        helpers
    }
    if (typeof window != 'undefined') {
        window[export_name] = export_object
    } else {
        module.exports = export_object
    }

})()
