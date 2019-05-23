class Structure {
    constructor (schema, data) {
        assert(is(schema, Types.Schema))
        assert(is(data, Types.Hash))
        this.schema = schema
        this.data = copy(data)
        let result = this.schema.check_all(this.data)
        if (!result.ok) {
            if (result.why == 'miss') {
                ensure(false, 'invalid_struct_init_miss', result.key)
            } else if (result.why == 'key') {
                ensure(false, 'invalid_struct_init_key', result.key)
            } else {
                ensure(false, 'invalid_struct_init_req')
            }
        }
        Object.freeze(this)
    }
    get (key) {
        let s = this.schema
        ensure(s.has_key(key), 'struct_key_error', key)
        let ok = s.check_key(this.data, key) && s.check_requirement(this.data)
        ensure(ok, 'struct_inconsistent', key)
        return s.get_value(this.data, key)
    }
    set (key, value) {
        let s = this.schema
        ensure(s.has_key(key), 'struct_key_error', key)
        let old_value = this.data[key]
        this.data[key] = value
        if (!s.check_key(this.data, key)) {
            this.data[key] = old_value
            ensure(false, 'struct_key_invalid', key)
        }
        if (!s.check_requirement(this.data)) {
            this.data[key] = old_value
            ensure(false, 'struct_req_violated', key)
        }
    }
    get [Symbol.toStringTag]() {
        return 'Structure'
    }
}

class Schema {
    constructor (name, table, defaults = {}, req = null, ops = {}) {
        let requirement = req
        let operators = ops
        assert(is(name, Types.String))
        assert(is(table, TypedHash.of(Type)))
        assert(is(defaults, Types.Hash))
        assert(forall(Object.keys(defaults), k => has(k, table)))
        assert(is(requirement, Uni(ES.Null, Types.Function)))
        if (requirement != null) {
            assert(requirement[WrapperInfo].proto.value_type === Types.Bool)
        }
        assert(is(operators, TypedHash.of(Types.Function)))
        this.name = name
        this.table = copy(table)
        this.defaults = copy(defaults)
        this.requirement = requirement
        this.operators = copy(operators)
        Object.freeze(this.table)
        Object.freeze(this.defaults)
        Object.freeze(this.operators)
        this.create = fun (
            'function create_structure (h: Hash) -> Structure',
                h => new Structure(this, h)
        )
        this[Checker] = x => (x instanceof Structure && x.schema === this)
        Object.freeze(this)
    }
    has_key (key) {
        return has(key, this.table)
    }
    get_value (hash, key) {
        assert(this.has_key(key))
        if (has(key, hash)) {
            return hash[key]
        } else if (has(key, this.defaults)) {
            return this.defaults[key]
        } else {
            assert(false)
        }
    }
    check_all (hash) {
        assert(is(hash, Types.Hash))
        let table = this.table
        for (let k of Object.keys(table)) {
            let T = table[k]
            if (has(k, hash)) {
                if (is(hash[k], T)) {
                    continue
                } else {
                    return { ok: false, why: 'key', key: k }
                }
            } else {
                return { ok: false, why: 'miss', key: k }
            }
        }
        if (this.check_requirement(hash)) {
            return { ok: true }
        } else {
            return { ok: false, why: 'req' }
        }
    }
    check_key (hash, key) {
        assert(is(hash, Types.Hash))
        assert(is(key, Types.String))
        assert(has(key, this.table))
        return (has(key, hash) && is(hash[key], this.table[key]))
    }
    check_requirement (hash) {
        assert(is(hash, Types.Hash))
        if (this.requirement !== null) {
            return call(this.requirement, [hash])
        } else {
            return true
        }
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

let SchemaType = $(x => x instanceof Schema)
let StructureType = $(x => x instanceof Structure)
let StructOperand = template (
    'function DefinedOperator (name: String) -> Type',
        name => $(x => is(x, Types.Structure) && x.defined_operator(name))
)

function create_schema (name, table, defaults, config) {
    let { req, ops } = config
    return new Schema(name, table, defaults, req, ops)
}

function new_structure (schema, hash) {
    ensure(is(schema, Types.Schema), 'not_schema')
    assert(is(hash, Types.Hash))
    return schema.create(hash)
}

function get_common_schema (s1, s2) {
    assert(is(s1, StructureType))
    assert(is(s2, StructureType))
    esnure(s1.schema === s2.schema, 'different_schema')
    return s1.schema
}

function apply_operator (name, s1, s2) {
    assert(is(name, Types.String))
    let s = get_common_schema(s1, s2)
    return call(s.get_operator(name), [s1, s2])
}

function apply_unary (name, s) {
    assert(is(name, Types.String))
    return call(s.schema.get_operator(name), [s])
}
