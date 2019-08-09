register_simple_module('Math', {
    PI: Math.PI,
    E: Math.E,
    sqrt: fun (
        'function sqrt (x: Number) -> Number',
            x => {
                ensure(x >= 0, 'sqrt_of_negative')
                return Math.sqrt(x)
            }
    ),
    cbrt: fun (
        'function cbrt (x: Number) -> Number',
            x => Math.cbrt(x)
    ),
    exp: fun (
        'function exp (x: Number) -> Number',
            x => Math.exp(x)
    ),
    log: fun (
        'function log (x: Number) -> Number',
            x => {
                ensure(x > 0, 'log_of_non_positive')
                return Math.log(x)
            }
    ),
    rad2deg: fun (
        'function rad2deg (x: Number) -> Number',
            x => 180 * x / Math.PI
    ),
    deg2rad: fun (
        'function deg2rad (x: Number) -> Number',
            x => Math.PI * x / 180
    ),
    sin: fun (
        'function sin (x: Number) -> Number',
            x => Math.sin(x)
    ),
    cos: fun (
        'function cos (x: Number) -> Number',
            x => Math.cos(x)
    ),
    tan: fun (
        'function tan (x: Number) -> Number',
            x => Math.tan(x)
    ),
    asin: fun (
        'function asin (x: Number) -> Number',
            x => {
                ensure(-1 <= x && x <= 1, 'asin_out_of_domain')
                return Math.asin(x)
            }
    ),
    acos: fun (
        'function acos (x: Number) -> Number',
            x => {
                ensure(-1 <= x && x <= 1, 'acos_out_of_domain')
                return Math.acos(x)
            }
    ),
    atan: fun (
        'function atan (x: Number) -> Number',
            x => Math.atan(x),
    ),
    atan2: fun (
        'function atan2 (y: Number, x: Number) -> Number',
            (y, x) => Math.atan2(y, x)
    )
})
