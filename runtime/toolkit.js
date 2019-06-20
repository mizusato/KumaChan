/**
 *  Tool Functions
 *
 *  Note that some of these functions return an iterator rather than an array,
 *    which is intentional to make chain operations more efficient.
 */
const ALPHABET = 'abcdefghijklmnopqrstuvwxyz'
const NotFound = { tip: 'Object Not Found' }
const CR = '\r'
const LF = '\n'
const TAB = '\t'


/**
 *  Checks if `object[key]` exists as an own property of `object`
 *
 *  @param key String | Symbol
 *  @param object Object
 *  @return Boolean
 */
function has (key, object) {
    assert(typeof key == 'string' || typeof key == 'symbol')
    assert(typeof object == 'object' && object !== null)
    return Object.prototype.hasOwnProperty.call(object, key)
}


/**
 *  Simple substitution of `object.__proto__`
 *
 *  @param object Object
 *  @return Object
 */
function get_proto (object) {
    assert(typeof object == 'object')
    return Object.getPrototypeOf(object)
}


/**
 *  Shorthand of `Object.assign`
 *
 *  @param o1 Object
 *  @param o2 Object
 *  @return Object (o1)
 */
function pour (o1, o2) {
    assert(typeof o1 == 'object' && o1 !== null)
    assert(typeof o2 == 'object' && o2 !== null)
    return Object.assign(o1, o2)
}


/**
 *  Converts iterable object to array
 *
 *  @param iterable Iterable
 *  @return Array
 */
function list (iterable) {
    let result = []
    for (let I of iterable) {
        result.push(I)
    }
    return result
}


/**
 *  Creates reversed iterator for array
 *
 *  @param list Array
 *  @return Iterator
 */
function *rev (list) {
    assert(list instanceof Array)
    for (let i=list.length-1; i>=0; i--) {
        yield list[i]
    }
}


/**
 *  Similar to Array.prototype.map, but maps iterable into iterator
 *
 *  @param iterable Iterable
 *  @return Iterator
 */
function *map (iterable, f) {
    let index = 0
    for (let I of iterable) {
        yield f(I, index)
        index += 1
    }
}


/**
 *  Takes until the nth element of given iterable
 *
 *  @param iterable Iterable
 *  @param n Integer
 *  @return Iterator
 */
function *take (iterable, n) {
    assert(Number.isSafeInteger(n))
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


/**
 *  Performs a parallel merge on an array of iterable objects
 *
 *  @param it_list Iterable[]
 *  @param f Function
 *  @return Iterator
 */
function *zip (it_list, f) {
    assert(it_list instanceof Array)
    assert(typeof f == 'function')
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


/**
 *  Creates a new hash table with keys from `object` mapped by `f`
 *
 *  @param object Object
 *  @param f Function
 *  @return Object
 */
function mapkey (object, f) {
    assert(object instanceof object)
    assert(typeof f == 'function')
    let mapped = {}
    for (let key of Object.keys(object)) {
        let value = object[key]
        mapped[f(key, value)] = value
    }
    return mapped
}


/**
 *  Creates a new hash table with values from `object` mapped by `f`
 *
 *  @param object Object
 *  @param f Function
 *  @return Object
 */
function mapval (object, f) {
    let mapped = {}
    for (let key of Object.keys(object)) {
        mapped[key] = f(object[key], key)
    }
    return mapped
}


/**
 *  Creates an iterator with elements mapped from `object` as `f(key, value)`
 *
 *  @param object Object
 *  @param f Function
 *  @return Iterator
 */
function *mapkv (object, f) {
    for (let key of Object.keys(object)) {
        yield f(key, object[key])
    }
}


/**
 *  Performs a shallow copy of an array or a hash table
 *
 *  @param object Array | Object
 *  @return Array | Object
 */
function copy (object) {
    if (object instanceof Array) {
        return list(map(object, x => x))
    } else {
        assert(typeof object == 'object')
        return mapval(object, x => x)
    }
}


/**
 *  Performs a shallow equality test on a pair of arrays or hash tables
 *
 *  @param o1 Array | Object
 *  @param o2 Array | Object
 *  @param cmp Function?
 *  @return Boolean
 */
function equal (o1, o2, cmp = (x, y) => (x === y)) {
    if (o1 instanceof Array && o2 instanceof Array) {
        return (
            o1.length == o2.length
                && forall(o1, (e,i) => cmp(e, o2[i]))
        )
    } else {
        assert(typeof o1 == 'object' && typeof o2 == 'object')
        let k1 = Object.keys(o1)
        let k2 = Object.keys(o2)
        return (
            k1.length == k2.length
                && forall(k1, k => has(k,o2) && cmp(o1[k], o2[k]))
        )
    }
}


/**
 *  Performs a shallow equality test on two sets
 *
 *  @param s1 Set
 *  @param s2 Set
 *  @return Boolean
 */
function set_equal (s1, s2) {
    assert(s1 instanceof Set)
    assert(s2 instanceof Set)
    if (s1.size != s2.size) { return false }
    return forall(s1, e => s2.has(e))
}


/**
 *  Apply `f` on elements of an iterable object or entries of a hash table
 *
 *  @param something Iterable | Object
 *  @param f Function
 */
function foreach (something, f) {
    if (typeof something[Symbol.iterator] == 'function') {
        let i = 0
        for (let I of something) {
            f(I, i)
            i += 1
        }
    } else {
        assert(typeof something == 'object')
        for (let key of Object.keys(something)) {
            f(key, something[key])
        }
    }
}


/**
 *  Similar to Array.prototype.filter but filters iterable into iterator
 *
 *  @param iterable Iterable
 *  @param f Function
 *  @return Iterator
 */
function *filter (iterable, f) {
    let index = 0
    for (let I of iterable) {
        if (f(I, index)) {
            yield I
        }
        index += 1
    }
}


/**
 *  Creates a new hash table with keys and values from `object` filtered by `f`
 *
 *  @param object Object
 *  @param f Function
 *  @return Object
 */
function flkv (object, f) {
    let result = {}
    for (let key of Object.keys(object)) {
        if (f(key, object[key])) {
            result[key] = object[key]
        }
    }
    return result
}


/**
 *  Concatenates `iterables`
 *
 *  @param ...iterables Iterable[]
 *  @return Iterable
 */
function *cat (...iterables) {
    for (let iterable of iterables) {
        for (let I of iterable) {
            yield I
        }
    }
}


/**
 *  Flattens an iterable of iterable objects
 *
 *  @param iterable_of_iterable Iterable
 *  @return Iterable
 */
function *flat (iterable_of_iterable) {
    for (let iterable of iterable_of_iterable) {
        for (let I of iterable) {
            yield I
        }
    }
}


/**
 *  Similar to Array.prototype.join but creates string from iterable object
 *
 *  @param iterable Iterable
 *  @return String
 */
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


/**
 *  Looks for an element in `iterable` that makes f(element) a truthy value,
 *    returns `NotFound` when such element does not exist
 *
 *  @param iterable Iterable
 *  @param f Function
 *  @return Any
 */
function find (iterable, f) {
    let index = 0
    for (let I of iterable) {
        if (f(I, index)) {
            return I
        }
        index += 1
    }
    return NotFound
}


/**
 *  Evaluate `next_of(next_of(...(next_of(initial))))` until terminate
 *    condition is satisfied
 *
 *  @param initial Any
 *  @param next_of Function
 *  @return Iterator
 */
function *iterate (initial, next_of, terminate) {
    // apply next_of() on value until terminal condition satisfied
    let value = initial
    while (!terminate(value)) {
        yield value
        value = next_of(value)
    }
}


/**
 *  Similar to Array.prototype.reduce but reduces iterable object
 *
 *  @param iterable Iterable
 *  @param initial Any
 *  @param f Function
 *  @param terminate Function
 *  @return Any
 */
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


/**
 *  Tests if ∀ I ∈ iterable, f(I) === true
 *
 *  @param iterable Iterable
 *  @param f Function (must return boolean value)
 *  @return Boolean
 */
function forall (iterable, f) {
    return fold(
        iterable, true, ((e,v,i) => v && f(e,i)), (v => {
            assert(typeof v == 'boolean')
            return (v == false)
        })
    )
}


/**
 *  Tests if ∃ I ∈ iterable, f(I) === true
 *
 *  @param iterable Iterable
 *  @param f Function (must return boolean value)
 *  @return Boolean
 */
function exists (iterable, f) {
    return fold(
        iterable, false, ((e,v,i) => v || f(e,i)), (v => {
            assert(typeof v == 'boolean')
            return (v == true)
        })
    )
}


/**
 *  Composite `functions`
 *
 *  @param functions Function[]
 *  @return Function
 */
function chain (functions) {
    assert(functions instanceof Array)
    assert(functions.every(f => typeof f == 'function'))
    return ( x => fold(functions, x, (f, v) => f(v)) )
}


/**
 *  Count from 0 to n-1
 *
 *  @param n Integer
 *  @return Iterator
 */
function *count (n) {
    assert(Number.isSafeInteger(n))
    let i = 0
    while (i < n) {
        yield i
        i += 1
    }
}


/**
 *  Checks if there is no repeated element in the `iterable`
 *
 *  @param iterable Iterable
 *  @return Boolean
 */
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


/**
 *  Wraps function with specified arity given to it
 *
 *  @param f Function
 *  @param n Integer
 *  @return Function
 */
function give_arity(f, n) {
    assert(Number.isSafeInteger(n))
    assert(0 <= n && n <= ALPHABET.length)
    let para_list = join(filter(ALPHABET, (e,i) => i < n), ',')
    let g = new Function(para_list, 'return this.apply(null, arguments)')
    return g.bind(f)
}


/**
 *  Extracts a summary of `string`
 *
 *  @param string String
 *  @param n Integer
 *  @return String
 */
function get_summary (string, n = 60) {
    assert(typeof string == 'string')
    assert(Number.isSafeInteger(n))
    let t = string.substring(0, n).replace(/[\n \t]+/g, ' ').trim()
    if (string.length <= n) {
        return t
    } else {
        return (t + '...')
    }
}
