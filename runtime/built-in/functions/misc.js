pour(built_in_functions, {
    // OO
    get_class: fun (
        'function get_class (i: Instance) -> Class',
            i => i.class_
    ),
    // Input
    read: fun (
        'function read (format: String) -> Object',
            format => {
                try {
                    require.resolve('scanf')
                } catch (e) {
                    if (e.code == 'MODULE_NOT_FOUND') {
                        ensure(false, 'scanf_not_found')
                    }
                }
                let result = require('scanf')(format)
                let normalize = x => {
                    if (x === null || Number.isNaN(x)) {
                        return Nil
                    } else {
                        return x
                    }
                } 
                if (is(result, Types.List)) {
                    return result.map(normalize)
                } else {
                    return normalize(result)
                }
            }
    ),
    // Ouput
    print: f (
        'print',
        'function print (p: Bool) -> Void',
            x => (console.log(x.toString()), Void),
        'function print (x: Number) -> Void',
            x => (console.log(x.toString()), Void),
        'function print (s: String) -> Void',
            s => (console.log(s), Void)
    ),
    // Error Handling
    custom_error: f (
        'custom_error',
        'function custom_error (msg: String) -> Error',
            msg => new CustomError(msg),
        'function custom_error (name: String, msg: String) -> Error',
            (name, msg) => new CustomError(msg, name),
        'function custom_error (name: String, msg: String, data: Hash) -> Error'
            ,(name, msg, data) => new CustomError(msg, name, data)
    ),
})
