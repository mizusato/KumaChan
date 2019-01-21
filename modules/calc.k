module calc


export pi, e
export abs, sqrt, cbrt
export sin, cos, tan
export asin, acos, atan, atan2
export exp, log


let pi = /~ Math.PI ~/
let e = /~ Math.exp(1) ~/


local abs (Number x) {
    return /~ Math.abs(id('x')) ~/
}


local sqrt (Number x) {
    return /~ Math.sqrt(id('x')) ~/
}


local cbrt (Number x) {
    return /~ Math.pow(id('x'), 1/3) ~/
}


local sin (Number x) {
    return /~ Math.sin(id('x')) ~/
}


local cos (Number x) {
    return /~ Math.cos(id('x')) ~/
}


local tan (Number x) {
    return /~ Math.tan(id('x')) ~/
}


local asin (Number x) {
    return /~ Math.asin(id('x')) ~/
}


local acos (Number x) {
    return /~ Math.acos(id('x')) ~/
}


local atan (Number x) {
    return /~ Math.atan(id('x')) ~/
}


local atan2 (Number y, Number x) {
    return /~ Math.atan2(id('y'), id('x')) ~/
}


local exp (Number x) {
    return /~ Math.exp(id('x')) ~/
}


local log (Number x) {
    return /~ Math.log(id('x')) ~/
}
