pour(built_in_functions, {
    // OO
    get_class: fun (
        'function get_class (i: Instance) -> Class',
            i => i.class_
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
