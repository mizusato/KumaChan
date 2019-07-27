pour(built_in_functions, {
    get_local_unix_time: fun (
        'function get_local_unix_time () -> Int',
            () => Date.now()
    ),
    get_local_timezone_offset: fun (
        'function get_local_timezone_offset () -> Int',
            () => -((new Date()).getTimezoneOffset()) / 60
    )
})
