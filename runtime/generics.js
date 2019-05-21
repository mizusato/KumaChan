/**
 *  TypeTemplate: Something just like generics but not generics
 *
 *  A type template is a function-like type object, which can be inflated by
 *    a fixed number of arguments (types or primitives) and returns a new type.
 *  The arguments and returned type will be cached. If the same arguments
 *    are provided in the next inflation, the cached type will be returned.
 *  For any object x and type template TT, x ∈ TT if and only if there exists
 *    a cached type T of TT that satisfies x ∈ T.
 */
class TypeTemplate {
    constructor (inflater) {
        assert(is(inflater, Types.Function))
        this.inflater = inflater
        this.cache = []
        this.inflate = this.inflate.bind(this)
        inject_desc(this.inflate, 'inflate_template')
        this[Checker] = (x => exists(this.cache, item => is(x, item.type)))
        Object.freeze(this)
    }
    inflate (...args) {
        let cached = find(this.cache, item => equal(item.args, args))
        if (cached !== NotFound) {
            return cached.type
        } else {
            foreach(args, (arg, i) => {
                let is_type = is(arg, Type) || is(arg, Types.Primitive)
                ensure(is_type, 'arg_invalid_inflate', `#${i+1}`)
            })
            let type = call(this.inflater, args)
            ensure(is(type, Type), 'retval_invalid_inflate')
            this.cache.push({
                args: copy(args),
                type: type
            })
            return type
        }
    }
    get [Symbol.toStringTag]() {
        return 'TypeTemplate'
    }
}


class ClassTemplate {
    constructor (inflater) {
        this.type_template = new TypeTemplate(inflater)
        this.inflated = new WeakSet()
        this[Checker] = (
            x => is(x, Types.Instance) && this.inflated.has(x.class_)
        )
        Object.freeze(this)
    }
    inflate (...args) {
        let C = this.type_template.inflate.apply(null, args)
        assert(is(C, Types.Class))
        this.inflated.add(C)
        return C
    }
}


Types.ClassTemplate = $(x => x instanceof ClassTemplate)
Types.TypeTemplate = Uni(
    Types.ClassTemplate,
    $(x => x instanceof TypeTemplate)
)


/* shorthand */
let template = ((p,r) => new TypeTemplate(fun(p,r)))
let class_template = ((p,r) => new ClassTemplate(fun(p,r)))
