pour(built_in_functions, {
    abs: fun (
        'function abs (x: Number) -> Number',
            x => Math.abs(x)
    ),
    rand: fun (
        'function rand () -> Number',
            () => Math.random()
    ),
    floor: fun (
        'function floor (x: Number) -> Number',
            x => Math.floor(x)
    ),
    ceil: fun (
        'function ceil (x: Number) -> Number',
            x => Math.ceil(x)
    ),
    round: fun (
        'function round (x: Number) -> Number',
            x => Math.round(x)
    )
})
