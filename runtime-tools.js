'use strict';


function FormatString (string, id_ref) {
    check(FormatString, arguments, { string: Str, id_ref: Function })
    return string.replace(/{([^}]+)}/g, (_, arg) => id_ref(arg))
}
