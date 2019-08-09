let string_format = f (
    'string_format',
    'function string_format (s: String, v: Representable) -> String',
        (s, v) => {
            return s.replace('${}', str(v))
        },
    'function string_format (s: String, t: Struct) -> String',
        (s, t) => {
            let used = 0
            let result = s.replace(/\$\{([^}]+)\}/g, (match, p1) => {
                used += 1
                let key = p1
                ensure(t.schema.has_key(key), 'format_invalid_key', key)
                let value = t.get(key)
                ensure(is(value, Types.Representable), 'not_repr', p1)
                return str(value)
            })
            ensure(used != 0, 'format_none_converted')
            return result
        },
    'function string_format (s: String, h: Hash) -> String',
        (s, h) => {
            let used = 0
            let result = s.replace(/\$\{([^}]+)\}/g, (match, p1) => {
                used += 1
                let key = p1
                let ok = has(key, h)
                ensure(ok, 'format_invalid_key', key)
                let value = h[key]
                ensure(is(value, Types.Representable), 'not_repr', p1)
                return str(value)
            })
            ensure(used != 0, 'format_none_converted')
            return result
        },
    'function string_format (s: String, l: List) -> String',
        (s, l) => {
            ensure(l.length != 0, 'format_empty_list')
            let used = 0
            let result = s.replace(/\$\{(\d+)\}/g, (match, p1) => {
                let index = parseInt(p1) - 1
                let ok = (0 <= index && index < l.length)
                ensure(ok, 'format_invalid_index', index)
                used += 1
                let value = l[index]
                ensure(is(value, Types.Representable), 'not_repr', p1)
                return str(value)
            })
            let ok = (used === l.length)
            ensure(ok, 'format_not_all_converted')
            return result
        }
)
