Types.Getter = template (
    'function Getter (Key: Type, Val: Type) -> Interface',
        (KT, VT) => create_interface('Getter', [
            { name: 'get', f: { parameters: [
                { name: 'key', type: KT },
                { name: 'nf', type: Types.Bool }
            ], value_type: Types.Maybe.inflate(VT) }}
        ], null)
)

Types.Setter = template (
    'function Setter (Key: Type, Val: Type) -> Interface',
        (KT, VT) => create_interface('Setter', [
            { name: 'set', f: { parameters: [
                { name: 'key', type: KT },
                { name: 'value', type: VT }
            ], value_type: Void } }
        ], null)
)

Types.SliceIndexDefault = create_value('SliceIndexDefault')

Types.SliceIndex = Uni(Types.Index, Types.SliceIndexDefault)

Types.SliceGetter = create_interface('SliceGetter', [
    { name: 'slice', f: { parameters: [
        { name: 'low', type: Types.SliceIndex },
        { name: 'high', type: Types.SliceIndex }
    ], value_type: Types.Any } }
], null)
