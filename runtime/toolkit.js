const ALPHABET = 'abcdefghijklmnopqrstuvwxyz'
const NotFound = { tip: 'Object Not Found' }


function pour (o1, o2) {
    return Object.assign(o1, o2)
}


function list (iterable) {
    // convert iterable to array
    let result = []
    for (let I of iterable) {
        result.push(I)
    }
    return result
}


function run (iterable) {
    for (let I of iterable) {}
}


function *rev (list) {
    for (let i=list.length-1; i>=0; i--) {
        yield list[i]
    }
}


function *map (iterable, f) {
    let index = 0
    for (let I of iterable) {
        yield f(I, index)
        index += 1
    }
}


function mapkey (object, f) {
    let mapped = {}
    for (let key of Object.keys(object)) {
        let value = object[key]
        mapped[f(key, value)] = value
    }
    return mapped
}


function mapval (object, f) {
    let mapped = {}
    for (let key of Object.keys(object)) {
        mapped[key] = f(object[key], key)
    }
    return mapped
}


function *mapkv (object, f) {
    for (let key of Object.keys(object)) {
        yield f(key, object[key])
    }
}


function foreach (something, f) {
    if (typeof something[Symbol.iterator] == 'function') {
        return run(map(something, f))
    } else {
        return run(mapkv(something, f))
    }
}


function *filter (iterable, f) {
    let index = 0
    for (let I of iterable) {
        if (f(I, index)) {
            yield I
        }
        index += 1
    }
}


function flkv (object, f) {
    let result = {}
    for (let key of Object.keys(object)) {
        if (f(key, object[key])) {
            result[key] = object[key]
        }
    }
    return result
}


function *cat (...iterables) {
    for (let iterable of iterables) {
        for (let I of iterable) {
            yield I
        }
    }
}


function *flat (iterable_of_iterable) {
    for (let iterable of iterable_of_iterable) {
        for (let I of iterable) {
            yield I
        }
    }
}


function join (iterable, separator) {
    let string = ''
    let first = true
    for (let I of iterable) {
        if (first) {
            first = false
        } else {
            string += separator
        }
        string += I
    }
    return string
}


function find (iterable, f) {
    for (let I of iterable) {
        if (f(I)) {
            return I
        }
    }
    return NotFound
}


function *iterate (initial, next_of, terminate) {
    // apply next_of() on value until terminal condition satisfied
    let value = initial
    while (!terminate(value)) {
        yield value
        value = next_of(value)
    }
}


function fold (iterable, initial, f, terminate) {
    // reduce() with a terminal condition
    let index = 0
    let value = initial
    for (let I of iterable) {
        value = f(I, value, index)
        index += 1
        if (terminate && terminate(value)) {
            break
        }
    }
    return value
}


function forall (iterable, f) {
    // ∀ I ∈ iterable, f(I) == true
    return fold(
        iterable, true, ((e,v,i) => v && f(e,i)), (v => v == false)
    )
}


function exists (iterable, f) {
    // ∃ I ∈ iterable, f(I) == true
    return fold(
        iterable, false, ((e,v,i) => v || f(e,i)), (v => v == true)
    )
}


function chain (functions) {
    return ( x => fold(functions, x, (f, v) => f(v)) )
}


function *count (n) {
    let i = 0
    while (i < n) {
        yield i
    }
}


function no_repeat (iterable) {
    let s = new Set()
    for (let I of iterable) {
        if (!s.has(I)) {
            s.add(I)
        } else {
            return false
        }
    }
    return true
}


function give_arity(f, n) {
    // Tool to fix arity of wrapped function
    let para_list = join(filter(ALPHABET, (e,i) => i < n), ',')
    let g = new Function(para_list, 'return this.apply(null, arguments)')
    return g.bind(f)
}

