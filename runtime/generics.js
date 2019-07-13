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
        let self = {
            inflater,
            cache: new VectorMapCache(),
            inflated: {
                classes: new Set(),
                interfaces: new Set(),
                schemas: new Set(),
                others: new Set(),
                all: new Set()
            }
        }
        this.inflate = this.inflate.bind(self)
        inject_desc(this.inflate, 'inflate_template')
        this.has_inflated = this.has_inflated.bind(self)
        this[Checker] = this.check.bind(self)
        Object.freeze(this)
    }
    check (x) {
        assert(!(this instanceof TypeTemplate))
        let { classes, interfaces, schemas, others } = this.inflated
        if (is(x, Types.Instance)) {
            if (exists(x.class_.super_classes, C => classes.has(C))) {
                return true
            }
            if (exists(x.class_.super_interfaces, I => interfaces.has(I))) {
                return true
            }
        } else if (is(x, Types.Struct)) {
            if (schemas.has(x.schema)) {
                return true
            }
        }
        return exists(others, T => is(x, T))
    }
    inflate (...args) {
        assert(is(args, Types.List))
        let cached = this.cache.find(args)
        if (cached !== NotFound) {
            return cached
        }
        foreach(args, (arg, i) => {
            let valid = is(arg, Type) || is(arg, Types.Primitive)
            ensure(valid, 'arg_invalid_inflate', `#${i+1}`)
        })
        let type = call(this.inflater, args)
        ensure(is(type, Type), 'retval_invalid_inflate')
        this.cache.set(args, type)
        if (is(type, Types.Class)) {
            this.inflated.classes.add(type)
        } else if (is(type, Types.Interface)) {
            this.inflated.interfaces.add(type)
        } else if (is(type, Types.Schema)) {
            this.inflated.schemas.add(type)
        } else {
            this.inflated.others.add(type)
        }
        this.inflated.all.add(type)
        return type
    }
    has_inflated (type) {
        assert(is(type, Types.Type))
        return this.inflated.all.has(type)
    }
    get [Symbol.toStringTag]() {
        return 'TypeTemplate'
    }
}

Types.TypeTemplate = $(x => x instanceof TypeTemplate)

/* shorthand */
let template = ((p,r) => new TypeTemplate(fun(p,r)))
