Types.Maybe = template (
    'function Maybe (T: Type) -> Type',
        T => Uni(Nil, T)
)

Types.Arity = template (
    'function Arity (n: Int) -> Type',
        n => Ins(Types.Function, $(
            f => f[WrapperInfo].proto.parameters.length === n
        ))
)
Types.Callable = Uni(ES.Function, Types.TypeTemplate, Types.Class, Types.Schema)

Types.Impl = template (
    'function Impl (i: Interface) -> Type',
        i => i.Impl
)

Types.Operand = template (
    'function Operand (op: String) -> Type',
        op => Uni (
            Ins(Types.Struct, $(s => s.schema.defined_operator(op))),
            Ins(Types.Instance, $(i => i.class_.defined_operator(op)))
        )
)

Types.Index = Ins(Types.Int, $(x => x >= 0))
Types.Size = Types.Index
Types.Char = Ins(Types.String, $(x => {
    let i = 0
    for (let _ of x) {
        if (i > 0) {
            return false
        }
        i += 1
    }
    return (i == 1)
}))
Types.Representable = Uni (
    Types.Primitive,
    Types.Operand.inflate('str')
)

Types.Error = $(x => x instanceof Error)

Types.NotFound = create_value('NotFound')  // Types.NotFound !== NotFound
