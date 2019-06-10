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
        this.inflated = {
            classes: new Set(),
            schema: new Set(),
            interfaces: new Set(),
            others: new Set()
        }
        this.inflate = this.inflate.bind(this)
        inject_desc(this.inflate, 'inflate_template')
        this[Checker] = (x => {
            if (is(x, Types.Instance)) {
                if (this.inflated.classes.has(x.class_)) {
                    return true
                } else if (this.inflated.interfaces.size == 0) {
                    return false
                }
                let is_inflated = I => this.inflated.interfaces.has(I)
                if (exists(x.class_.super_interfaces, is_inflated)) {
                    return true
                } else if (this.inflated.others.size == 0) {
                    return false
                }
            } else if (is(x, Types.Struct)) {
                if (this.inflated.schema.has(x.schema)) {
                    return true
                } else if (this.inflated.others.size == 0) {
                    return false
                }
            }
            return exists(this.cache, item => is(x, item.type))
        })
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
            if (is(type, Types.Class)) {
                this.inflated.classes.add(type)
            } else if (is(type, Types.Schema)) {
                this.inflated.schema.add(type)
            } else if (is(type, Types.Interface)) {
                this.inflated.interfaces.add(type)
            } else {
                this.inflated.others.add(type)
            }
            return type
        }
    }
    get [Symbol.toStringTag]() {
        return 'TypeTemplate'
    }
}

Types.TypeTemplate = $(x => x instanceof TypeTemplate)

/* shorthand */
let template = ((p,r) => new TypeTemplate(fun(p,r)))
