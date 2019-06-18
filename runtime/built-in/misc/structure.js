class Struct {
    constructor (schema, data, is_view = false) {
        assert(is(schema, Types.Schema))
        assert(is(data, Types.Hash))
        this.schema = schema
        this.data = is_view? data: copy(data)
        Object.freeze(this)
        let result = this.schema.check_all(this.data, is_view)
        if (!result.ok) {
            if (result.why == 'miss') {
                ensure(false, 'invalid_struct_init_miss', result.key)
            } else if (result.why == 'key') {
                ensure(false, 'invalid_struct_init_key', result.key)
            } else {
                ensure(false, 'invalid_struct_init_req')
            }
        }
    }
    create_view (schema) {
        return new Struct(schema, this.data, true)
    }
    has (key) {
        return this.schema.has_key(key)
    }
    get (key) {
        ensure(has(key, this.data), 'struct_field_missing', key)
        let value = this.data[key]
        let ok = this.schema.check(key, value)
        ensure(ok, 'struct_inconsistent', key)
        return value
    }
    set (key, value) {
        ensure(has(key, this.data), 'struct_field_missing', key)
        let ok = this.schema.check(key, value)
        ensure(ok, 'struct_field_invalid', key)
        this.data[key] = value
    }
    keys () {
        return this.schema.get_keys()
    }
    get [Symbol.toStringTag]() {
        return 'Struct'
    }
}

class Schema {
    constructor (name, table, defaults = {}, operators = {}) {
        assert(is(name, Types.String))
        assert(is(table, TypedHash.of(Type)))
        assert(is(defaults, Types.Hash))
        assert(forall(Object.keys(defaults), k => has(k, table)))
        assert(is(operators, TypedHash.of(Types.Function)))
        foreach(defaults, (field, value) => {
            ensure(is(value, table[field]), 'schema_invalid_default', field)
        })
        this.name = name
        this.table = copy(table)
        this.defaults = copy(defaults)
        this.operators = copy(operators)
        Object.freeze(this.table)
        Object.freeze(this.defaults)
        Object.freeze(this.operators)
        this.create_struct_from_another = fun (
            'function create_struct_from_another (s: Struct) -> Struct',
                s => s.create_view(this)
        )
        this[Checker] = x => (x instanceof Struct && x.schema === this)
        Object.freeze(this)
    }
    has_key (key) {
        return has(key, this.table)
    }
    get_keys () {
        return Object.keys(this.table)
    }
    check_all (hash, is_view) {
        assert(is(hash, Types.Hash))
        let table = this.table
        let defaults = this.defaults
        for (let k of Object.keys(table)) {
            let T = table[k]
            if (has(k, hash)) {
                if (is(hash[k], T)) {
                    continue
                } else {
                    return { ok: false, why: 'key', key: k }
                }
            } else if (!is_view && has(k, defaults)) {
                hash[k] = defaults[k]
                continue
            } else {
                return { ok: false, why: 'miss', key: k }
            }
        }
        return { ok: true }
    }
    check (key, value) {
        assert(has(key, this.table))
        return is(value, this.table[key])
    }
    defined_operator (name) {
        return has(name, this.operators)
    }
    get_operator (name) {
        assert(this.defined_operator(name))
        return this.operators[name]
    }
    get [Symbol.toStringTag]() {
        return 'Schema'
    }
}

Types.Schema = $(x => x instanceof Schema)
Types.Struct = $(x => x instanceof Struct)

function create_schema (name, table, defaults, config) {
    let { ops } = config
    foreach(table, (name, constraint) => {
        ensure(is(constraint, Type), 'schema_invalid_field', name)
    })
    return new Schema(name, table, defaults, ops)
}

function new_structure (schema, hash) {
    ensure(is(schema, Types.Schema), 'not_schema')
    assert(is(hash, Types.Hash))
    return new Struct(schema, hash)
}

function get_common_schema (s1, s2) {
    assert(is(s1, Types.Struct))
    assert(is(s2, Types.Struct))
    ensure(s1.schema === s2.schema, 'different_schema')
    return s1.schema
}
