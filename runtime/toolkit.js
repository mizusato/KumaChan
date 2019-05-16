const ALPHABET = 'abcdefghijklmnopqrstuvwxyz'
const NotFound = { tip: 'Object Not Found' }
const hasOwnProperty = Object.prototype.hasOwnProperty
const CR = '\r'
const LF = '\n'
const TAB = '\t'


function has (key, object) {
    assert(typeof key == 'string' || typeof key == 'symbol')
    return hasOwnProperty.call(object, key)
}


function get_keys (object) {
    return Object.keys(object)
}


function get_vals (object) {
    return map(get_keys(object), k => object[k])
}


function get_proto (object) {
    return Object.getPrototypeOf(object)
}


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


function *take (iterable, n) {
    let index = 0
    for (let I of iterable) {
        if (index < n) {
            yield I
            index += 1
        } else {
            break
        }
    }
}


function *zip (it_list, f) {
    let iterators = it_list.map(iterable => iterable[Symbol.iterator]())
    while (true) {
        let results = iterators.map(it => it.next())
        if (exists(results, r => r.done)) {
            break
        } else {
            yield f(results.map(r => r.value))
        }
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


function copy (object) {
    if (object instanceof Array) {
        return list(map(object, x => x))
    } else {
        assert(typeof object == 'object')
        return mapval(object, x => x)
    }
}


function equal (o1, o2) {
    if (o1 instanceof Array && o2 instanceof Array) {
        return (
            o1.length == o2.length
                && forall(o1, (e,i) => e === o2[i])
        )
    } else {
        assert(typeof o1 == 'object' && typeof o2 == 'object')
        let k1 = Object.keys(o1)
        let k2 = Object.keys(o2)
        return (
            k1.length == k2.length
                && forall(k1, k => has(k,o2) && o1[k] === o2[k])
        )
    }
}


function foreach (something, f) {
    if (typeof something[Symbol.iterator] == 'function') {
        let i = 0
        for (let I of something) {
            f(I, i)
            i += 1
        }
    } else {
        for (let key of Object.keys(something)) {
            f(key, something[key])
        }
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
        iterable, true, ((e,v,i) => v && f(e,i)), (v => {
            assert(typeof v == 'boolean')
            return (v === false)
        })
    )
}


function exists (iterable, f) {
    // ∃ I ∈ iterable, f(I) == true
    return fold(
        iterable, false, ((e,v,i) => v || f(e,i)), (v => {
            assert(typeof v == 'boolean')
            return (v === true)
        })
    )
}


function chain (functions) {
    return ( x => fold(functions, x, (f, v) => f(v)) )
}


function *count (n) {
    let i = 0
    while (i < n) {
        yield i
        i += 1
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


function get_summary (string, n = 60) {
    let t = string.substring(0, n).replace(/[\n \t]+/g, ' ').trim()
    if (string.length <= n) {
        return t
    } else {
        return (t + '...')
    }
}
