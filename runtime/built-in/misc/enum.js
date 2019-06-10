/**
 *  Enum: An encapsulation of TypedHash<Singleton>
 */
class Enum {
    constructor (enum_name, item_names) {
        assert(is(enum_name, Types.String))
        assert(is(item_names, TypedList.of(Types.String)))
        this.hash = {}
        this.items = []
        this.map = new Map()
        foreach(item_names, name => {
            let item = create_value(`${enum_name}.${name}`)
            this.hash[name] = item
            this.items.push(item)
            this.map.set(item, name)
        })
        Object.freeze(this.hash)
        Object.freeze(this.items)
        this[Checker] = (x => is(x, Types.Singleton) && this.map.has(x))
        Object.freeze(this)
    }
    has (key) {
        return has(key, this.hash)
    }
    get (key) {
        ensure(this.has(key), 'key_error', key)
        return this.hash[key]
    }
    keys () {
        return Object.keys(this.hash)
    }
    get [Symbol.toStringTag]() {
        return 'Enum'
    }
}

Types.Enum = $(x => x instanceof Enum)
