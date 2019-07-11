Types.Iterator = Ins (
    ES.Iterable, ES.Object,
    $(x => typeof x.next == 'function')
)

Types.AsyncIterator = Ins (
    ES.AsyncIterable,
    $(x => typeof x.next == 'function')
)

Types.EntryList = create_schema('EntryList', {
    keys: Types.List,
    values: Types.List
}, {}, [], { guard: fun (
    'function struct_guard (fields: Hash) -> Void',
    fields => {
        let ok = (fields.keys.length == fields.values.length)
        ensure(ok, 'bad_entry_list')
        return Void
    }
)})

Types.Iterable = Uni (
    ES.Iterable, Types.Enum,
    Types.Operand.inflate('iter')
)

Types.AsyncIterable = Uni (
    ES.AsyncIterable,
    Types.Operand.inflate('async_iter')
)

Types.Enumerable = Uni (
    Types.Hash, Types.Struct, Types.Enum,
    Types.Operand.inflate('enum')
)