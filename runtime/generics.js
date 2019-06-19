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
            interfaces: new Set()
        }
        this.info = { only_classified: true }
        this.inflate = this.inflate.bind(this)
        inject_desc(this.inflate, 'inflate_template')
        this[Checker] = (x => {
            if (is(x, Types.Instance)) {
                let is_inflated_class = C => this.inflated.classes.has(C)
                if (exists(x.class_.super_classes, is_inflated_class)) {
                    return true
                } else if (this.inflated.interfaces.size == 0) {
                    return false
                }
                let is_inflated_interface = I => this.inflated.interfaces.has(I)
                if (exists(x.class_.super_interfaces, is_inflated_interface)) {
                    return true
                } else if (this.info.only_classified) {
                    return false
                }
            } else if (is(x, Types.Struct)) {
                if (this.inflated.schema.has(x.schema)) {
                    return true
                } else if (this.info.only_classified) {
                    return false
                }
            }
            return exists(this.cache, item => is(x, item.type))
        })
        Object.freeze(this)
    }
    inflate (...args) {
        let cached = find(this.cache, item => {
            return equal(item.args, args, (A, B) => {
                if (is(A, Type) && is(B, Type)) {
                    return type_equivalent(A, B)
                } else {
                    return A === B
                }
            })
        })
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
            } else if (is(type, Types.Interface)) {
                this.inflated.interfaces.add(type)
            } else if (is(type, Types.Schema)) {
                this.inflated.schema.add(type)
            } else {
                this.info.only_classified = false
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
